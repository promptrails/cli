package cmd

import "testing"

func TestMaskSecret(t *testing.T) {
	cases := map[string]string{
		"pr_live_abcdefghijklmnop": "...mnop",
		"abcde":                    "...bcde",
		"abcd":                     "****",
		"ab":                       "**",
		"":                         "",
	}
	for in, want := range cases {
		if got := maskSecret(in); got != want {
			t.Errorf("maskSecret(%q) = %q, want %q", in, got, want)
		}
	}
}
