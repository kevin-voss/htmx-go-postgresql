# Step 32: CI, production image, README polish

| Field | Value |
| ----- | ----- |
| ID | `STEP-32` |
| Milestone | M6 — Portfolio quality |
| Status | `todo` |
| Depends on | STEP-31 |
| Unlocks | — (project complete pending DoD walkthrough) |
| Estimated scope | L |

---

## Goal

CI runs tests; production Docker target exists; README enables clone→make dev walkthrough with architecture/threat-model notes.

## Description

Final portfolio packaging. Satisfy definition of done checklist items for CI, prod image, README, diagrams/notes.

## References

- DoD: [DEFINITION_OF_DONE.md](../../DEFINITION_OF_DONE.md)
- Milestones: [milestones.md](../milestones.md)
- Product: [product.md](../../specs/product.md)

## Prerequisites

- Tests green locally.

## Scope

### In

- CI workflow (GitHub Actions or equivalent)
- Dockerfile production target (multi-stage)
- Polished README: stack, make dev, demo login, screenshots placeholders
- Architecture diagram (mermaid or image)
- Threat model notes (sessions, CSRF, XSS via templates, SQLi via params)
- Verify quantitative DoD minimums or document gaps

### Out

- Kubernetes manifests
- Paid hosting setup

## Implementation checklist

- [ ] Add CI YAML
- [ ] Production image builds
- [ ] README complete
- [ ] Walk DoD reviewer list manually once

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `.github/workflows/ci.yml` | create | or equivalent |
| `Dockerfile` | modify | production stage |
| `README.md` | modify | portfolio quality |
| `docs/architecture/*` | optional | diagram link |

## Technical notes

Do not commit secrets. Pin actions versions. Keep HTMX pin mentioned in README.

## Acceptance criteria

- [ ] CI runs make test or equivalent on PRs
- [ ] Production image builds successfully
- [ ] README documents make dev and demo credentials
- [ ] Architecture diagram present
- [ ] Threat model notes present
- [ ] Reviewer can complete DoD walkthrough

## Verification

```bash
make test
docker build --target production -t forgeboard:prod .
# manual DoD walkthrough
```

## Commit

**Subject (required):**

```text
chore(step-32): add CI, production image, and portfolio README
```

**Body (optional):**

```text
Complete STEP-32 so the next agent can continue from a green tree.
```

## Handoff to next agent

Project implementation plan complete. Validate against DEFINITION_OF_DONE.md.

After commit, mark this step `done` in any tracker and **stop** — do not start — (project complete pending DoD walkthrough).
