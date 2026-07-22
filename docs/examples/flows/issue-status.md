# Flow — Issue status

## Statuses (v1)

```text
Backlog
Todo
In Progress
Done
```

## Where status can change

- issue detail page
- project issue list
- board view

## Interaction model (v1)

Use **buttons** or a `<select>` — **not** drag and drop.

Drag and drop would require extra JavaScript and is not needed to demonstrate HTMX.

## Request shape (example)

```text
PATCH /w/{workspaceSlug}/issues/{issueNumber}/status
```

Prefer HTMX partial update of the card / column without full reload.

## Authorization

- Member and Owner can change status.
- Viewer cannot.

## Related

- Roles: [../../specs/roles.md](../../specs/roles.md)
- HTTP: [../../specs/http-conventions.md](../../specs/http-conventions.md)
