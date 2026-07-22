> **Structured docs live in [`docs/`](docs/README.md).**  
> This file is the original monolithic draft. Prefer `docs/specs/`, `docs/architecture/`, `docs/examples/flows/`, and especially `docs/implementation/` for agent-driven build steps.

---

## HTMX 2 vs HTMX 4

For this project, I agree: **use HTMX 4**.

But with one important condition: as of **July 22, 2026**, HTMX 4 is still officially described as **“under construction”**, and the current documented version is `4.0.0-beta5`. It is therefore suitable for a learning project, but I would not currently choose it for a conservative production system. Pin the exact version so a future beta does not unexpectedly break the project. ([htmx][1])

Use:

```html
<script src="/static/vendor/htmx-4.0.0-beta5.min.js"></script>
```

I would download the file and commit it into the repository instead of using npm or depending on a CDN. HTMX itself requires no build step. ([htmx][2])

### Important differences

| Area                   | HTMX 2                                | HTMX 4                                 |
| ---------------------- | ------------------------------------- | -------------------------------------- |
| Request implementation | `XMLHttpRequest`                      | Native `fetch()`                       |
| Attribute inheritance  | Implicit by default                   | Explicit using `:inherited`            |
| Error responses        | `4xx` and `5xx` generally not swapped | Error HTML is swapped by default       |
| History cache          | May use `localStorage`                | Re-fetches pages instead               |
| Request timeout        | No timeout by default                 | 60 seconds by default                  |
| Event naming           | `htmx:afterSwap`                      | `htmx:after:swap`                      |
| Extensions             | Enabled with `hx-ext`                 | Script inclusion activates them        |
| Multiple updates       | Mainly out-of-band swaps              | New `<hx-partial>` support             |
| Swap options           | Traditional replacements              | Adds morphing and `textContent`        |
| Status handling        | Usually custom event logic            | `hx-status:422`, `hx-status:5xx`, etc. |
| JavaScript helpers     | More HTMX utility methods             | Prefers native DOM APIs                |

These are the most relevant changes for our project. HTMX 4 also changes `hx-delete` so it no longer automatically includes enclosing form data, moves request queuing to `hx-sync`, renames several attributes and events, and swaps the main response before out-of-band elements. ([htmx][2])

### Why HTMX 4 works well here

We are starting from scratch, so we do not have migration concerns. We can learn the newer concepts directly:

* explicit inheritance
* error fragments using HTTP status codes
* `fetch()`-based requests
* `<hx-partial>` for multiple page updates
* `hx-status` for validation errors
* morph swaps
* cleaner event names
* native JavaScript instead of HTMX utility wrappers

For example:

```html
<form
    hx-post="/projects"
    hx-target="#project-list"
    hx-swap="beforeend"
    hx-status:422="target:#project-form-errors swap:innerHTML"
>
    ...
</form>
```

When validation fails, Go returns:

```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: text/html
```

And the response body contains an HTML error fragment. HTMX 4 can swap that fragment directly without custom JavaScript. ([htmx][2])

---

# Project decision

## Product name: Forgeboard

A lightweight, server-rendered issue and project tracker for small development teams.

Think of it as:

> A compact combination of Linear, GitHub Issues and a simple team activity feed.

The project demonstrates:

* Go fundamentals
* `net/http`
* server-side rendering
* HTMX 4
* PostgreSQL
* authentication
* authorization
* relational database design
* modern CSS
* Docker
* testing
* secure web development

It is not intended to compete with Jira. It is a focused portfolio application.

---

# 1. Product specification

## 1.1 Problem

Small software teams often need a lightweight place to:

* create projects
* track issues
* assign work
* discuss tasks
* monitor progress

Existing products can be too complex for small projects. Forgeboard offers a deliberately simple workflow.

## 1.2 Target users

Primary users:

* small software teams
* startup teams
* freelance developers
* technical project owners
* individual developers managing side projects

## 1.3 Core product promise

A user should be able to:

1. create an account
2. create a workspace
3. create a project
4. add issues
5. assign and prioritize issues
6. move issues through a workflow
7. collaborate through comments
8. see recent project activity

