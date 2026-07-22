package request_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/request"
)

func TestIsPartialRequest(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		header string
		want   bool
	}{
		{name: "partial", header: "partial", want: true},
		{name: "full", header: "full", want: false},
		{name: "absent", header: "", want: false},
		{name: "other", header: "boosted", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/w/acme/projects/platform", nil)
			if tc.header != "" {
				req.Header.Set("HX-Request-Type", tc.header)
			}

			if got := request.IsPartialRequest(req); got != tc.want {
				t.Fatalf("IsPartialRequest = %v, want %v", got, tc.want)
			}
		})
	}
}
