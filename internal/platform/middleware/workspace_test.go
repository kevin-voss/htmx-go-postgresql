package middleware_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
)

func TestRequireMembershipSetsContext(t *testing.T) {
	t.Parallel()

	notFound := errors.New("missing")
	h := middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := middleware.WorkspaceIDFromContext(r.Context())
			if !ok || id != "w1" {
				t.Fatalf("workspace id = %q ok=%v", id, ok)
			}
			role, ok := middleware.WorkspaceRoleFromContext(r.Context())
			if !ok || role != "member" {
				t.Fatalf("role = %q ok=%v", role, ok)
			}
			w.WriteHeader(http.StatusOK)
		}),
		middleware.RequireMembership(
			func(context.Context) (string, bool) { return "u1", true },
			func(context.Context, string, string) (string, string, string, string, error) {
				return "w1", "Acme", "acme", "member", nil
			},
			notFound,
			slog.New(slog.NewTextHandler(io.Discard, nil)),
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/w/acme", nil)
	req.SetPathValue("workspaceSlug", "acme")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRequireMembershipNotFound(t *testing.T) {
	t.Parallel()

	notFound := errors.New("missing")
	h := middleware.Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("handler should not run")
		}),
		middleware.RequireMembership(
			func(context.Context) (string, bool) { return "u1", true },
			func(context.Context, string, string) (string, string, string, string, error) {
				return "", "", "", "", notFound
			},
			notFound,
			nil,
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/w/missing", nil)
	req.SetPathValue("workspaceSlug", "missing")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestRequireRoleAndMutation(t *testing.T) {
	t.Parallel()

	ownerOnly := middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		middleware.RequireRole("owner"),
	)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	ctx := middleware.ContextWithWorkspaceAccess(req.Context(), "w1", "Acme", "acme", "viewer")
	rr := httptest.NewRecorder()
	ownerOnly.ServeHTTP(rr, req.WithContext(ctx))
	if rr.Code != http.StatusForbidden {
		t.Fatalf("viewer settings status = %d, want 403", rr.Code)
	}

	mutate := middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		middleware.RequireMutation(func(role string) bool { return role == "owner" || role == "member" }),
	)
	rr = httptest.NewRecorder()
	mutate.ServeHTTP(rr, req.WithContext(ctx))
	if rr.Code != http.StatusForbidden {
		t.Fatalf("viewer mutate status = %d, want 403", rr.Code)
	}
}
