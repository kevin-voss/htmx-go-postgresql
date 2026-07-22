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

## Commit

**Subject (required):**

```text
feat(step-24): add labels and issue labeling
```

**Body (optional):**

```text
Complete STEP-24 so the next agent can continue from a green tree.
```

## Handoff to next agent

Label fields: name, color, workspace_id.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-25.
