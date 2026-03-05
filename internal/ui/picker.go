package ui

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/rainoffallingstar/codexswitch/internal/types"
)

// PickProvider shows an interactive menu and returns the selected provider.
func PickProvider(list types.ProviderList) (types.Provider, error) {
	return PickProviderWithLabel(list, "Select provider")
}

// PickProviderWithLabel shows an interactive menu with a custom label and returns the selected provider.
func PickProviderWithLabel(list types.ProviderList, label string) (types.Provider, error) {
	if len(list.Providers) == 0 {
		return types.Provider{}, nil
	}

	labels := make([]string, len(list.Providers))
	startIdx := 0
	for i, p := range list.Providers {
		labels[i] = p.DisplayName
		if p.Slug == list.CurrentSlug {
			labels[i] += "  [active]"
			startIdx = i
		}
	}

	prompt := promptui.Select{
		Label:     label,
		Items:     labels,
		CursorPos: startIdx,
		Size:      10,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return fallbackMenu(label+":", list)
	}
	return list.Providers[idx], nil
}

// PromptAddProvider collects provider details from flags and optional interactive prompts.
func PromptAddProvider(seed types.AddProviderInput, interactive bool) (types.AddProviderInput, error) {
	input := seed

	fields := []struct {
		label    string
		value    *string
		required bool
		dflt     string
		mask     rune
	}{
		{label: "Provider slug", value: &input.Slug, required: true},
		{label: "Display name", value: &input.DisplayName, required: true},
		{label: "API Key", value: &input.APIKey, required: true, mask: '*'},
		{label: "Model", value: &input.Model, required: true},
		{label: "Base URL", value: &input.BaseURL, required: true},
		{label: "Wire API", value: &input.WireAPI, required: false, dflt: "responses"},
		{label: "Reasoning effort", value: &input.ReasoningEffort, required: false, dflt: types.DefaultReasoningEffort},
	}

	for _, f := range fields {
		*f.value = strings.TrimSpace(*f.value)
		if *f.value != "" {
			continue
		}

		if !interactive {
			if f.required {
				return types.AddProviderInput{}, fmt.Errorf("missing required option: %s (set via flags in non-interactive mode)", f.label)
			}
			*f.value = f.dflt
			continue
		}

		prompt := promptui.Prompt{
			Label:   f.label,
			Default: f.dflt,
		}
		if f.mask != 0 {
			prompt.Mask = f.mask
		}

		val, err := prompt.Run()
		if err != nil {
			return types.AddProviderInput{}, fmt.Errorf("read %s: %w", strings.ToLower(f.label), err)
		}
		val = strings.TrimSpace(val)
		if val == "" {
			val = f.dflt
		}
		if f.required && val == "" {
			return types.AddProviderInput{}, fmt.Errorf("%s is required", strings.ToLower(f.label))
		}
		*f.value = val
	}

	if input.WireAPI == "" {
		input.WireAPI = "responses"
	}
	slug, err := types.NormalizeSlug(input.Slug)
	if err != nil {
		return types.AddProviderInput{}, err
	}
	input.Slug = slug

	reasoningEffort, err := types.NormalizeReasoningEffort(input.ReasoningEffort)
	if err != nil {
		return types.AddProviderInput{}, err
	}
	input.ReasoningEffort = reasoningEffort

	return input, nil
}

