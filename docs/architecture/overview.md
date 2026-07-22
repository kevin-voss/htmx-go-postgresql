# Architecture overview

Forgeboard is a **modular monolith**.

```text
HTTP request
    ↓
Middleware
    ↓
Handler
    ↓
Service
    ↓
Repository
    ↓
PostgreSQL
```

Do not over-engineer every layer with interfaces. Use interfaces only where they make testing or boundaries clearer.

---

## Layer responsibilities

### Handler

- parse URL and form input
- basic input validation
- call services
- select HTTP status
- render templates
- add HTMX response headers

### Service

- business rules
- authorization
- transactions
- coordinate repositories
- activity creation

### Repository

- SQL execution
- scanning rows
- database-specific errors
- **no** rendering
- **no** HTTP concepts

### Templates

- full pages
- reusable fragments
- form errors
- issue cards
- project lists
- activity rows

---

## Domain modules (internal packages)

Expected modules under `internal/`:

- `auth` — users, sessions, passwords, auth middleware
- `workspace` — workspaces, settings
- `member` — membership & invitations
- `project` — projects
- `issue` — issues, labels linkage, validation
- `comment` — issue comments
- `activity` — activity events
- `mail` — SMTP / Mailpit
- `database` — pool, migrate helpers
- `config` — env configuration
- `platform` — shared middleware, render, request/response helpers

## Related

- Structure: [project-structure.md](project-structure.md)
- Middleware: [middleware.md](middleware.md)
- Stack: [../specs/technology-stack.md](../specs/technology-stack.md)
