# Forgeboard documentation

Canonical documentation for **Forgeboard** — a lightweight, server-rendered issue and project tracker built with Go, HTMX 4, and PostgreSQL.

The original monolithic draft lives at [`../project-specs.md`](../project-specs.md). This `docs/` tree is the structured, agent-ready source of truth.

---

## How to use these docs

| Audience | Start here |
| -------- | ---------- |
| Human reader / portfolio reviewer | [specs/product.md](specs/product.md) → [architecture/overview.md](architecture/overview.md) |
| AI agent implementing the app | [AGENT_GUIDE.md](AGENT_GUIDE.md) → [implementation/README.md](implementation/README.md) |
| Understanding a user journey | [examples/flows/](examples/flows/) |
| Checking “are we done?” | [DEFINITION_OF_DONE.md](DEFINITION_OF_DONE.md) |

---

## Folder map

```text
docs/
├── README.md                 ← you are here
├── AGENT_GUIDE.md            ← how one agent executes one step
├── DEFINITION_OF_DONE.md     ← project completion checklist
│
├── specs/                    ← product & technical requirements
├── architecture/             ← system design & structure
├── examples/
│   └── flows/                ← end-to-end user/system flows
└── implementation/
    ├── README.md             ← step index & rules
    ├── _template.md          ← copy for new steps
    ├── milestones.md         ← milestone grouping
    └── steps/                ← ordered, atomic work units
```

---

## Reading order (recommended)

1. [specs/product.md](specs/product.md) — what we are building
2. [specs/technology-stack.md](specs/technology-stack.md) — with what
3. [architecture/overview.md](architecture/overview.md) — how it is layered
4. [examples/flows/](examples/flows/) — how users move through the product
5. [implementation/README.md](implementation/README.md) — how to build it step by step

---

## Product in one sentence

Forgeboard is a compact combination of Linear, GitHub Issues, and a simple team activity feed — intentionally small, portfolio-quality, not a Jira clone.

---

## Non-goals (global)

Do not add: payments, OAuth, file uploads, WebSockets, Kubernetes, microservices, rich-text editors, time tracking, roadmaps, sprint planning, or external integrations.

See [specs/product.md](specs/product.md) for the full include/exclude list.
