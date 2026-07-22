# Database specification

## Tables

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

## Relationships (simplified)

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

Human-readable issue numbers **within a project**:

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

Constraint:

```sql
UNIQUE (project_id, issue_number)
```

(Display prefix may come from project/workspace key; storage is the integer `issue_number` per project.)

## Tooling

- Migrations: **goose**
- Driver: **pgx**
- Optional later: **sqlc** under `db/queries/`

## Related

- Auth tables: [../specs/authentication.md](../specs/authentication.md)
- Structure: [project-structure.md](project-structure.md)
- Docker Postgres: [docker.md](docker.md)
