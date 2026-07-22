# Technology stack

```text
Language:          Go
HTTP:              net/http
Router:            http.ServeMux
Templates:         html/template
Dynamic UI:        HTMX 4.0.0-beta5 (pinned, vendored)
CSS:               handwritten CSS (no Tailwind / Sass / PostCSS)
Database:          PostgreSQL
Driver:            pgx
SQL:               handwritten SQL
SQL generation:    sqlc (optional later; omit until first SQL flows work)
Migrations:        goose
Authentication:    custom session authentication
Password hashing:  Argon2id
Email development: Mailpit
Containers:        Docker Compose
Testing:           testing + httptest
Logging:           log/slog
Build tooling:     Make
```

## Router decision: use `net/http`

Use `http.NewServeMux()` — **not** Chi, Gin, Echo, or Fiber.

Since Go 1.22, `ServeMux` supports method matching, path wildcards, path parameters, automatic `405`, and route specificity.

```go
mux := http.NewServeMux()

mux.HandleFunc("GET /", app.home)
mux.HandleFunc("GET /login", app.showLogin)
mux.HandleFunc("POST /login", app.login)
mux.HandleFunc("POST /logout", app.logout)

mux.HandleFunc(
    "GET /w/{workspaceSlug}/projects/{projectSlug}",
    app.showProject,
)
```

Path values:

```go
workspaceSlug := r.PathValue("workspaceSlug")
```

### Why this matters for the portfolio

Using `net/http` demonstrates handlers, middleware, contexts, cookies, headers, methods, status codes, server configuration, graceful shutdown, and HTML responses — without a framework hiding fundamentals.

## Related

- HTMX pin: [htmx-decision.md](htmx-decision.md)
- Architecture: [../architecture/overview.md](../architecture/overview.md)
- Docker: [../architecture/docker.md](../architecture/docker.md)
