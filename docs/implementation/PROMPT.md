# Step prompt (change only `NN`)

Copy everything below into a new agent chat. Replace **both** `NN` placeholders with the step number (`01`, `02`, … `32`).

---

```text
Do STEP-NN only per docs/AGENT_GUIDE.md and docs/implementation/steps/NN-*.md.

Rules:
- Read the step file + all linked references first
- Stay strictly in that step’s In/Out scope
- Meet every acceptance criterion
- Run the Verification commands
- Update docs/implementation/STATUS.md for this step only
- Do not start the next step

When verification passes, finish “Commit & push”:
1. Commit using the step’s subject + body in conventional form (see the single example in docs/AGENT_GUIDE.md)
2. Push to origin (`git push -u origin HEAD`)
3. Confirm git status is clean and not ahead of remote
4. Stop

This message authorizes commit and push for STEP-NN only.
```
