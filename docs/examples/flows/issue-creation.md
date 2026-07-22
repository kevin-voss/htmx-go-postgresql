# Flow — Issue creation

## Diagram

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

## Success UX

- No full page reload.
- New issue card appended (or list refreshed) via HTMX.
- Ideally update count / activity in the same response (multi-target / `<hx-partial>` once that step exists).

## Validation failure

- `422` + HTML fragment into form error region (`hx-status:422=...`).

## Related

- HTMX: [../../specs/htmx-decision.md](../../specs/htmx-decision.md)
- Statuses: [issue-status.md](issue-status.md)
- Database issue numbers: [../../architecture/database.md](../../architecture/database.md)
