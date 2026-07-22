# Product specification — Forgeboard

## Problem

Small software teams often need a lightweight place to:

- create projects
- track issues
- assign work
- discuss tasks
- monitor progress

Existing products can be too complex for small projects. Forgeboard offers a deliberately simple workflow.

## Target users

- small software teams
- startup teams
- freelance developers
- technical project owners
- individual developers managing side projects

## Core product promise

A user should be able to:

1. create an account
2. create a workspace
3. create a project
4. add issues
5. assign and prioritize issues
6. move issues through a workflow
7. collaborate through comments
8. see recent project activity

## Positioning

Think of it as:

> A compact combination of Linear, GitHub Issues and a simple team activity feed.

It demonstrates Go fundamentals, `net/http`, SSR, HTMX 4, PostgreSQL, authz, Docker, testing, and secure web development. It is **not** intended to compete with Jira.

## Scope — included

- account registration
- login and logout
- email verification
- password reset
- secure sessions
- workspaces
- workspace membership
- workspace invitations
- projects
- issues
- issue statuses
- issue priorities
- assignments
- comments
- labels
- search and filtering
- activity history
- user preferences
- responsive interface
- light and dark color scheme
- Docker-based local environment

## Scope — excluded

- payments / subscriptions
- file uploads
- OAuth / social login
- mobile apps
- complex notifications
- WebSockets
- Kubernetes
- microservices
- rich-text editor
- time tracking
- roadmaps
- sprint planning
- external integrations

## Related

- Roles: [roles.md](roles.md)
- Flows: [../examples/flows/](../examples/flows/)
- Done criteria: [../DEFINITION_OF_DONE.md](../DEFINITION_OF_DONE.md)
