# Step 14: Mail & email verification

| Field | Value |
| ----- | ----- |
| ID | `STEP-14` |
| Milestone | M2 — Authentication |
| Status | `todo` |
| Depends on | STEP-13 |
| Unlocks | STEP-15 |
| Estimated scope | M |

---

## Goal

App sends email via Mailpit SMTP; registration issues a verification token; user can verify email.

## Description

Introduce internal/mail and email_verification_tokens. In development, messages appear in Mailpit UI.

## References

- Flow: [registration.md](../../examples/flows/registration.md)
- Docker: [docker.md](../../architecture/docker.md)
- DB: [database.md](../../architecture/database.md)

## Prerequisites

- Mailpit in Compose.
- Registration exists.

## Scope

### In

- SMTP mailer using SMTP_HOST/PORT
- Migration: email_verification_tokens
- Send on register
- GET /verify-email (+ token query or path)
- Mark user verified

### Out

- Production email provider
- Fancy HTML templates beyond simple body

## Implementation checklist

- [ ] mail package
- [ ] verification token create/consume
- [ ] wire register → send
- [ ] manual check in Mailpit

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/mail/*.go` | create | SMTP |
| `db/migrations/*_email_verification.sql` | create | tokens |
| `web/templates/pages/verify_email*.html` | create | result pages |

## Technical notes

Tokens hashed at rest like sessions if possible. Expiry required.

## Acceptance criteria

- [ ] Registration produces a message visible in Mailpit
- [ ] Valid verification link marks email verified
- [ ] Invalid/expired token shows error page
- [ ] No SMTP credentials committed

## Verification

```bash
make dev
# register user, open Mailpit :8025, click link
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(mail): add Mailpit mailer and email verification
```

**Body:**

```text
Send verification messages in development and mark emails verified so
account trust can grow without a real SMTP provider yet.

STEP-14
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-15

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Verify route shape: ____. Token TTL: ____.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-15.
