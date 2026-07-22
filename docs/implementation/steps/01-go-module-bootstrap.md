# Step 01: Go module & package skeleton

| Field | Value |
| ----- | ----- |
| ID | `STEP-01` |
| Milestone | M1 — Foundation |
| Status | `todo` |
| Depends on | — (first step) |
| Unlocks | STEP-02 |
| Estimated scope | S |

---

## Goal

A compilable Go module exists with the agreed package layout and a minimal cmd/web entrypoint.

## Description

Bootstrap the Forgeboard repository as a Go module. Create the empty modular-monolith package tree so later steps drop code into predictable locations. Do not implement product features yet — only scaffolding that `go build ./...` can succeed against.

## References

- Architecture: [project-structure.md](../../architecture/project-structure.md)
- Stack: [technology-stack.md](../../specs/technology-stack.md)
- Agent guide: [AGENT_GUIDE.md](../../AGENT_GUIDE.md)

## Prerequisites

- Empty/new repo with docs already present.
- Go toolchain available (locally or later via Docker).

## Scope

### In

- Initialize go.mod with a sensible module path (e.g. github.com/<user>/forgeboard or module forgeboard).
- Create cmd/web/main.go that prints a startup message or exits 0 after minimal setup.
- Create empty internal package directories matching architecture (app, auth, config, database, platform, …) with .gitkeep or stub packages as needed.
- Add .gitignore for binaries, .env, IDE junk.
- Add .env.example with placeholder keys (no secrets).

### Out

- Docker, HTTP routes, DB, templates, CSS.

## Implementation checklist

- [ ] Create go.mod
- [ ] Create cmd/web/main.go
- [ ] Create internal/* skeleton dirs
- [ ] Add .gitignore and .env.example
- [ ] Verify go build ./...

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `go.mod` | create | module path decision documented in handoff |
| `cmd/web/main.go` | create | minimal main |
| `internal/**` | create | package skeleton |
| `.gitignore` | create | standard Go ignores |
| `.env.example` | create | APP_ENV, APP_ADDRESS, DATABASE_URL, SMTP_* placeholders |

## Technical notes

Use net/http later — do not add Chi/Gin/Echo. Do not introduce sqlc yet. Module path should be stable; changing it later is painful.

## Acceptance criteria

- [ ] `go build ./...` succeeds
- [ ] Directory layout matches docs/architecture/project-structure.md at a high level
- [ ] .env.example exists and contains no real secrets
- [ ] .gitignore excludes .env and build artifacts

## Verification

```bash
go build ./...
ls cmd/web internal
```

## Commit

**Subject (required):**

```text
chore(step-01): bootstrap Go module and package layout
```

**Body (optional):**

```text
Complete STEP-01 so the next agent can continue from a green tree.
```

## Handoff to next agent

Module path is ____. Skeleton packages may be empty; step 02 adds Compose.

After commit, mark this step `done` in any tracker and **stop** — do not start STEP-02.
