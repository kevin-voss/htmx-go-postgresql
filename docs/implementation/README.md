# Implementation plan

Ordered, atomic steps for building Forgeboard. **One AI agent executes one step**, verifies it, **commits and pushes**, then stops.

Start a step with [PROMPT.md](PROMPT.md) (change only the step number). Read [../AGENT_GUIDE.md](../AGENT_GUIDE.md) before starting.

---

## Rules

1. Execute steps in numeric order.
2. Do not skip dependencies listed in a step.
3. Meet every acceptance criterion before committing.
4. Use the step’s `type(scope): summary` + body (see the single example in [AGENT_GUIDE.md](../AGENT_GUIDE.md)).
5. **Commit and push** at the end of every step — mandatory, not optional.
6. Prefer linking existing specs/architecture/flows over inventing behavior.

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
| [01](steps/01-go-module-bootstrap.md) | Go module & package skeleton | `chore(repo): bootstrap Go module and package layout` |
| [02](steps/02-docker-compose.md) | Docker Compose services | `chore(docker): add Compose for app, Postgres, and Mailpit` |
| [03](steps/03-config-and-logging.md) | Config & slog | `feat(config): load env settings and structured slog logging` |
| [04](steps/04-http-server-health.md) | HTTP server, health, shutdown | `feat(server): add net/http server with health and shutdown` |
| [05](steps/05-templates-and-static.md) | Templates & static files | `feat(web): add template rendering and static file serving` |
| [06](steps/06-css-foundation.md) | CSS foundation | `feat(css): add layered tokens and light/dark foundation` |
| [07](steps/07-database-and-migrations.md) | DB pool & goose migrations | `feat(database): connect PostgreSQL and add goose migrations` |
| [08](steps/08-makefile-dev.md) | Makefile & `make dev` | `chore(dx): add Makefile and migrate-then-run entrypoint` |
| [09](steps/09-landing-page.md) | Landing page | `feat(landing): add public Forgeboard landing page` |
| [10](steps/10-password-service.md) | Argon2id password service | `feat(auth): add Argon2id password hashing service` |
| [11](steps/11-users-and-registration.md) | Users & registration | `feat(auth): add user registration with validation` |
| [12](steps/12-sessions-login-logout.md) | Sessions, login, logout | `feat(auth): add session-based login and logout` |
| [13](steps/13-auth-middleware-csrf.md) | Auth middleware & CSRF | `feat(auth): add auth middleware and CSRF protection` |
| [14](steps/14-mail-and-email-verification.md) | Mail & email verification | `feat(mail): add Mailpit mailer and email verification` |
| [15](steps/15-password-reset-and-rate-limit.md) | Password reset & rate limit | `feat(auth): add password reset and login rate limiting` |
| [16](steps/16-account-sessions.md) | Account session list | `feat(auth): add account session management page` |
| [17](steps/17-workspaces.md) | Workspaces | `feat(workspaces): add workspace creation and slug routes` |
| [18](steps/18-membership-and-roles.md) | Membership & RBAC | `feat(authz): add workspace membership and role checks` |
| [19](steps/19-onboarding.md) | Onboarding | `feat(onboarding): add first-time workspace and project setup` |
| [20](steps/20-invitations-and-members.md) | Invitations & members | `feat(workspaces): add invitations and member management` |
| [21](steps/21-projects.md) | Projects | `feat(projects): add projects within workspaces` |
| [22](steps/22-issues-core.md) | Issues core | `feat(issues): add issue create, list, and detail` |
| [23](steps/23-issue-workflow-fields.md) | Status, priority, assignee, archive | `feat(issues): add status, priority, assignee, and archive` |
| [24](steps/24-labels.md) | Labels | `feat(labels): add workspace labels and issue tagging` |
| [25](steps/25-search-and-filter.md) | Search & filter | `feat(issues): add search and filtering on issue lists` |
| [26](steps/26-htmx-partial-rendering.md) | HTMX vendor & partials | `feat(htmx): vendor HTMX 4 and partial render helpers` |
| [27](steps/27-htmx-validation-and-inline-ux.md) | Validation UX & inline updates | `feat(htmx): add 422 fragments and inline issue updates` |
| [28](steps/28-comments-multi-partial.md) | Comments multi-partial | `feat(comments): add multi-target hx-partial comment posts` |
| [29](steps/29-activity-feed.md) | Activity feed | `feat(activity): add activity events and project feed` |
| [30](steps/30-seed-demo-data.md) | Seed demo data | `feat(seed): add demo account and seed command` |
| [31](steps/31-integration-tests.md) | Integration tests | `test(integration): add authz and repository integration tests` |
| [32](steps/32-ci-prod-readme.md) | CI, prod image, README polish | `chore(release): add CI, production image, and portfolio README` |

---

## Status tracking (optional)

If using beads / issues, create one task per step with `blocks` dependencies matching the order above. Link the step file path in the issue body.

## Template

New steps must follow [_template.md](_template.md).
