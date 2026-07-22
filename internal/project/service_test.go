package project

import "testing"

func TestSlugFromName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "simple", in: "Launch Board", want: "launch-board"},
		{name: "trim", in: "  My App  ", want: "my-app"},
		{name: "punctuation", in: "Hello, World!", want: "hello-world"},
		{name: "empty", in: "   ", want: "project"},
		{name: "symbols only", in: "!!!", want: "project"},
		{name: "collapse hyphens", in: "a -- b", want: "a-b"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := SlugFromName(tc.in); got != tc.want {
				t.Fatalf("SlugFromName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
