# Flow — Workspace invitation

## Diagram

```text
Owner enters email
    ↓
Create invitation token
    ↓
Send invitation email
    ↓
User opens invitation link
    ↓
Existing user logs in
or
New user registers
    ↓
Accept invitation
    ↓
Membership is created
```

## Notes

- Public route: `GET /invites/{token}`
- Email delivery in development via Mailpit.
- Accepting creates `workspace_members` with the invited role (default **Member** unless specified).
- Owner/admin permissions for inviting: see [../../specs/roles.md](../../specs/roles.md).

## Failure cases

- Expired / unknown token → clear error page.
- Already a member → idempotent success or friendly message (pick one; document in implementation).
- Insufficient role to invite → `403`.

## Related

- Roles: [../../specs/roles.md](../../specs/roles.md)
- Database: [../../architecture/database.md](../../architecture/database.md)