## 1.4 Scope

### Included

* account registration
* login and logout
* email verification
* password reset
* secure sessions
* workspaces
* workspace membership
* workspace invitations
* projects
* issues
* issue statuses
* issue priorities
* assignments
* comments
* labels
* search and filtering
* activity history
* user preferences
* responsive interface
* light and dark color scheme
* Docker-based local environment

### Excluded

* payments
* subscriptions
* file uploads
* OAuth
* social login
* mobile apps
* complex notifications
* WebSockets
* Kubernetes
* microservices
* rich-text editor
* time tracking
* roadmaps
* sprint planning
* external integrations

---

# 2. User roles

## Workspace owner

Can:

* update workspace settings
* invite and remove members
* change member roles
* create and archive projects
* transfer workspace ownership
* delete the workspace

## Workspace admin

Can:

* invite members
* manage projects
* manage labels
* edit all issues
* remove regular members

Cannot:

* delete the workspace
* transfer ownership
* remove the owner

## Member

Can:

* view workspace projects
* create issues
* edit issues
* comment
* assign issues
* change issue status

## Viewer

Can:

* view projects
* view issues
* view comments

Cannot change data.

For the first version, we could simplify this to:

```text
Owner
Member
Viewer
```

That is enough to demonstrate role-based authorization without creating excessive complexity.

---

# 3. Main product flows

## 3.1 Registration flow

```text
Landing page
    ↓
Registration form
    ↓
Validate email, display name and password
    ↓
Create user
    ↓
Send verification email through Mailpit
    ↓
Create session
    ↓
Redirect to onboarding
```

Registration fields:

* display name
* email
* password
* password confirmation
* acceptance of terms checkbox

Validation:

* valid email address
* normalized lowercase email
* unique email
* display name between 2 and 50 characters
* password of at least 12 characters
* matching password confirmation

## 3.2 Login flow

```text
Login form
    ↓
Look up user by normalized email
    ↓
Verify password hash
    ↓
Generate random session token
    ↓
Store token hash in PostgreSQL
    ↓
Set secure cookie
    ↓
Redirect to workspace
```

Login error messages should not reveal whether an account exists:

```text
Invalid email or password.
```

## 3.3 First-time onboarding

```text
Login
    ↓
No workspace membership found
    ↓
Create workspace
    ↓
Create first project
    ↓
Open project board
```

Onboarding should require only:

* workspace name
* workspace slug
* first project name

## 3.4 Invitation flow

```text
Owner enters email
    ↓
Create invitation token
    ↓
Send invitation email
    ↓
User opens invitation link
    ↓
Existing user logs in
or
New user registers
    ↓
Accept invitation
    ↓
Membership is created
```

## 3.5 Issue creation flow

```text
Open project board
    ↓
Click “New issue”
    ↓
Inline form or dialog opens
    ↓
Enter title, description and priority
    ↓
Submit through HTMX
    ↓
Server validates data
    ↓
Issue card appears without full reload
    ↓
Issue count and activity feed update
```

## 3.6 Issue status flow

Statuses:

```text
Backlog
Todo
In Progress
Done
```

Users can change status from:

* issue detail page
* project issue list
* board view

For the first version, use buttons or a `<select>` instead of drag and drop. Drag and drop would require additional JavaScript and is not important for demonstrating HTMX.

## 3.7 Comment flow

```text
Open issue
    ↓
Write comment
    ↓
POST through HTMX
    ↓
Append rendered comment
    ↓
Clear textarea
    ↓
Update comment count
```

This is a good place to use HTMX 4 `<hx-partial>`:

```html
<hx-partial hx-target="#comment-list" hx-swap="beforeend">
    <!-- rendered new comment -->
</hx-partial>

<hx-partial hx-target="#comment-count">
    <span>4 comments</span>
</hx-partial>

<hx-partial hx-target="#comment-form">
    <!-- cleared comment form -->
</hx-partial>
```

---

# 4. Main pages

## Public pages

```text
GET /
GET /login
GET /register
GET /verify-email
GET /forgot-password
GET /reset-password/{token}
GET /invites/{token}
```

## Authenticated pages

