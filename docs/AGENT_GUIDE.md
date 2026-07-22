# Agent guide — one step, one agent, one commit, one push

This repository is designed so **one AI agent implements one implementation step**, then **commits and pushes**, then stops. The next agent (or session) picks up the next unblocked step.

---

## Golden rules

1. **Read before write** — open the step file, every linked spec/architecture/flow doc, and the previous step’s handoff notes.
2. **One step only** — do not start the next step, even if it looks trivial.
3. **Stay in scope** — implement only what the step lists under Scope / Files. No drive-by refactors.
4. **Prove it** — meet every acceptance criterion; run the listed verification commands.
5. **Commit and push** — after verification, create **exactly one** commit using the step’s subject + body, then **push to `origin`**. Mandatory — do not wait for the human to ask again.
6. **Hand off cleanly** — update `docs/implementation/STATUS.md`; leave the tree green and remote up to date.

---

## Session workflow

```text
1. Open docs/implementation/PROMPT.md (change NN only)
2. Open docs/implementation/steps/NN-....md
3. Read all “References” links in that step
4. Mark STATUS.md as in_progress for this step
5. Implement checklist items
6. Run verification + acceptance criteria checks
7. Update STATUS.md to done
8. Commit (format below) + push
9. Confirm git status clean / not ahead of origin
10. Stop — do not start the next step
```

---

## Commit message format (best practice)

Use [Conventional Commits](https://www.conventionalcommits.org/):

```text
type(scope): short imperative summary

Body explains why this change exists and what it enables.
Wrap near 72 chars. No secrets. Reference STEP-NN once in the body.
```

| Part | Rule |
| ---- | ---- |
| `type` | Usually `feat`. Use `fix`, `test`, `chore`, `ci`, `docs` only when that is the primary nature of the change. |
| `scope` | Product/area name (`auth`, `issues`, `docker`) — **not** `step-01`. |
| Summary | Imperative, lowercase after colon, no period, ~50 chars. |
| Body | Required. Focus on **why** / outcome. Mention `STEP-NN` for traceability. |

### Example (the only full example — copy this shape)

```bash
git add <paths for this step> docs/implementation/STATUS.md
git commit -m "$(cat <<'EOF'
feat(auth): add session-based login and logout

Server-side sessions let users stay signed in without JWTs,
matching the auth spec and unblocking protected /app routes.

STEP-12
EOF
)"
git push -u origin HEAD
git status
```

Each step file lists **its** subject and body under “Commit & push”. Reuse the command shape above; substitute that step’s subject/body.

### Other rules

- Never commit `.env` or secrets.
- Never skip hooks unless the human explicitly requests it.
- Never force-push to `main`/`master`.
- One step → one commit → one push.
- If a hook rejects the commit: fix, create a **new** commit (do not amend unless amend rules are fully satisfied), then push.

---

## What “done” means for a step

- [ ] Every acceptance criterion is satisfied
- [ ] Verification commands pass
- [ ] `STATUS.md` updated for this step
- [ ] One conventional commit (`type(scope): …` + body) created
- [ ] Pushed to `origin`; working tree clean; branch not ahead
- [ ] Stopped without starting the next step

---

## Where to look for answers

| Question | Doc |
| -------- | --- |
| Copy-paste prompt | [implementation/PROMPT.md](implementation/PROMPT.md) |
| Product | [specs/product.md](specs/product.md) |
| Roles | [specs/roles.md](specs/roles.md) |
| Routes | [specs/pages-and-routes.md](specs/pages-and-routes.md) |
| Architecture | [architecture/overview.md](architecture/overview.md) |
| Structure | [architecture/project-structure.md](architecture/project-structure.md) |
| HTMX | [specs/htmx-decision.md](specs/htmx-decision.md) |
| Flows | [examples/flows/](examples/flows/) |
| Project done? | [DEFINITION_OF_DONE.md](DEFINITION_OF_DONE.md) |

---

## Failure / partial completion

If blocked: stop before a bad commit; document the blocker; do not push broken work; do not invent spec violations.

If verification fails: fix in scope, re-verify, then commit and push.
