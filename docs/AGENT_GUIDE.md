# Agent guide — one step, one agent, one commit

This repository is designed so **one AI agent implements one implementation step**, then stops. The next agent (or session) picks up the next unblocked step.

---

## Golden rules

1. **Read before write** — open the step file, every linked spec/architecture/flow doc, and the previous step’s handoff notes.
2. **One step only** — do not start the next step, even if it looks trivial.
3. **Stay in scope** — implement only what the step lists under Scope / Files. No drive-by refactors.
4. **Prove it** — meet every acceptance criterion; run the listed verification commands.
5. **Commit once** — after the step passes verification, create **exactly one** commit using the step’s commit message (user must request the commit unless your session protocol says otherwise — follow the human’s git rules).
6. **Hand off cleanly** — update status notes if a tracker is used; leave the tree green.

---

## Session workflow

```text
1. Open docs/implementation/README.md
2. Identify the next incomplete step (lowest number not done)
3. Open docs/implementation/steps/NN-....md
4. Read all “References” links in that step
5. Claim the work (bd / notes / PR description — whatever the project uses)
6. Implement checklist items
7. Run verification + acceptance criteria checks
8. Commit with the prescribed message
9. Mark step complete; stop
```

Do **not** batch multiple steps into one session unless a human explicitly asks.

---

## What “done” means for a step

A step is complete only when **all** of the following are true:

- [ ] Every acceptance criterion is satisfied
- [ ] Verification commands pass
- [ ] No unrelated files changed
- [ ] Commit message matches the step (conventional commit)
- [ ] Handoff section in the step (or tracker note) mentions anything the next agent must know

---

## Where to look for answers

| Question | Doc |
| -------- | --- |
| What is the product? | [specs/product.md](specs/product.md) |
| Who can do what? | [specs/roles.md](specs/roles.md) |
| Which routes exist? | [specs/pages-and-routes.md](specs/pages-and-routes.md) |
| How are layers split? | [architecture/overview.md](architecture/overview.md) |
| Where do files go? | [architecture/project-structure.md](architecture/project-structure.md) |
| How does HTMX behave? | [specs/htmx-decision.md](specs/htmx-decision.md), [architecture/rendering.md](architecture/rendering.md) |
| How does a user journey work? | [examples/flows/](examples/flows/) |
| Is the whole project finished? | [DEFINITION_OF_DONE.md](DEFINITION_OF_DONE.md) |

---

## Commit discipline

- Use the **exact commit subject** from the step (or a trivially equivalent conventional form).
- Prefer body text that explains *why* when non-obvious.
- Never commit secrets (`.env`, credentials, private keys).
- Never skip hooks unless the human explicitly requests it.
- One step → one commit. If a step needs a follow-up fix, that is a new commit referencing the same step id in the message, e.g. `fix(step-14): correct session cookie flags`.

---

## Out of scope for agents on implementation steps

- Rewriting these docs (unless the step is a docs step)
- Expanding product scope beyond [specs/product.md](specs/product.md)
- Introducing frameworks banned by the specs (Chi/Gin/Echo, React, Tailwind, Alpine, etc.)
- Upgrading HTMX past the pinned version without a human decision

---

## Failure / partial completion

If blocked:

1. Stop before a half-baked commit if possible.
2. Document the blocker in the step handoff or issue tracker.
3. Do not invent workarounds that violate specs (e.g. switching to JWT, adding OAuth).

If verification fails:

1. Fix within the step’s scope.
2. Re-run verification.
3. Only then commit.
