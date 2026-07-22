# Step 10: Argon2id password service

| Field | Value |
| ----- | ----- |
| ID | `STEP-10` |
| Milestone | M2 — Authentication |
| Status | `todo` |
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

## Commit

**Subject (required):**

```text
feat(step-10): add Argon2id password hashing service
```

**Body (optional):**

```text
Complete STEP-10 so the next agent can continue from a green tree.
```

## Handoff to next agent

Encoded hash format: ____. Ready for users table.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-11.
