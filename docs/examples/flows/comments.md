# Flow — Comments (multi-target HTMX)

## Diagram

```text
Open issue
    ↓
Write comment
    ↓
POST through HTMX
    ↓
Append rendered comment
    ↓
Clear textarea
    ↓
Update comment count
```

## HTMX 4 `<hx-partial>` example

```html
<hx-partial hx-target="#comment-list" hx-swap="beforeend">
    <!-- rendered new comment -->
</hx-partial>

<hx-partial hx-target="#comment-count">
    <span>4 comments</span>
</hx-partial>

<hx-partial hx-target="#comment-form">
    <!-- cleared comment form -->
</hx-partial>
```

This is the canonical **multi-target HTMX update** required by the definition of done.

## Related

- Rendering: [../../architecture/rendering.md](../../architecture/rendering.md)
- HTMX: [../../specs/htmx-decision.md](../../specs/htmx-decision.md)
- Definition of done: [../../DEFINITION_OF_DONE.md](../../DEFINITION_OF_DONE.md)
