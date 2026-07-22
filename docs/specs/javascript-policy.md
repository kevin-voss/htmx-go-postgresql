# Minimal JavaScript policy

The application should work mainly through:

- HTML
- HTMX
- CSS
- Go-rendered responses

Custom JavaScript is allowed **only** for behavior HTMX and native HTML do not express cleanly.

## Allowed uses

- dialog open / close
- small keyboard shortcuts
- copying issue links
- preserving temporary UI preferences
- enhancing form focus after a swap

## Budget

```text
Under 200 lines of custom JavaScript
```

Vendored HTMX does **not** count toward that budget.

## Forbidden

- React
- Alpine.js
- Stimulus
- jQuery
- a frontend bundler
- TypeScript
- npm during normal development

## Related

- HTMX: [htmx-decision.md](htmx-decision.md)
- Structure: [../architecture/project-structure.md](../architecture/project-structure.md)