```text
GET /app
GET /app/workspaces/new
GET /w/{workspaceSlug}
GET /w/{workspaceSlug}/projects
GET /w/{workspaceSlug}/projects/{projectSlug}
GET /w/{workspaceSlug}/projects/{projectSlug}/issues
GET /w/{workspaceSlug}/issues/{issueNumber}
GET /w/{workspaceSlug}/members
GET /w/{workspaceSlug}/settings
GET /account/settings
GET /account/sessions
```

---

# 5. Technical specification

## 5.1 Technology stack

```text
Language:          Go
HTTP:              net/http
Router:            http.ServeMux
Templates:         html/template
Dynamic UI:        HTMX 4.0.0-beta5
CSS:               handwritten CSS
Database:          PostgreSQL
Driver:            pgx
SQL:               handwritten SQL
SQL generation:    sqlc
Migrations:        goose
Authentication:    custom session authentication
Password hashing:  Argon2id
Email development: Mailpit
Containers:        Docker Compose
Testing:           testing + httptest
Logging:           log/slog
Build tooling:     Make
```

---

# 6. Should we use `net/http`?

## Yes

For this project, use:

```go
http.NewServeMux()
```

and not Chi, Gin, Echo or Fiber.

Since Go 1.22, the standard `ServeMux` supports:

* HTTP method matching
* path wildcards
* path parameters
* automatic `405 Method Not Allowed`
* route specificity handling

For example:

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

mux.HandleFunc(
    "POST /w/{workspaceSlug}/projects/{projectSlug}/issues",
    app.createIssue,
)
```

Path values are retrieved through:

```go
workspaceSlug := r.PathValue("workspaceSlug")
projectSlug := r.PathValue("projectSlug")
```

The Go team added these features specifically after studying frequently used functionality in third-party routers. For a project of this size, `net/http` is fully sufficient and removes one more dependency. ([Go][3])

I would not claim that `net/http` is necessarily “more popular than all Go routers” without survey data. However, it is Go’s standard HTTP server package, forms the foundation for many Go routers and frameworks, and is widely understood by Go developers. The standard library includes the server, handlers, routing, cookies, HTTP/2 support, timeouts and graceful shutdown primitives we need. ([Go Packages][4])

### Why it is valuable for your resume

Using `net/http` demonstrates that you understand:

* handlers
* middleware
* request contexts
* cookies
* headers
* HTTP methods
* status codes
* server configuration
* graceful shutdown
* routing
* HTML responses

It prevents the framework from hiding the fundamentals.

---

# 7. Application architecture

Use a modular monolith.

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

## Responsibilities

### Handler

Responsible for:

* parsing URL and form input
* basic input validation
* calling services
* selecting HTTP status
* rendering templates
* adding HTMX response headers

### Service

Responsible for:

* business rules
* authorization
* transactions
* coordinating repositories
* activity creation

### Repository

Responsible for:

* SQL execution
* scanning database rows
* database-specific errors
* no rendering
* no HTTP concepts

### Templates

Responsible for:

* full pages
* reusable fragments
* form errors
* issue cards
* project lists
* activity rows

---

# 8. Proposed project structure

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
│       │   ├── htmx-4.0.0-beta5.min.js
│       │   └── app.js
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
├── Dockerfile
├── compose.yaml
├── Makefile
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

For a smaller starting point, we can omit `sqlc` initially and add it once the first SQL flows work.

---

# 9. Middleware design

Middleware order:

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
Authentication requirement where applicable
    ↓
Workspace membership authorization
    ↓
Handler
```

Possible middleware signature:

```go
type Middleware func(http.Handler) http.Handler
```

Composition:

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

Use different chains:

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

---

# 10. Rendering strategy

Every important view has two representations:

1. full page
2. fragment

A normal browser navigation receives the complete page.

An HTMX request usually receives only the relevant fragment.

## Detecting an HTMX 4 request

HTMX 4 introduces:

```http
HX-Request-Type: partial
```

or:

```http
HX-Request-Type: full
```

It also sends an `Accept: text/html` header. ([htmx][2])

Helper:

