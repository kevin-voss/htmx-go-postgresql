package member

import (
	"context"
	"log/slog"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
)

// RequireMembership returns middleware that resolves /w/{workspaceSlug} membership.
// Non-members and unknown workspaces receive 404 (fail closed).
func RequireMembership(svc *Service, logger *slog.Logger) middleware.Middleware {
	return middleware.RequireMembership(
		func(ctx context.Context) (string, bool) {
			u, ok := auth.UserFromContext(ctx)
			if !ok {
				return "", false
			}
			return u.ID, true
		},
		func(ctx context.Context, slug, userID string) (id, name, resolvedSlug, role string, err error) {
			access, err := svc.ResolveAccessBySlug(ctx, slug, userID)
			if err != nil {
				return "", "", "", "", err
			}
			return access.WorkspaceID, access.WorkspaceName, access.WorkspaceSlug, string(access.Membership.Role), nil
		},
		ErrNotFound,
		logger,
	)
}

// RequireOwner allows only workspace Owners (after RequireMembership).
func RequireOwner() middleware.Middleware {
	return middleware.RequireRole(string(RoleOwner))
}

// RequireCanMutate allows Owner/Member and rejects Viewer (after RequireMembership).
func RequireCanMutate() middleware.Middleware {
	return middleware.RequireMutation(func(role string) bool {
		return Role(role).CanMutate()
	})
}
