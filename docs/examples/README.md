# Examples

Worked examples of end-to-end behavior. Prefer these over inventing new UX during implementation.

## Flows

| Flow | File |
| ---- | ---- |
| Registration | [flows/registration.md](flows/registration.md) |
| Login | [flows/login.md](flows/login.md) |
| First-time onboarding | [flows/onboarding.md](flows/onboarding.md) |
| Invitation | [flows/invitation.md](flows/invitation.md) |
| Issue creation | [flows/issue-creation.md](flows/issue-creation.md) |
| Issue status | [flows/issue-status.md](flows/issue-status.md) |
| Comments (multi-partial) | [flows/comments.md](flows/comments.md) |

## How agents should use flows

1. Implement the matching step(s) in [../implementation/steps/](../implementation/steps/).
2. Match request/response and UX described here.
3. Do not invent drag-and-drop, OAuth, or extra screens unless a step says so.
