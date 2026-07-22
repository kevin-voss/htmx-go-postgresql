# Step 10: Argon2id password service

| Field | Value |
| ----- | ----- |
| ID | `STEP-10` |
| Milestone | M2 — Authentication |
| Status | `done` |
| Depends on | STEP-09 |
| Unlocks | STEP-11 |
| Estimated scope | S |

---

## Goal

A password service can hash and verify passwords with Argon2id, storing parameters with the hash.

## Description

Implement cryptography boundaries before any user table. No HTTP yet — pure package + tests.

## References

- Auth spec: [authentication.md](../../specs/authentication.md)

## Prerequisites

- Go module builds.

## Scope

### In

- Hash(password) → encoded hash string
- Compare(password, encoded) in constant-time friendly way
- Random salt per password
- Unit tests for round-trip and mismatch
- Document parameters (time, memory, threads)

### Out

- User persistence
- Login UI

## Implementation checklist

- [ ] Implement internal/auth/password.go (or package)
- [ ] Add tests
- [ ] Do not roll your own Argon2 primitive

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/auth/password.go` | create | hash/compare |
| `internal/auth/password_test.go` | create | tests |

## Technical notes

Use golang.org/x/crypto/argon2. Permit future rehashing by storing params.

## Acceptance criteria

- [ ] Correct password verifies
- [ ] Wrong password fails
- [ ] Two hashes of same password differ (salt)
- [ ] No plaintext password logging

## Verification

```bash
go test ./internal/auth/...
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(auth): add Argon2id password hashing service
```

**Body:**

```text
Provide a tested password boundary before any user table so credentials
are never stored or compared incorrectly.

STEP-10
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-11

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

Encoded hash format: `$argon2id$v=19$m=65536,t=3,p=4$<salt_b64>$<hash_b64>` (PHC). Params: time=3, memory=64MiB, threads=4, salt=16B, key=32B. API: `auth.Hash` / `auth.Compare` / `auth.NeedsRehash`. Ready for users table.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-11.
