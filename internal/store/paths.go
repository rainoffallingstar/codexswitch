package store

// SwitchDir returns the path to codexswitch's provider store (usually ~/.codexswitch).
func SwitchDir() (string, error) {
	return switchDir()
}

// CodexDir returns the path to Codex's config directory (usually ~/.codex).
func CodexDir() (string, error) {
	return codexDir()
}
