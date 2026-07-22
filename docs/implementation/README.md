# Implementation plan

Ordered, atomic steps for building Forgeboard. **One AI agent executes one step**, verifies it, commits it, then stops.

Read [../AGENT_GUIDE.md](../AGENT_GUIDE.md) before starting.

---

## Rules

1. Execute steps in numeric order.
2. Do not skip dependencies listed in a step.
3. Meet every acceptance criterion before committing.
4. Use the step’s commit message.
5. Prefer linking existing specs/architecture/flows over inventing behavior.

---

## Milestone map

See [milestones.md](milestones.md).

| Milestone | Steps | Theme |
| --------- | ----- | ----- |
| M1 Foundation | 01–09 | Repo, Docker, server, DB, CSS, landing |
| M2 Authentication | 10–16 | Passwords, sessions, mail, reset, CSRF |
| M3 Workspaces | 17–20 | Workspaces, roles, onboarding, invites |
| M4 Projects & issues | 21–25 | Projects, issues, labels, filters |
| M5 HTMX experience | 26–28 | Partials, validation UX, comments multi-target |
| M6 Portfolio quality | 29–32 | Activity, seed, tests, CI, prod, README |

---

## Step index

| Step | Title | Commit (subject) |
| ---- | ----- | ---------------- |
| [01](steps/01-go-module-bootstrap.md) | Go module & package skeleton | `chore(step-01): bootstrap Go module and package layout` |
| [02](steps/02-docker-compose.md) | Docker Compose services | `chore(step-02): add Docker Compose for app, Postgres, Mailpit` |
| [03](steps/03-config-and-logging.md) | Config & slog | `feat(step-03): add env config and structured logging` |
| [04](steps/04-http-server-health.md) | HTTP server, health, shutdown | `feat(step-04): add net/http server with health and graceful shutdown` |
| [05](steps/05-templates-and-static.md) | Templates & static files | `feat(step-05): add html/template rendering and embedded static files` |
| [06](steps/06-css-foundation.md) | CSS foundation | `feat(step-06): add modern CSS layers and design tokens` |
| [07](steps/07-database-and-migrations.md) | DB pool & goose migrations | `feat(step-07): connect PostgreSQL and add goose migrations` |
| [08](steps/08-makefile-dev.md) | Makefile & `make dev` | `chore(step-08): add Makefile and development entrypoint` |
| [09](steps/09-landing-page.md) | Landing page | `feat(step-09): add public landing page` |
| [10](steps/10-password-service.md) | Argon2id password service | `feat(step-10): add Argon2id password hashing service` |
| [11](steps/11-users-and-registration.md) | Users & registration | `feat(step-11): add users table and registration flow` |
| [12](steps/12-sessions-login-logout.md) | Sessions, login, logout | `feat(step-12): add session auth with login and logout` |
| [13](steps/13-auth-middleware-csrf.md) | Auth middleware & CSRF | `feat(step-13): add auth middleware and CSRF protection` |
| [14](steps/14-mail-and-email-verification.md) | Mail & email verification | `feat(step-14): add Mailpit mailer and email verification` |
| [15](steps/15-password-reset-and-rate-limit.md) | Password reset & rate limit | `feat(step-15): add password reset and login rate limiting` |
| [16](steps/16-account-sessions.md) | Account session list | `feat(step-16): add account sessions management page` |
| [17](steps/17-workspaces.md) | Workspaces | `feat(step-17): add workspace creation and slug routing` |
| [18](steps/18-membership-and-roles.md) | Membership & RBAC | `feat(step-18): add workspace membership and role authorization` |
| [19](steps/19-onboarding.md) | Onboarding | `feat(step-19): add first-time workspace and project onboarding` |
| [20](steps/20-invitations-and-members.md) | Invitations & members | `feat(step-20): add workspace invitations and member management` |
| [21](steps/21-projects.md) | Projects | `feat(step-21): add projects within workspaces` |
| [22](steps/22-issues-core.md) | Issues core | `feat(step-22): add issue creation, list, and detail` |
| [23](steps/23-issue-workflow-fields.md) | Status, priority, assignee, archive | `feat(step-23): add issue status, priority, assignee, and archive` |
| [24](steps/24-labels.md) | Labels | `feat(step-24): add labels and issue labeling` |
| [25](steps/25-search-and-filter.md) | Search & filter | `feat(step-25): add issue search and filtering` |
| [26](steps/26-htmx-partial-rendering.md) | HTMX vendor & partials | `feat(step-26): vendor HTMX 4 and add partial rendering helpers` |
| [27](steps/27-htmx-validation-and-inline-ux.md) | Validation UX & inline updates | `feat(step-27): add 422 fragments and inline HTMX issue updates` |
| [28](steps/28-comments-multi-partial.md) | Comments multi-partial | `feat(step-28): add comments with multi-target hx-partial updates` |
| [29](steps/29-activity-feed.md) | Activity feed | `feat(step-29): add activity events and project activity feed` |
| [30](steps/30-seed-demo-data.md) | Seed demo data | `feat(step-30): add seed command and demo account` |
| [31](steps/31-integration-tests.md) | Integration tests | `test(step-31): add integration tests for authz and repositories` |
| [32](steps/32-ci-prod-readme.md) | CI, prod image, README polish | `chore(step-32): add CI, production image, and portfolio README` |

---

## Status tracking (optional)

If using beads / issues, create one task per step with `blocks` dependencies matching the order above. Link the step file path in the issue body.

## Template

New steps must follow [_template.md](_template.md).
