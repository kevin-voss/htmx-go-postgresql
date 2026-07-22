# HTMX decision — use HTMX 4 (pinned)

**Decision:** use **HTMX 4**, pinned to `4.0.0-beta5` (as of 2026-07-22 still described as “under construction”). Suitable for a learning/portfolio project; pin the exact file so a future beta cannot break the build unexpectedly.

Vendor the file into the repo (no npm, no CDN dependency for the app itself):

```html
<script src="/static/vendor/htmx-4.0.0-beta5.min.js"></script>
```

HTMX requires no build step.

Sources: [htmx four](https://four.htmx.org/), [htmx 4 docs](https://four.htmx.org/docs).

---

## HTMX 2 vs HTMX 4 (project-relevant)

| Area | HTMX 2 | HTMX 4 |
| ---- | ------ | ------ |
| Request implementation | `XMLHttpRequest` | Native `fetch()` |
| Attribute inheritance | Implicit by default | Explicit using `:inherited` |
| Error responses | `4xx`/`5xx` generally not swapped | Error HTML swapped by default |
| History cache | May use `localStorage` | Re-fetches pages instead |
| Request timeout | No timeout by default | 60 seconds by default |
| Event naming | `htmx:afterSwap` | `htmx:after:swap` |
| Extensions | Enabled with `hx-ext` | Script inclusion activates them |
| Multiple updates | Mainly out-of-band swaps | `<hx-partial>` support |
| Swap options | Traditional replacements | Morphing and `textContent` |
| Status handling | Custom event logic | `hx-status:422`, `hx-status:5xx`, etc. |
| JavaScript helpers | More HTMX utilities | Prefers native DOM APIs |

Other relevant changes: `hx-delete` no longer auto-includes enclosing form data; request queuing moves to `hx-sync`; several attributes/events renamed; main response swaps before out-of-band elements.

---

## Why HTMX 4 fits Forgeboard

Greenfield project — learn newer concepts directly:

- explicit inheritance
- error fragments via HTTP status codes
- `fetch()`-based requests
- `<hx-partial>` for multi-target updates
- `hx-status` for validation errors
- morph swaps
- cleaner event names
- native JS instead of HTMX utility wrappers

### Validation error pattern

```html
<form
    hx-post="/projects"
    hx-target="#project-list"
    hx-swap="beforeend"
    hx-status:422="target:#project-form-errors swap:innerHTML"
>
    ...
</form>
```

Server returns:

```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: text/html
```

Body is an HTML error fragment; HTMX 4 swaps it without custom JS.

## Related

- Rendering strategy: [../architecture/rendering.md](../architecture/rendering.md)
- Comment multi-partial example: [../examples/flows/comments.md](../examples/flows/comments.md)
- JS policy: [javascript-policy.md](javascript-policy.md)
