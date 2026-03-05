package types

import "testing"

func TestNormalizeSlug(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "basic", in: "openai", want: "openai"},
		{name: "with underscore", in: "gpt5_proxy", want: "gpt5_proxy"},
		{name: "with dash", in: "foo-1", want: "foo-1"},
		{name: "upper gets normalized", in: "OpenAI", want: "openai"},
		{name: "path traversal", in: "../evil", wantErr: true},
		{name: "absolute path", in: "/tmp/evil", wantErr: true},
		{name: "separator", in: "a/b", wantErr: true},
		{name: "dot prefix", in: ".hidden", wantErr: true},
		{name: "empty", in: "", wantErr: true},
		{name: "too long", in: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeSlug(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NormalizeSlug(%q) expected error", tt.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeSlug(%q) error = %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeSlug(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
