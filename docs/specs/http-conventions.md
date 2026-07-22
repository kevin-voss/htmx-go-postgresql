# HTTP conventions

Use normal HTTP semantics.

```text
GET     Read a page or fragment
POST    Create a resource
PATCH   Update part of a resource
DELETE  Delete or archive a resource
```

## Examples

```text
GET    /w/acme/projects/platform
POST   /w/acme/projects/platform/issues
PATCH  /w/acme/issues/42/status
PATCH  /w/acme/issues/42/assignee
POST   /w/acme/issues/42/comments
DELETE /w/acme/issues/42/comments/7
```

## Response statuses

```text
200 OK                    Successful rendered response
201 Created               Resource created
204 No Content            Successful action with no swap
303 See Other             Non-HTMX form redirect
400 Bad Request           Malformed request
401 Unauthorized          User not authenticated
403 Forbidden             User lacks permission
404 Not Found             Resource unavailable
409 Conflict              Duplicate or state conflict
422 Unprocessable Entity  Form validation failed
429 Too Many Requests     Rate limit exceeded
500 Internal Server Error Unexpected server failure
```

## HTMX note

In HTMX 4, error response HTML is swapped by default. Error responses should contain useful **HTML fragments**, not plain strings.

## Related

- Routes: [pages-and-routes.md](pages-and-routes.md)
- HTMX: [htmx-decision.md](htmx-decision.md)
- Rendering: [../architecture/rendering.md](../architecture/rendering.md)
