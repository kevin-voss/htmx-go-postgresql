# Step NN: Title

| Field | Value |
| ----- | ----- |
| ID | `STEP-NN` |
| Milestone | M? — Name |
| Status | `todo` \| `in_progress` \| `done` |
| Depends on | `STEP-XX` |
| Unlocks | `STEP-YY` |
| Estimated scope | S / M / L |

---

## Goal

One sentence: what capability exists when this step is finished.

## Description

2–5 paragraphs: context, why this step exists, and how it fits the architecture.

## References

- Spec: …
- Architecture: …
- Flow: …

## Prerequisites

- Previous step(s) completed, committed, and pushed.
- Any tools/services that must be running.

## Scope

### In

- …

### Out

- … (explicitly deferred to later steps)

## Implementation checklist

- [ ] …
- [ ] …
- [ ] …

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `…` | create / modify | … |

## Technical notes

Constraints, snippets, pitfalls, security reminders.

## Acceptance criteria

- [ ] …
- [ ] …
- [ ] …

## Verification

```bash
# commands the agent must run
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (one example there). Subject and body for **this** step:

**Subject:**

```text
feat(scope): short imperative summary
```

**Body:**

```text
Why this change matters and what it unlocks.

STEP-NN
```

**Required actions:**

- [ ] Update `STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with subject + body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start the next step

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

What the next step can assume is true.
