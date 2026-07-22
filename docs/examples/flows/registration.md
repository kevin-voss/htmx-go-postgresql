# Flow — Registration

## Diagram

```text
Landing page
    ↓
Registration form
    ↓
Validate email, display name and password
    ↓
Create user
    ↓
Send verification email through Mailpit
    ↓
Create session
    ↓
Redirect to onboarding
```

## Fields

- display name
- email
- password
- password confirmation
- acceptance of terms checkbox

## Validation

- valid email address
- normalized lowercase email
- unique email
- display name between 2 and 50 characters
- password of at least 12 characters
- matching password confirmation

## Success path

1. User submitted valid form (POST `/register`).
2. User row persisted with Argon2id password hash.
3. Verification email visible in Mailpit (`http://localhost:8025`).
4. Session cookie set (dev name `forgeboard_session`).
5. Redirect toward onboarding (see [onboarding.md](onboarding.md)).

## Failure path

- Invalid input → `422` with HTML error fragment (HTMX) or re-rendered form with errors (full page).

## Related

- Auth: [../../specs/authentication.md](../../specs/authentication.md)
- Steps: registration / mail / verification under [../../implementation/steps/](../../implementation/steps/)
