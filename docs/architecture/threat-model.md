# Threat model notes

Forgeboard is a learning/portfolio app. These notes describe the intentional
security controls and residual risks — not a formal STRIDE assessment.

## Assets

- User credentials (Argon2id password hashes)
- Server-side session tokens (hashed at rest; raw token only in HttpOnly cookie)
- Workspace data (projects, issues, comments, memberships)
- Invitation and password-reset tokens (single-use / time-limited)

## Controls

### Sessions

- Opaque session IDs stored as HttpOnly cookies (`Secure` in production)
- Tokens hashed before persistence; logout and revoke invalidate the row
- Account session list lets users revoke other devices

### CSRF

- Synchronizer token on state-changing requests (`csrf_token` form field)
- Cookie + server validation middleware on the global chain
- Safe methods (GET/HEAD/OPTIONS) skip validation

### XSS

- Go `html/template` auto-escapes interpolated values by default
- Avoid `template.HTML` / unescaped sinks for user content
- Security headers middleware sets a basic Content-Security / frame posture

### SQL injection

- Handwritten SQL with parameterized queries via `pgx` (`$1`, `$2`, …)
- No string-concatenated SQL from request input
- Repository layer owns query execution; handlers never build SQL

### Authorization

- Workspace membership required for `/w/{slug}/…` routes (fail closed → 404)
- Roles: Owner, Member, Viewer (`RequireCanMutate` / `RequireOwner`)
- Cross-workspace access denied when the caller is not a member

## Residual risks / non-goals

- No OAuth, MFA, or bot/CAPTCHA beyond login rate limiting
- Demo seed credentials are for local/dev only — never ship as production defaults
- Email is local Mailpit in development; production SMTP must be configured explicitly
- CSP and cookie flags are baseline; harden further before real internet exposure

## Related

- Auth: [../specs/authentication.md](../specs/authentication.md)
- Middleware: [middleware.md](middleware.md)
- Roles: [../specs/roles.md](../specs/roles.md)
