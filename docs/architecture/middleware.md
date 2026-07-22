# Middleware design

## Order

```text
Request ID
    ↓
Structured logging
    ↓
Panic recovery
    ↓
Security headers
    ↓
Session loading
    ↓
CSRF protection
    ↓
Authentication requirement (where applicable)
    ↓
Workspace membership authorization
    ↓
Handler
```

## Signature

```go
type Middleware func(http.Handler) http.Handler
```

## Composition

```go
func chain(
    handler http.Handler,
    middleware ...Middleware,
) http.Handler {
    for i := len(middleware) - 1; i >= 0; i-- {
        handler = middleware[i](handler)
    }
    return handler
}
```

## Chains

```go
publicHandler := chain(
    mux,
    recoverPanic,
    securityHeaders,
    requestLogger,
    loadSession,
)

authenticatedHandler := chain(
    appMux,
    requireAuthentication,
)
```

Workspace-scoped routes add membership/authorization middleware after authentication.

## Related

- Auth spec: [../specs/authentication.md](../specs/authentication.md)
- Roles: [../specs/roles.md](../specs/roles.md)
- Overview: [overview.md](overview.md)