```go
func isPartialRequest(r *http.Request) bool {
    return r.Header.Get("HX-Request-Type") == "partial"
}
```

Handler example:

```go
func (app *Application) showProject(
    w http.ResponseWriter,
    r *http.Request,
) {
    data, err := app.projects.GetPageData(
        r.Context(),
        r.PathValue("workspaceSlug"),
        r.PathValue("projectSlug"),
    )
    if err != nil {
        app.handleError(w, r, err)
        return
    }

    if isPartialRequest(r) {
        app.render(w, http.StatusOK, "project-content", data)
        return
    }

    app.render(w, http.StatusOK, "project-page", data)
}
```

---

# 11. HTTP conventions

Use normal HTTP semantics.

```text
GET     Read a page or fragment
POST    Create a resource
PATCH   Update part of a resource
DELETE  Delete or archive a resource
```

Examples:

```text
GET    /w/acme/projects/platform
POST   /w/acme/projects/platform/issues
PATCH  /w/acme/issues/42/status
PATCH  /w/acme/issues/42/assignee
POST   /w/acme/issues/42/comments
DELETE /w/acme/issues/42/comments/7
```

Response statuses:

```text
200 OK                    Successful rendered response
201 Created               Resource created
204 No Content            Successful action with no swap
303 See Other             Non-HTMX form redirect
400 Bad Request           Malformed request
401 Unauthorized          User not authenticated
403 Forbidden             User lacks permission
404 Not Found             Resource unavailable
409 Conflict              Duplicate or state conflict
422 Unprocessable Entity  Form validation failed
429 Too Many Requests     Rate limit exceeded
500 Internal Server Error Unexpected server failure
```

In HTMX 4, error response HTML is swapped by default. Therefore, error responses should contain useful HTML fragments rather than plain strings. ([htmx][2])

---

# 12. Database specification

## PostgreSQL tables

```text
users
sessions
email_verification_tokens
password_reset_tokens

workspaces
workspace_members
workspace_invitations

projects
issues
issue_comments

labels
issue_labels

activity_events
```

## Simplified relationships

```text
users
  ├── sessions
  ├── workspace_members
  ├── issues assigned to user
  └── comments

workspaces
  ├── workspace_members
  ├── projects
  ├── labels
  └── activity_events

projects
  └── issues

issues
  ├── comments
  ├── labels
  └── activity_events
```

## IDs

Use UUIDs internally:

```sql
id UUID PRIMARY KEY DEFAULT gen_random_uuid()
```

Use human-readable issue numbers inside a project:

```text
FORGE-1
FORGE-2
FORGE-3
```

Store both:

```text
id: UUID
issue_number: INTEGER
```

Create a unique constraint:

```sql
UNIQUE (project_id, issue_number)
```

---

# 13. Authentication specification

## Authentication type

Use server-side sessions.

Do not use JWT for the web interface.

## Session cookie

```go
http.Cookie{
    Name:     "__Host-forgeboard_session",
    Value:    rawToken,
    Path:     "/",
    HttpOnly: true,
    Secure:   production,
    SameSite: http.SameSiteLaxMode,
    MaxAge:   60 * 60 * 24 * 7,
}
```

The `__Host-` prefix only works when:

* `Secure` is enabled
* `Path=/`
* no `Domain` attribute is set

For local HTTP development, use a simpler cookie name such as:

```text
forgeboard_session
```

## Session database representation

```text
id
user_id
token_hash
created_at
last_seen_at
expires_at
user_agent
ip_address
revoked_at
```

Never store the raw session token.

Store:

```go
sha256(rawToken)
```

## Password hashing

Use Argon2id.

The application should:

1. generate a random salt
2. hash the password using Argon2id
3. store algorithm parameters with the hash
4. compare in constant time
5. permit future rehashing when parameters change

Do not implement Argon2 itself. Implement the surrounding password service yourself using Go’s cryptographic library.

---

# 14. CSS specification

No Tailwind, Sass, PostCSS or CSS build step.

Use plain modern CSS.

## Browser strategy

The demo targets current evergreen browsers:

* Chrome
* Edge
* Firefox
* Safari

