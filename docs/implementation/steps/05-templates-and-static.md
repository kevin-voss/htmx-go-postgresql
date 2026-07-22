# Step 05: Templates & static files

| Field | Value |
| ----- | ----- |
| ID | `STEP-05` |
| Milestone | M1 — Foundation |
| Status | `todo` |
| Depends on | STEP-04 |
| Unlocks | STEP-06 |
| Estimated scope | M |

---

## Goal

html/template rendering and static file serving (embed or filesystem) are available with layouts/pages/fragments folders.

## Description

Establish the rendering platform used by every page and HTMX fragment. Prefer embed for production simplicity; development may still use embed for parity unless a step later chooses FS reload.

## References

- Rendering: [rendering.md](../../architecture/rendering.md)
- Structure: [project-structure.md](../../architecture/project-structure.md)

## Prerequisites

- HTTP server runs.

## Scope

### In

- web/templates/{layouts,pages,components,fragments}
- web/static skeleton
- internal/platform/render helper
- Serve /static/...
- Base layout with title block

### Out

- Full CSS design (step 06).
- HTMX vendor (step 26) — can leave script placeholder.

## Implementation checklist

- [ ] Create template directories + base layout
- [ ] Implement render helper
- [ ] Mount static file server
- [ ] Smoke-render a trivial page template in a test or temp route

## Files to create / modify

| Path | Action | Notes |
| ---- | ------ | ----- |
| `web/templates/layouts/base.html` | create | base layout |
| `internal/platform/render/*.go` | create | Render helpers |
| `internal/app/routes.go` | modify | static mount |

## Technical notes

Use html/template only (auto-escaping). Never use text/template for HTML.

## Acceptance criteria

- [ ] Static file URL returns a known asset
- [ ] Render helper can execute a layout+page without panic
- [ ] Directory structure matches architecture doc

## Verification

```bash
go test ./internal/platform/render/...
go build ./...
```

## Commit & push (mandatory)

Use the commit command shape from [AGENT_GUIDE.md](../../AGENT_GUIDE.md) (single example there). Subject and body for **this** step:

**Subject:**

```text
feat(web): add template rendering and static file serving
```

**Body:**

```text
Introduce html/template layouts and /static assets as the shared
rendering platform for full pages and future HTMX fragments.

STEP-05
```

**Required actions:**

- [ ] Update `docs/implementation/STATUS.md` → `done`
- [ ] Stage this step’s files + `STATUS.md`
- [ ] Commit with the subject and body above
- [ ] `git push -u origin HEAD`
- [ ] Confirm clean / not ahead of `origin`
- [ ] Stop — do not start STEP-06

Never commit `.env` or secrets. Never `--force` push to `main`.

## Handoff to next agent

How to add a page: add `web/templates/pages/<name>.html` with `title`/`content` blocks; call `render.Render(w, status, "<name>", data)`. Static URL prefix: /static/.

After a successful push, mark this step `done` in any tracker and **stop** — do not start STEP-06.
