# Step 24: Labels

| Field | Value |
| ----- | ----- |
| ID | `STEP-24` |
| Milestone | M4 — Projects & issues |
| Status | `todo` |
| Depends on | STEP-23 |
| Unlocks | STEP-25 |
| Estimated scope | M |

---

## Goal

Workspaces have labels; issues can be tagged via issue_labels.

## Description

Add labels and issue_labels tables. Owners/members manage labels; viewers read.

## References

- DB: [database.md](../../architecture/database.md)
- Product scope: [product.md](../../specs/product.md)

## Prerequisites

- Issues exist.

## Scope

### In

- Migrations: labels, issue_labels
- CRUD labels (minimal)
- Attach/detach on issue
- Show labels on cards/detail

### Out

- Fancy label colors picker beyond simple color field

## Implementation checklist

- [ ] migrations
- [ ] service/handlers
- [ ] UI to add label to issue

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `db/migrations/*_labels.sql` | create |  |
| `internal/issue or label module` | create/modify |  |
| `web/templates/**` | modify |  |

## Technical notes

Labels are workspace-scoped per DB diagram.

## Acceptance criteria

- [ ] Create label in workspace
- [ ] Attach label to issue
- [ ] Remove label from issue
- [ ] Labels visible on issue UI
- [ ] Non-member cannot manage labels

## Verification

```bash
go test ./internal/issue/...
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(labels): add workspace labels and issue tagging
```

**Body:**

```text
Enable lightweight categorization across issues without introducing
external tagging systems.

STEP-24
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-25

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Label fields: name, color, workspace_id.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-25.
