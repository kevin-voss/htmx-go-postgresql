# Rendering strategy

Every important view has two representations:

1. **full page** — normal browser navigation
2. **fragment** — HTMX partial request

## Detecting an HTMX 4 request

HTMX 4 may send:

```http
HX-Request-Type: partial
```

or:

```http
HX-Request-Type: full
```

It also sends `Accept: text/html`.

```go
func isPartialRequest(r *http.Request) bool {
    return r.Header.Get("HX-Request-Type") == "partial"
}
```

## Handler pattern

```go
func (app *Application) showProject(
    w http.ResponseWriter,
    r *http.Request,
) {
    data, err := app.projects.GetPageData(
        r.Context(),
        r.PathValue("workspaceSlug"),
        r.PathValue("projectSlug"),
    )
    if err != nil {
        app.handleError(w, r, err)
        return
    }

    if isPartialRequest(r) {
        app.render(w, http.StatusOK, "project-content", data)
        return
    }

    app.render(w, http.StatusOK, "project-page", data)
}
```

## Multi-target updates

Prefer HTMX 4 `<hx-partial>` when one action must update several DOM regions (comments, counts, cleared forms). See [../examples/flows/comments.md](../examples/flows/comments.md).

## Errors

Return HTML fragments for `422` / other client errors so HTMX can swap them (`hx-status:422=...`).

## Related

- HTMX decision: [../specs/htmx-decision.md](../specs/htmx-decision.md)
- HTTP conventions: [../specs/http-conventions.md](../specs/http-conventions.md)
