# Step 03: Config & structured logging

| Field | Value |
| ----- | ----- |
| ID | `STEP-03` |
| Milestone | M1 — Foundation |
| Status | `todo` |
| Depends on | STEP-02 |
| Unlocks | STEP-04 |
| Estimated scope | S |

---

## Goal

Application configuration loads from environment variables and logs via log/slog.

## Description

Centralize env loading in internal/config and initialize structured logging. Downstream steps must not scrape os.Getenv ad hoc for core settings.

## References

- Stack: [technology-stack.md](../../specs/technology-stack.md)
- Structure: [project-structure.md](../../architecture/project-structure.md)

## Prerequisites

- STEP-01/02 done.

## Scope

### In

- Config struct: APP_ENV, APP_ADDRESS, DATABASE_URL, SMTP_HOST, SMTP_PORT, cookie/secure flags as needed
- Validation of required fields
- slog default logger with level by APP_ENV
- Wire config load from cmd/web

### Out

- HTTP server (step 04).
- Mail sending (step 14).

## Implementation checklist

- [ ] Implement internal/config
- [ ] Fail fast on missing required env in non-test
- [ ] Initialize slog in main
- [ ] Unit test for config defaults/validation if practical

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `internal/config/config.go` | create | Load() + Config type |
| `cmd/web/main.go` | modify | load config, setup logger |

## Technical notes

Prefer explicit errors over silent defaults for DATABASE_URL. Development defaults for APP_ADDRESS=:8080 are OK.

## Acceptance criteria

- [ ] Missing DATABASE_URL fails loudly when required
- [ ] Logs are structured (slog JSON or text) and include level
- [ ] Config is injectable into later Application struct

## Verification

```bash
go test ./internal/config/...
go build ./...
```

## Commit

**Subject (required):**

```text
feat(step-03): add env config and structured logging
```

**Body (optional):**

```text
Complete STEP-03 so the next agent can continue from a green tree.
```

## Handoff to next agent

Config fields available: list them in notes. Logger is global or passed via Application.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-04.
