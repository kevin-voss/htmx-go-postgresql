# Proposed project structure

```text
forgeboard/
├── cmd/
│   └── web/
│       └── main.go
│
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   ├── routes.go
│   │   └── server.go
│   │
│   ├── auth/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── middleware.go
│   │   ├── password.go
│   │   └── session.go
│   │
│   ├── workspace/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   │
│   ├── project/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   │
│   ├── issue/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── validation.go
│   │   └── model.go
│   │
│   ├── comment/
│   ├── activity/
│   ├── member/
│   ├── mail/
│   ├── database/
│   ├── config/
│   └── platform/
│       ├── middleware/
│       ├── render/
│       ├── request/
│       └── response/
│
├── web/
│   ├── templates/
│   │   ├── layouts/
│   │   ├── pages/
│   │   ├── components/
│   │   └── fragments/
│   └── static/
│       ├── css/
│       │   ├── reset.css
│       │   ├── tokens.css
│       │   ├── base.css
│       │   ├── layout.css
│       │   ├── components.css
│       │   ├── utilities.css
│       │   └── pages/
│       │       ├── auth.css
│       │       ├── project.css
│       │       └── issue.css
│       ├── js/
│       │   ├── htmx-4.0.0-beta5.min.js   # prefer vendor/ path if used
│       │   └── app.js
│       ├── vendor/                       # recommended for HTMX pin
│       │   └── htmx-4.0.0-beta5.min.js
│       └── images/
│
├── db/
│   ├── migrations/
│   ├── queries/
│   └── sqlc.yaml
│
├── tests/
│   ├── integration/
│   └── fixtures/
│
├── docs/                    ← this documentation tree
├── Dockerfile
├── compose.yaml
├── Makefile
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

## Notes

- Also expect `cmd/migrate` and `cmd/seed` as Makefile targets mature.
- `sqlc` may be omitted until the first SQL flows work; handwritten SQL via `pgx` is fine early.
- Prefer vendoring HTMX under `web/static/vendor/` as described in [../specs/htmx-decision.md](../specs/htmx-decision.md).

## Related

- Overview: [overview.md](overview.md)
- Docker: [docker.md](docker.md)
