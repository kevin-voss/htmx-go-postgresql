package member_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
)

type resolveStore struct {
	access map[string]member.Access
}

func (s *resolveStore) Create(context.Context, string, string, member.Role) (member.Membership, error) {
	return member.Membership{}, nil
}

func (s *resolveStore) GetByWorkspaceAndUser(context.Context, string, string) (member.Membership, error) {
	return member.Membership{}, member.ErrNotFound
}

func (s *resolveStore) GetAccessBySlug(_ context.Context, slug, userID string) (member.Access, error) {
	a, ok := s.access[slug+"|"+userID]
	if !ok {
		return member.Access{}, member.ErrNotFound
	}
	return a, nil
}

func (s *resolveStore) HasAny(context.Context, string) (bool, error) {
	return len(s.access) > 0, nil
}

func TestRequireMembershipOutsiderForbiddenAsNotFound(t *testing.T) {
	t.Parallel()

	svc := member.NewService(&resolveStore{
		access: map[string]member.Access{
			"acme|owner1": {
				WorkspaceID:   "w1",
				WorkspaceName: "Acme",
				WorkspaceSlug: "acme",
				Membership:    member.Membership{Role: member.RoleOwner, UserID: "owner1", WorkspaceID: "w1"},
			},
		},
	})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := middleware.Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("handler should not run for outsider")
		}),
		member.RequireMembership(svc, logger),
	)
	h = auth.RequireAuthentication(h)

	req := httptest.NewRequest(http.MethodGet, "/w/acme", nil)
	req.SetPathValue("workspaceSlug", "acme")
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{ID: "outsider", Email: "o@example.com"}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestRequireMembershipAllowsMember(t *testing.T) {
	t.Parallel()

	svc := member.NewService(&resolveStore{
		access: map[string]member.Access{
			"acme|u1": {
				WorkspaceID:   "w1",
				WorkspaceName: "Acme",
				WorkspaceSlug: "acme",
				Membership:    member.Membership{Role: member.RoleViewer, UserID: "u1", WorkspaceID: "w1"},
			},
		},
	})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	var sawRole string
	h := middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := middleware.WorkspaceRoleFromContext(r.Context())
			if !ok {
				t.Fatal("missing role in context")
			}
			sawRole = role
			w.WriteHeader(http.StatusOK)
		}),
		member.RequireMembership(svc, logger),
	)

	req := httptest.NewRequest(http.MethodGet, "/w/acme", nil)
	req.SetPathValue("workspaceSlug", "acme")
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{ID: "u1"}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if sawRole != string(member.RoleViewer) {
		t.Fatalf("role = %q, want viewer", sawRole)
	}
}

func TestViewerCannotMutate(t *testing.T) {
	t.Parallel()

	h := middleware.Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("viewer must not reach mutation handler")
		}),
		member.RequireCanMutate(),
	)

	req := httptest.NewRequest(http.MethodPost, "/w/acme/projects", nil)
	ctx := middleware.ContextWithWorkspaceAccess(req.Context(), "w1", "Acme", "acme", string(member.RoleViewer))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestMemberCanMutate(t *testing.T) {
	t.Parallel()

	called := false
	h := middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusNoContent)
		}),
		member.RequireCanMutate(),
	)

	req := httptest.NewRequest(http.MethodPost, "/w/acme/projects", nil)
	ctx := middleware.ContextWithWorkspaceAccess(req.Context(), "w1", "Acme", "acme", string(member.RoleMember))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if !called {
		t.Fatal("expected mutation handler to run")
	}
}

func TestRequireOwnerAllowsOwnerSettings(t *testing.T) {
	t.Parallel()

	h := middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		member.RequireOwner(),
	)

	req := httptest.NewRequest(http.MethodGet, "/w/acme/settings", nil)
	ctx := middleware.ContextWithWorkspaceAccess(req.Context(), "w1", "Acme", "acme", string(member.RoleOwner))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRequireOwnerRejectsMember(t *testing.T) {
	t.Parallel()

	h := middleware.Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("member must not access owner settings")
		}),
		member.RequireOwner(),
	)

	req := httptest.NewRequest(http.MethodGet, "/w/acme/settings", nil)
	ctx := middleware.ContextWithWorkspaceAccess(req.Context(), "w1", "Acme", "acme", string(member.RoleMember))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}
