# Flow — Login

## Diagram

```text
Login form
    ↓
Look up user by normalized email
    ↓
Verify password hash
    ↓
Generate random session token
    ↓
Store token hash in PostgreSQL
    ↓
Set secure cookie
    ↓
Redirect to workspace / app
```

## Security requirements

- Generic error message only:

```text
Invalid email or password.
```

- Never reveal whether the email exists.
- Store `sha256(rawToken)` only — never the raw token.
- Cookie flags per [../../specs/authentication.md](../../specs/authentication.md).

## Related routes

```text
GET  /login
POST /login
POST /logout
```

## Related

- Auth spec: [../../specs/authentication.md](../../specs/authentication.md)
- Sessions step(s): [../../implementation/steps/](../../implementation/steps/)
