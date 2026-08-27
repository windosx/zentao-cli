package cmd

import "testing"

func TestMaskCookie(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"zentaosid=abcdef123456", "zentaosid=****3456"},
		{"zentaosid=abcd", "zentaosid=****abcd"},
		{"novalue", "****alue"},
	}
	for _, tt := range tests {
		if got := maskCookie(tt.in); got != tt.want {
			t.Errorf("maskCookie(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
