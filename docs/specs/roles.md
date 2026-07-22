# User roles

Authorization is workspace-scoped. v1 uses three roles (simplified from a four-role draft).

## Roles (v1)

```text
Owner
Member
Viewer
```

This is enough to demonstrate RBAC without excessive complexity.

---

## Owner

Can:

- update workspace settings
- invite and remove members
- change member roles
- create and archive projects
- transfer workspace ownership
- delete the workspace

---

## Member

Can:

- view workspace projects
- create issues
- edit issues
- comment
- assign issues
- change issue status

(Admins from the fuller draft are folded into Owner capabilities where needed for v1, or treated as Member with elevated project powers — **prefer Owner / Member / Viewer only** unless a step explicitly adds Admin.)

---

## Viewer

Can:

- view projects
- view issues
- view comments

Cannot change data.

---

## Authorization principles

- Check membership **and** role before mutating workspace resources.
- Cross-workspace access must always fail closed (403 / 404 as designed).
- Invitation acceptance creates a membership with the invited role (default Member unless specified).

## Related

- Invitation flow: [../examples/flows/invitation.md](../examples/flows/invitation.md)
- Middleware: [../architecture/middleware.md](../architecture/middleware.md)
- Implementation: workspace/membership steps in [../implementation/steps/](../implementation/steps/)
