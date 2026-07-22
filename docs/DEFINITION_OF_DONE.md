# Definition of done — Forgeboard

The project is complete when a reviewer can execute this path without tribal knowledge.

---

## Reviewer walkthrough

```text
1.  Clone the repository
2.  Run make dev
3.  Open http://localhost:8080
4.  Register a user
5.  Create a workspace
6.  Create a project
7.  Create and update issues
8.  Add comments
9.  Invite another user
10. Observe authorization between workspaces
11. Run make test successfully
```

Related flows: [examples/flows/](examples/flows/).

---

## Quantitative demo requirements

The demo must contain at least:

| Requirement | Minimum |
| ----------- | ------- |
| Meaningful HTTP routes | 20 |
| Database tables | 10 |
| Reusable HTML fragments | 8 |
| Middleware components | 5 |
| Role / permission levels | 3 (Owner, Member, Viewer) |
| Handler tests | present |
| Authorization tests | present |
| Repository integration tests | present |
| Complete password-reset flow | 1 |
| Multi-target HTMX update | 1 |
| `422` validation HTML fragment | 1 |
| Transaction that also writes an activity event | 1 |

---

## Portfolio deliverables

- [ ] Seeded demo account
- [ ] Activity log visible in UI
- [ ] Architecture diagram (in README or docs)
- [ ] Threat model notes
- [ ] CI workflow
- [ ] Production Docker image target
- [ ] Polished README with clone → `make dev` instructions

Mailpit UI: `http://localhost:8025`  
App: `http://localhost:8080`  
PostgreSQL: `localhost:5432`

---

## Explicit non-goals (must still be true at the end)

No payments, OAuth, file uploads, WebSockets, Kubernetes, microservices, rich-text editor, time tracking, roadmaps, sprint planning, or external integrations.

See [specs/product.md](specs/product.md).
