# Step 04: HTTP server, health, graceful shutdown

| Field | Value |
| ----- | ----- |
| ID | `STEP-04` |
| Milestone | M1 — Foundation |
| Status | `todo` |
| Depends on | STEP-03 |
| Unlocks | STEP-05 |
| Estimated scope | M |

---

## Goal

A net/http server listens on APP_ADDRESS, exposes GET /health, and shuts down gracefully on SIGINT/SIGTERM.

## Description

Introduce internal/app server wiring with http.NewServeMux only. No third-party router. Health should be suitable for Compose healthchecks later.

## References

- Stack / ServeMux: [technology-stack.md](../../specs/technology-stack.md)
- Overview: [overview.md](../../architecture/overview.md)
- Middleware (later): [middleware.md](../../architecture/middleware.md)

## Prerequisites

- Config loads.

## Scope

### In

- http.Server with timeouts
- GET /health → 200 + simple body
- Graceful shutdown
- routes.go / server.go split as in architecture
- Request ID or recover middleware optional stub OK if tiny

### Out

- HTML templates
- Auth routes
- Full middleware chain

## Implementation checklist

- [ ] Create Application + routes + ListenAndServe wrapper
- [ ] Register GET /health
- [ ] Implement graceful shutdown
- [ ] Manual or httptest test for /health

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/app/app.go` | create | Application deps |
| `internal/app/routes.go` | create | ServeMux registration |
| `internal/app/server.go` | create | run + shutdown |
| `cmd/web/main.go` | modify | start server |

## Technical notes

Set ReadHeaderTimeout / IdleTimeout. Do not use Gin/Chi.

## Acceptance criteria

- [ ] GET /health returns 200
- [ ] Server uses http.ServeMux with method+path patterns
- [ ] SIGINT triggers graceful shutdown without panic
- [ ] No third-party HTTP router dependency

## Verification

```bash
go test ./internal/app/...
# or: go run ./cmd/web & curl -i localhost:8080/health
```

## Commit

**Subject (required):**

```text
feat(step-04): add net/http server with health and graceful shutdown
```

**Body (optional):**

```text
Complete STEP-04 so the next agent can continue from a green tree.
```

## Handoff to next agent

/health works. Ready for templates/static.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-05.
