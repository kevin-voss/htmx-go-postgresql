package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type workspaceContextKey int

const (
	workspaceIDContextKey workspaceContextKey = iota
	workspaceNameContextKey
	workspaceSlugContextKey
	workspaceRoleContextKey
)

// ContextWithWorkspaceAccess stores resolved workspace membership on ctx.
func ContextWithWorkspaceAccess(ctx context.Context, id, name, slug, role string) context.Context {
	ctx = context.WithValue(ctx, workspaceIDContextKey, id)
	ctx = context.WithValue(ctx, workspaceNameContextKey, name)
	ctx = context.WithValue(ctx, workspaceSlugContextKey, slug)
	ctx = context.WithValue(ctx, workspaceRoleContextKey, role)
	return ctx
}

// WorkspaceIDFromContext returns the workspace id set by RequireMembership.
func WorkspaceIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(workspaceIDContextKey).(string)
	return v, ok && v != ""
}

// WorkspaceNameFromContext returns the workspace name set by RequireMembership.
func WorkspaceNameFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(workspaceNameContextKey).(string)
	return v, ok && v != ""
}

// WorkspaceSlugFromContext returns the workspace slug set by RequireMembership.
func WorkspaceSlugFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(workspaceSlugContextKey).(string)
	return v, ok && v != ""
}

// WorkspaceRoleFromContext returns the caller's role set by RequireMembership.
func WorkspaceRoleFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(workspaceRoleContextKey).(string)
	return v, ok && v != ""
}

// WorkspaceAccessResolver resolves membership for a workspace slug + user.
// Missing workspace or membership should return notFound (fail closed → 404).
type WorkspaceAccessResolver func(ctx context.Context, slug, userID string) (id, name, resolvedSlug, role string, err error)

// UserIDFromContext extracts the authenticated user id from the request context.
type UserIDFromContext func(ctx context.Context) (string, bool)

// RequireMembership ensures the authenticated user is a member of /w/{workspaceSlug}.
// Cross-workspace and unknown slugs both return 404 (enumeration-safe fail closed).
func RequireMembership(
	userIDFrom UserIDFromContext,
	resolve WorkspaceAccessResolver,
	notFound error,
	logger *slog.Logger,
) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := userIDFrom(r.Context())
			if !ok || userID == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			slug := strings.TrimSpace(r.PathValue("workspaceSlug"))
			if slug == "" {
				http.NotFound(w, r)
				return
			}

			id, name, resolvedSlug, role, err := resolve(r.Context(), slug, userID)
			if err != nil {
				if notFound != nil && errors.Is(err, notFound) {
					http.NotFound(w, r)
					return
				}
				if logger != nil {
					logger.Error("workspace membership resolve failed", "err", err, "slug", slug)
				}
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ctx := ContextWithWorkspaceAccess(r.Context(), id, name, resolvedSlug, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole allows the request only when the membership role is one of allowed.
// Missing membership context or disallowed role → 403 Forbidden.
func RequireRole(allowed ...string) Middleware {
	set := make(map[string]struct{}, len(allowed))
	for _, role := range allowed {
		set[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := WorkspaceRoleFromContext(r.Context())
			if !ok {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			if _, ok := set[role]; !ok {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireMutation blocks roles that cannot change workspace data (e.g. Viewer).
// canMutate receives the role string from context; false → 403 Forbidden.
func RequireMutation(canMutate func(role string) bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := WorkspaceRoleFromContext(r.Context())
			if !ok || canMutate == nil || !canMutate(role) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