Because this is a learning project, we can intentionally use modern features without supporting old browsers.

## CSS features to use

* native CSS nesting
* cascade layers
* custom properties
* `color-mix()`
* `oklch()`
* container queries
* logical properties
* `:has()`
* `:is()`
* `:where()`
* `clamp()`
* CSS grid
* subgrid where useful
* view transitions
* `prefers-color-scheme`
* `prefers-reduced-motion`
* `light-dark()`
* individual transform properties

## Example foundation

```css
@layer reset, tokens, base, layout, components, utilities;

@layer tokens {
    :root {
        color-scheme: light dark;

        --color-background: light-dark(
            oklch(98% 0.005 250),
            oklch(18% 0.015 250)
        );

        --color-surface: light-dark(
            oklch(100% 0 0),
            oklch(23% 0.015 250)
        );

        --color-text: light-dark(
            oklch(22% 0.015 250),
            oklch(94% 0.005 250)
        );

        --color-primary: oklch(60% 0.18 255);
        --color-border: color-mix(
            in oklch,
            var(--color-text) 15%,
            transparent
        );

        --space-1: 0.25rem;
        --space-2: 0.5rem;
        --space-3: 0.75rem;
        --space-4: 1rem;
        --space-6: 1.5rem;
        --space-8: 2rem;

        --radius-small: 0.4rem;
        --radius-medium: 0.7rem;
        --radius-large: 1rem;
    }
}
```

## Native nesting example

```css
.issue-card {
    display: grid;
    gap: var(--space-3);
    padding: var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-medium);
    background: var(--color-surface);

    &:hover {
        border-color: var(--color-primary);
    }

    & .issue-card__header {
        display: flex;
        justify-content: space-between;
        gap: var(--space-3);
    }

    & .issue-card__title {
        font-weight: 600;
    }

    &[data-priority="high"] {
        border-inline-start: 0.25rem solid var(--color-danger);
    }
}
```

## Component strategy

Use classes with simple component naming:

```text
.button
.button--primary
.button--danger

.form-field
.form-field__label
.form-field__input
.form-field__error

.issue-card
.issue-card__header
.issue-card__title
.issue-card__meta
```

Avoid extremely strict BEM. Use predictable names, but keep the CSS readable.

---

# 15. Minimal JavaScript policy

The application should work mainly through:

* HTML
* HTMX
* CSS
* Go-rendered responses

Custom JavaScript is allowed only for behavior that HTMX and native HTML do not express cleanly.

Possible uses:

* dialog open and close behavior
* small keyboard shortcuts
* copying issue links
* preserving temporary UI preferences
* enhancing form focus after a swap

Target:

```text
Under 200 lines of custom JavaScript
```

Do not use:

* React
* Alpine.js
* Stimulus
* jQuery
* a frontend bundler
* TypeScript
* npm during normal development

---

# 16. Docker development environment

Services:

```text
app
database
mailpit
```

## `compose.yaml`

```yaml
services:
  app:
    build:
      context: .
      target: development
    ports:
      - "8080:8080"
    environment:
      APP_ENV: development
      APP_ADDRESS: ":8080"
      DATABASE_URL: postgres://forgeboard:forgeboard@database:5432/forgeboard?sslmode=disable
      SMTP_HOST: mailpit
      SMTP_PORT: "1025"
    volumes:
      - .:/app
    depends_on:
      database:
        condition: service_healthy

  database:
    image: postgres:18-alpine
    environment:
      POSTGRES_DB: forgeboard
      POSTGRES_USER: forgeboard
      POSTGRES_PASSWORD: forgeboard
    ports:
      - "5432:5432"
    volumes:
      - forgeboard_database:/var/lib/postgresql/data
    healthcheck:
      test:
        - CMD-SHELL
        - pg_isready -U forgeboard -d forgeboard
      interval: 5s
      timeout: 5s
      retries: 10

  mailpit:
    image: axllent/mailpit
    ports:
      - "8025:8025"
      - "1025:1025"

volumes:
  forgeboard_database:
```

---

# 17. Makefile specification

The primary command should be:

```bash
make dev
```

It should:

