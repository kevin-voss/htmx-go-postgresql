# Pages and routes

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

Also expected early:

```text
GET /health
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

## Mutation examples (not exhaustive)

```text
POST   /register
POST   /login
POST   /logout
POST   /w/{workspaceSlug}/projects
POST   /w/{workspaceSlug}/projects/{projectSlug}/issues
PATCH  /w/{workspaceSlug}/issues/{issueNumber}/status
PATCH  /w/{workspaceSlug}/issues/{issueNumber}/assignee
POST   /w/{workspaceSlug}/issues/{issueNumber}/comments
DELETE /w/{workspaceSlug}/issues/{issueNumber}/comments/{commentID}
```

See [http-conventions.md](http-conventions.md) for method/status semantics.

## Related

- HTTP conventions: [http-conventions.md](http-conventions.md)
- Rendering (full page vs fragment): [../architecture/rendering.md](../architecture/rendering.md)
- Flows: [../examples/flows/](../examples/flows/)