// PromptEditProvider updates a provider by combining flags with optional prompts.
func PromptEditProvider(seed types.Provider, opts types.AddProviderInput, interactive bool) (types.Provider, error) {
	out := seed

	trimmedName := strings.TrimSpace(opts.DisplayName)
	if trimmedName != "" {
		out.DisplayName = trimmedName
	} else if interactive {
		val, err := runPrompt("Display name", out.DisplayName, 0)
		if err != nil {
			return types.Provider{}, err
		}
		if val != "" {
			out.DisplayName = val
		}
	}

	trimmedAPIKey := strings.TrimSpace(opts.APIKey)
	if trimmedAPIKey != "" {
		out.APIKey = trimmedAPIKey
	} else if interactive {
		val, err := runPrompt("API Key (leave empty to keep current)", "", '*')
		if err != nil {
			return types.Provider{}, err
		}
		if val != "" {
			out.APIKey = val
		}
	}

	trimmedModel := strings.TrimSpace(opts.Model)
	if trimmedModel != "" {
		out.Model = trimmedModel
	} else if interactive {
		val, err := runPrompt("Model", out.Model, 0)
		if err != nil {
			return types.Provider{}, err
		}
		if val != "" {
			out.Model = val
		}
	}

	trimmedBaseURL := strings.TrimSpace(opts.BaseURL)
	if trimmedBaseURL != "" {
		out.BaseURL = trimmedBaseURL
	} else if interactive {
		val, err := runPrompt("Base URL", out.BaseURL, 0)
		if err != nil {
			return types.Provider{}, err
		}
		if val != "" {
			out.BaseURL = val
		}
	}

	trimmedWireAPI := strings.TrimSpace(opts.WireAPI)
	if trimmedWireAPI != "" {
		out.WireAPI = trimmedWireAPI
	} else if interactive {
		wireDefault := out.WireAPI
		if wireDefault == "" {
			wireDefault = "responses"
		}
		val, err := runPrompt("Wire API", wireDefault, 0)
		if err != nil {
			return types.Provider{}, err
		}
		if val != "" {
			out.WireAPI = val
		}
	}

	trimmedReasoningEffort := strings.TrimSpace(opts.ReasoningEffort)
	if trimmedReasoningEffort != "" {
		out.ReasoningEffort = trimmedReasoningEffort
	} else if interactive {
		reasoningDefault := out.ReasoningEffort
		if reasoningDefault == "" {
			reasoningDefault = types.DefaultReasoningEffort
		}
		val, err := runPrompt("Reasoning effort", reasoningDefault, 0)
		if err != nil {
			return types.Provider{}, err
		}
		if val != "" {
			out.ReasoningEffort = val
		}
	}

	if out.WireAPI == "" {
		out.WireAPI = "responses"
	}
	reasoningEffort, err := types.NormalizeReasoningEffort(out.ReasoningEffort)
	if err != nil {
		return types.Provider{}, err
	}
	out.ReasoningEffort = reasoningEffort

	return out, nil
}

// ConfirmRemoval asks for deletion confirmation unless autoYes is true.
func ConfirmRemoval(slug, displayName string, interactive, autoYes bool) (bool, error) {
	if autoYes {
		return true, nil
	}
	if !interactive {
		return false, fmt.Errorf("confirmation required in non-interactive mode; rerun with --yes")
	}

	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Remove provider '%s' (%s)?", displayName, slug),
		IsConfirm: true,
	}
	val, err := prompt.Run()
	if err != nil {
		return false, nil
	}
	return strings.EqualFold(strings.TrimSpace(val), "y"), nil
}

func runPrompt(label, dflt string, mask rune) (string, error) {
	prompt := promptui.Prompt{
		Label:   label,
		Default: dflt,
	}
	if mask != 0 {
		prompt.Mask = mask
	}
	val, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("read %s: %w", strings.ToLower(label), err)
	}
	return strings.TrimSpace(val), nil
}

func fallbackMenu(title string, list types.ProviderList) (types.Provider, error) {
	fmt.Println(title)
	for i, p := range list.Providers {
		marker := ""
		if p.Slug == list.CurrentSlug {
			marker = " [active]"
		}
		fmt.Printf("  %d) %s%s\n", i+1, p.DisplayName, marker)
	}
	fmt.Print("Enter number: ")

	var n int
	if _, err := fmt.Scan(&n); err != nil {
		return types.Provider{}, fmt.Errorf("read selection: %w", err)
	}
	if n < 1 || n > len(list.Providers) {
		return types.Provider{}, fmt.Errorf("invalid selection: %d", n)
	}
	return list.Providers[n-1], nil
}