1. build the development image
2. start PostgreSQL
3. start Mailpit
4. run database migrations
5. start the application
6. show application logs

## Suggested Makefile

```makefile
.DEFAULT_GOAL := help

APP_NAME := forgeboard
COMPOSE := docker compose

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make dev       Start the complete development environment"
	@echo "  make stop      Stop containers"
	@echo "  make reset     Delete containers and database data"
	@echo "  make test      Run all tests"
	@echo "  make lint      Run Go checks"
	@echo "  make migrate   Apply database migrations"
	@echo "  make seed      Insert development data"
	@echo "  make logs      Follow application logs"

.PHONY: dev
dev:
	$(COMPOSE) up --build

.PHONY: stop
stop:
	$(COMPOSE) down

.PHONY: reset
reset:
	$(COMPOSE) down --volumes --remove-orphans

.PHONY: test
test:
	$(COMPOSE) run --rm app go test ./...

.PHONY: lint
lint:
	$(COMPOSE) run --rm app sh -c \
		"go vet ./... && gofmt -l ."

.PHONY: migrate
migrate:
	$(COMPOSE) run --rm app \
		go run ./cmd/migrate up

.PHONY: seed
seed:
	$(COMPOSE) run --rm app \
		go run ./cmd/seed

.PHONY: logs
logs:
	$(COMPOSE) logs --follow app
```

However, migrations will not automatically run before the app with this exact setup. The cleanest version is for the app container’s development entrypoint to run:

```sh
go run ./cmd/migrate up
go run ./cmd/web
```

Then this remains true:

```bash
git clone ...
cd forgeboard
make dev
```

And the application becomes available at:

```text
Application: http://localhost:8080
Mailpit:     http://localhost:8025
PostgreSQL: localhost:5432
```

---

# 18. Development milestones

## Milestone 1: Foundation

Deliver:

* Go project
* `net/http` server
* graceful shutdown
* templates
* embedded static files
* PostgreSQL connection
* migrations
* Docker Compose
* `make dev`
* health endpoint
* basic CSS foundation

Routes:

```text
GET /health
GET /
```

## Milestone 2: Authentication

Deliver:

* registration
* login
* logout
* sessions
* auth middleware
* password hashing
* login rate limiting
* protected dashboard
* account session list

## Milestone 3: Workspaces

Deliver:

* create workspace
* workspace slug
* membership
* authorization
* workspace switcher
* member list
* invite flow

## Milestone 4: Projects and issues

Deliver:

* project creation
* issue creation
* issue list
* issue detail
* status changes
* priority changes
* assignment
* archive issue

## Milestone 5: HTMX experience

Deliver:

* inline validation
* partial issue creation
* filtering
* updating status without reload
* comment append
* multiple target updates
* loading indicators
* disabled submit controls
* browser history
* error fragments

## Milestone 6: Portfolio quality

Deliver:

* activity log
* seeded demo account
* integration tests
* screenshots
* architecture diagram
* threat model
* CI workflow
* production Docker image
* polished README

---

# 19. Definition of done

The project is complete when a reviewer can:

```text
1. Clone the repository
2. Run make dev
3. Open localhost:8080
4. Register a user
5. Create a workspace
6. Create a project
7. Create and update issues
8. Add comments
9. Invite another user
10. Observe authorization between workspaces
11. Run make test successfully
```

The demo should contain at least:

* 20 meaningful HTTP routes
* 10 database tables
* 8 reusable HTML fragments
* 5 middleware components
* 3 role or permission levels
* handler tests
* authorization tests
* repository integration tests
* one complete password-reset flow
* one multi-target HTMX update
* one `422` validation fragment
* one transaction involving an activity event

That gives us a project that is **not tiny**, but remains controlled. It demonstrates considerably more than CRUD without growing into a six-month Jira clone.

[1]: https://four.htmx.org/ "htmx"
[2]: https://four.htmx.org/docs "Documentation ~ htmx"
[3]: https://go.dev/blog/routing-enhancements "Routing Enhancements for Go 1.22 - The Go Programming Language"
[4]: https://pkg.go.dev/net/http "http package - net/http - Go Packages"
