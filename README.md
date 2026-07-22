# Forgeboard

Issue tracker demo: HTMX + Go + PostgreSQL.

## Quick start

```bash
git clone <repo-url>
cd htmx-go-postgresql
make dev
```

Then open:

| Service      | URL                       |
| ------------ | ------------------------- |
| Application  | http://localhost:8080     |
| Mailpit UI   | http://localhost:8025     |
| PostgreSQL   | localhost:5432            |

`make dev` builds the app image, starts Postgres and Mailpit, runs migrations, and serves the web app. Use `make help` for other targets.
