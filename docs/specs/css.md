# CSS specification

No Tailwind, Sass, PostCSS, or CSS build step. Use plain modern CSS.

## Browser strategy

Target current evergreen browsers: Chrome, Edge, Firefox, Safari. Learning project — modern features OK; no legacy browser support.

## Features to use

- native CSS nesting
- cascade layers
- custom properties
- `color-mix()`, `oklch()`, `light-dark()`
- container queries
- logical properties
- `:has()`, `:is()`, `:where()`
- `clamp()`
- CSS grid / subgrid where useful
- view transitions
- `prefers-color-scheme`, `prefers-reduced-motion`
- individual transform properties

## Cascade layers

```css
@layer reset, tokens, base, layout, components, utilities;
```

## Token foundation (example)

```css
@layer tokens {
    :root {
        color-scheme: light dark;

        --color-background: light-dark(
            oklch(98% 0.005 250),
            oklch(18% 0.015 250)
        );

        --color-surface: light-dark(
            oklch(100% 0 0),
            oklch(23% 0.015 250)
        );

        --color-text: light-dark(
            oklch(22% 0.015 250),
            oklch(94% 0.005 250)
        );

        --color-primary: oklch(60% 0.18 255);
        --color-border: color-mix(
            in oklch,
            var(--color-text) 15%,
            transparent
        );

        --space-1: 0.25rem;
        --space-2: 0.5rem;
        --space-3: 0.75rem;
        --space-4: 1rem;
        --space-6: 1.5rem;
        --space-8: 2rem;

        --radius-small: 0.4rem;
        --radius-medium: 0.7rem;
        --radius-large: 1rem;
    }
}
```

## Component naming

Predictable class names (light BEM-ish, stay readable):

```text
.button
.button--primary
.button--danger

.form-field
.form-field__label
.form-field__input
.form-field__error

.issue-card
.issue-card__header
.issue-card__title
.issue-card__meta
```

## Suggested file layout

```text
web/static/css/
├── reset.css
├── tokens.css
├── base.css
├── layout.css
├── components.css
├── utilities.css
└── pages/
    ├── auth.css
    ├── project.css
    └── issue.css
```

## Related

- Project structure: [../architecture/project-structure.md](../architecture/project-structure.md)
- JS policy: [javascript-policy.md](javascript-policy.md)
