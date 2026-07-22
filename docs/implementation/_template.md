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

- Previous step(s) completed and committed.
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

## Commit

**Subject (required):**

```text
type(step-NN): description
```

**Body (optional):**

```text
Why this change matters.
```

## Handoff to next agent

What the next step can assume is true. Any landmines, TODOs left intentional, env vars added, etc.
