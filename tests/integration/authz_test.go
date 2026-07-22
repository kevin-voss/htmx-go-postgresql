package integration

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kevin-voss/htmx-go-postgresql/internal/app"
	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/config"
	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
)

func TestCrossWorkspaceAccessDenied(t *testing.T) {
	pool := Pool(t)
	owner := SeedWorkspace(t, pool)
	outsider := SeedWorkspace(t, pool)

	application := app.New(
		config.Config{Env: "test", Address: ":0", CookieSecure: false},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		pool,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	authRepo := auth.NewRepository(pool)
	svc := auth.NewService(authRepo, authRepo, authRepo, authRepo)
	_, rawToken, err := svc.CreateSession(ctx, auth.CreateSessionInput{UserID: outsider.UserID})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	path := "/w/" + owner.WorkspaceSlug + "/projects/" + owner.ProjectSlug + "/issues"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName(false), Value: rawToken})
	rr := httptest.NewRecorder()
	application.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%q", rr.Code, http.StatusNotFound, rr.Body.String())
	}
}

func TestMembershipResolveAccessBySlugFailClosed(t *testing.T) {
	pool := Pool(t)
	owner := SeedWorkspace(t, pool)
	outsider := SeedWorkspace(t, pool)

	svc := member.NewService(member.NewRepository(pool))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := svc.ResolveAccessBySlug(ctx, owner.WorkspaceSlug, outsider.UserID)
	if !errors.Is(err, member.ErrNotFound) {
		t.Fatalf("ResolveAccessBySlug: err = %v, want member.ErrNotFound", err)
	}

	access, err := svc.ResolveAccessBySlug(ctx, owner.WorkspaceSlug, owner.UserID)
	if err != nil {
		t.Fatalf("owner ResolveAccessBySlug: %v", err)
	}
	if access.WorkspaceID != owner.WorkspaceID {
		t.Fatalf("WorkspaceID = %q, want %q", access.WorkspaceID, owner.WorkspaceID)
	}
	if access.Membership.Role != member.RoleOwner {
		t.Fatalf("Role = %q, want %q", access.Membership.Role, member.RoleOwner)
	}
}
