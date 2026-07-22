# Development milestones

Grouped delivery themes. Implementation steps map 1:1 into these milestones.

---

## Milestone 1: Foundation (steps 01–09)

Deliver:

- Go project
- `net/http` server
- graceful shutdown
- templates
- embedded static files
- PostgreSQL connection
- migrations
- Docker Compose
- `make dev`
- health endpoint
- basic CSS foundation
- landing page

Routes:

```text
GET /health
GET /
```

---

## Milestone 2: Authentication (steps 10–16)

Deliver:

- registration
- login / logout
- sessions
- auth middleware
- password hashing
- login rate limiting
- CSRF
- email verification
- password reset
- protected dashboard
- account session list

Flows: [../examples/flows/registration.md](../examples/flows/registration.md), [login.md](../examples/flows/login.md).

---

## Milestone 3: Workspaces (steps 17–20)

Deliver:

- create workspace
- workspace slug
- membership
- authorization
- onboarding
- member list
- invite flow
- workspace settings (basics)

Flows: [onboarding.md](../examples/flows/onboarding.md), [invitation.md](../examples/flows/invitation.md).

---

## Milestone 4: Projects and issues (steps 21–25)

Deliver:

- project creation
- issue creation / list / detail
- status / priority / assignment / archive
- labels
- search and filtering

Flows: [issue-creation.md](../examples/flows/issue-creation.md), [issue-status.md](../examples/flows/issue-status.md).

---

## Milestone 5: HTMX experience (steps 26–28)

Deliver:

- vendored HTMX 4
- partial vs full rendering
- inline validation (`422` fragments)
- partial issue creation / status updates
- comment append with `<hx-partial>` multi-target
- loading indicators / disabled submit / history (as scoped in steps)

Flow: [comments.md](../examples/flows/comments.md).

---

## Milestone 6: Portfolio quality (steps 29–32)

Deliver:

- activity log
- seeded demo account
- integration tests
- CI workflow
- production Docker image
- polished README
- (architecture diagram / threat model notes in README or docs)

Global completion: [../DEFINITION_OF_DONE.md](../DEFINITION_OF_DONE.md).
