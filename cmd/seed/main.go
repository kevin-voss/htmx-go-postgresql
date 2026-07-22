// Command seed inserts a deterministic local demo account and workspace.
//
// Idempotency / reset safety:
//   - Refuses to run when APP_ENV=production.
//   - Keys demo data by demo email + workspace slug "demo".
//   - Re-running deletes only the "demo" workspace (CASCADE) and recreates
//     its project, issues, comments, labels, and activity; the demo user is
//     upserted. Other users and workspaces are left untouched.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/kevin-voss/htmx-go-postgresql/internal/activity"
	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/database"
	"github.com/kevin-voss/htmx-go-postgresql/internal/issue"
	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
)

// Well-known local/dev credentials only — never used as a production default.
const (
	demoEmail       = "demo@forgeboard.local"
	demoPassword    = "demo-password"
	demoDisplayName = "Demo User"

	demoWorkspaceName = "Demo Workspace"
	demoWorkspaceSlug = "demo"
	demoProjectName   = "Platform"
	demoProjectSlug   = "platform"
)

type seedIssue struct {
	Title       string
	Description string
	Status      string
	Priority    string
	Assign      bool
	Labels      []string
	Comments    []string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	env := strings.TrimSpace(os.Getenv("APP_ENV"))
	if env == "" {
		env = "development"
	}
	if env == "production" {
		return fmt.Errorf("refusing to seed when APP_ENV=production (dev-oriented only)")
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := database.Open(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer database.Close(pool)

	passwordHash, err := auth.Hash(demoPassword)
	if err != nil {
		return fmt.Errorf("hash demo password: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	userID, err := upsertDemoUser(ctx, tx, passwordHash)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM workspaces WHERE slug = $1`, demoWorkspaceSlug); err != nil {
		return fmt.Errorf("delete demo workspace: %w", err)
	}

	workspaceID, projectID, err := createWorkspaceAndProject(ctx, tx, userID)
	if err != nil {
		return err
	}

	labelIDs, err := createLabels(ctx, tx, workspaceID)
	if err != nil {
		return err
	}

	nIssues, nComments, nEvents, err := createIssues(ctx, tx, workspaceID, projectID, userID, labelIDs)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	fmt.Println("seed: demo data ready (idempotent for workspace slug \"demo\")")
	fmt.Printf("  email:     %s\n", demoEmail)
	fmt.Printf("  password:  %s\n", demoPassword)
	fmt.Printf("  workspace: /w/%s\n", demoWorkspaceSlug)
	fmt.Printf("  project:   /w/%s/projects/%s\n", demoWorkspaceSlug, demoProjectSlug)
	fmt.Printf("  issues:    %d  comments: %d  activity: %d\n", nIssues, nComments, nEvents)
	fmt.Println("  note:      local/dev only — re-run replaces the demo workspace only")
	return nil
}

func upsertDemoUser(ctx context.Context, tx pgx.Tx, passwordHash string) (string, error) {
	now := time.Now().UTC()
	const q = `
		INSERT INTO users (email, display_name, password_hash, email_verified_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE
		SET display_name = EXCLUDED.display_name,
		    password_hash = EXCLUDED.password_hash,
		    email_verified_at = COALESCE(users.email_verified_at, EXCLUDED.email_verified_at),
		    updated_at = now()
		RETURNING id`

	var id string
	if err := tx.QueryRow(ctx, q, demoEmail, demoDisplayName, passwordHash, now).Scan(&id); err != nil {
		return "", fmt.Errorf("upsert demo user: %w", err)
	}
	return id, nil
}

func createWorkspaceAndProject(ctx context.Context, tx pgx.Tx, userID string) (workspaceID, projectID string, err error) {
	const insertWorkspace = `
		INSERT INTO workspaces (name, slug, created_by)
		VALUES ($1, $2, $3)
		RETURNING id`
	if err := tx.QueryRow(ctx, insertWorkspace, demoWorkspaceName, demoWorkspaceSlug, userID).Scan(&workspaceID); err != nil {
		return "", "", fmt.Errorf("create workspace: %w", err)
	}

	const insertMember = `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, $3)`
	if _, err := tx.Exec(ctx, insertMember, workspaceID, userID, string(member.RoleOwner)); err != nil {
		return "", "", fmt.Errorf("create owner membership: %w", err)
	}

	const insertProject = `
		INSERT INTO projects (workspace_id, name, slug, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	if err := tx.QueryRow(ctx, insertProject, workspaceID, demoProjectName, demoProjectSlug, userID).Scan(&projectID); err != nil {
		return "", "", fmt.Errorf("create project: %w", err)
	}
	return workspaceID, projectID, nil
}

func createLabels(ctx context.Context, tx pgx.Tx, workspaceID string) (map[string]string, error) {
	defs := []struct {
		Name  string
		Color string
	}{
		{"bug", "#dc2626"},
		{"feature", "#2563eb"},
		{"docs", "#64748b"},
		{"ux", "#9333ea"},
	}

	ids := make(map[string]string, len(defs))
	const q = `
		INSERT INTO labels (workspace_id, name, color)
		VALUES ($1, $2, $3)
		RETURNING id`
	for _, d := range defs {
		var id string
		if err := tx.QueryRow(ctx, q, workspaceID, d.Name, d.Color).Scan(&id); err != nil {
			return nil, fmt.Errorf("create label %q: %w", d.Name, err)
		}
		ids[d.Name] = id
	}
	return ids, nil
}

func createIssues(
	ctx context.Context,
	tx pgx.Tx,
	workspaceID, projectID, userID string,
	labelIDs map[string]string,
) (nIssues, nComments, nEvents int, err error) {
	seeds := []seedIssue{
		{
			Title:       "Welcome reviewers to Forgeboard",
			Description: "Seeded issue for portfolio walkthroughs. Explore status, labels, and comments.",
			Status:      issue.StatusDone,
			Priority:    issue.PriorityMedium,
			Assign:      true,
			Labels:      []string{"docs"},
			Comments:    []string{"Thanks for trying the demo workspace."},
		},
		{
			Title:       "Polish landing page hero typography",
			Description: "Tighten hierarchy on the marketing page for mobile widths.",
			Status:      issue.StatusInProgress,
			Priority:    issue.PriorityHigh,
			Assign:      true,
			Labels:      []string{"ux", "feature"},
			Comments:    []string{"Started with the headline scale; next pass is CTA spacing."},
		},
		{
			Title:       "Fix CSRF error flash on expired session",
			Description: "Users who leave a form open overnight see a confusing error after submit.",
			Status:      issue.StatusTodo,
			Priority:    issue.PriorityUrgent,
			Assign:      true,
			Labels:      []string{"bug"},
			Comments:    []string{"Reproduced after a 24h idle session."},
		},
		{
			Title:       "Add empty-state illustration for new projects",
			Description: "When a project has zero issues, show a short guided empty state.",
			Status:      issue.StatusBacklog,
			Priority:    issue.PriorityLow,
			Labels:      []string{"ux", "feature"},
		},
		{
			Title:       "Document make seed credentials in README",
			Description: "README step will cover clone → make dev → make seed for reviewers.",
			Status:      issue.StatusTodo,
			Priority:    issue.PriorityMedium,
			Labels:      []string{"docs"},
			Comments:    []string{"Credentials are printed by the seed command for now."},
		},
		{
			Title:       "Archive completed onboarding checklist",
			Description: "Once the walkthrough is done, archive this issue to demonstrate archive filters.",
			Status:      issue.StatusDone,
			Priority:    issue.PriorityLow,
			Labels:      []string{"docs"},
		},
	}

	const insertIssue = `
		INSERT INTO issues (
			project_id, issue_number, title, description,
			status, priority, assignee_id, created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, '')::uuid, $8)
		RETURNING id`

	const insertLabel = `INSERT INTO issue_labels (issue_id, label_id) VALUES ($1, $2)`
	const insertComment = `
		INSERT INTO issue_comments (issue_id, author_id, body)
		VALUES ($1, $2, $3)`
	const insertEvent = `
		INSERT INTO activity_events (workspace_id, project_id, issue_id, actor_id, event_type, summary)
		VALUES ($1, $2, $3, $4, $5, $6)`

	for i, s := range seeds {
		num := i + 1
		assignee := ""
		if s.Assign {
			assignee = userID
		}

		var issueID string
		if err := tx.QueryRow(
			ctx,
			insertIssue,
			projectID,
			num,
			s.Title,
			s.Description,
			s.Status,
			s.Priority,
			assignee,
			userID,
		).Scan(&issueID); err != nil {
			return 0, 0, 0, fmt.Errorf("create issue %d: %w", num, err)
		}
		nIssues++

		if _, err := tx.Exec(
			ctx,
			insertEvent,
			workspaceID,
			projectID,
			issueID,
			userID,
			activity.TypeIssueCreated,
			`Created issue "`+s.Title+`"`,
		); err != nil {
			return 0, 0, 0, fmt.Errorf("activity issue.created %d: %w", num, err)
		}
		nEvents++

		if s.Status != issue.StatusBacklog {
			if _, err := tx.Exec(
				ctx,
				insertEvent,
				workspaceID,
				projectID,
				issueID,
				userID,
				activity.TypeIssueStatusChanged,
				"Changed status to "+issue.StatusLabel(s.Status)+` on "`+s.Title+`"`,
			); err != nil {
				return 0, 0, 0, fmt.Errorf("activity status_changed %d: %w", num, err)
			}
			nEvents++
		}

		for _, name := range s.Labels {
			labelID, ok := labelIDs[name]
			if !ok {
				return 0, 0, 0, fmt.Errorf("unknown label %q", name)
			}
			if _, err := tx.Exec(ctx, insertLabel, issueID, labelID); err != nil {
				return 0, 0, 0, fmt.Errorf("label issue %d: %w", num, err)
			}
		}

		for _, body := range s.Comments {
			if _, err := tx.Exec(ctx, insertComment, issueID, userID, body); err != nil {
				return 0, 0, 0, fmt.Errorf("comment issue %d: %w", num, err)
			}
			nComments++

			summary := "Commented: " + body
			if len(body) > 80 {
				summary = "Commented: " + body[:77] + "..."
			}
			if _, err := tx.Exec(
				ctx,
				insertEvent,
				workspaceID,
				projectID,
				issueID,
				userID,
				activity.TypeCommentCreated,
				summary,
			); err != nil {
				return 0, 0, 0, fmt.Errorf("activity comment %d: %w", num, err)
			}
			nEvents++
		}
	}

	return nIssues, nComments, nEvents, nil
}
