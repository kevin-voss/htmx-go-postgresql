package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/issue"
)

func TestAuthRepositoryPersistsUser(t *testing.T) {
	pool := Pool(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	suffix := uniqueSuffix(t)
	email := "repo-" + suffix + "@example.com"
	hash, err := auth.Hash("repository-password")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	repo := auth.NewRepository(pool)
	user, err := repo.Create(ctx, email, "Repo User", hash)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM users WHERE id = $1`, user.ID)
	})

	got, err := repo.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if got.ID != user.ID {
		t.Fatalf("GetByEmail id = %q, want %q", got.ID, user.ID)
	}
	if got.PasswordHash != hash {
		t.Fatal("GetByEmail password hash mismatch")
	}
}

func TestIssueRepositoryPersistsAgainstPostgres(t *testing.T) {
	pool := Pool(t)
	fx := SeedWorkspace(t, pool)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	repo := issue.NewRepository(pool)
	created, err := repo.Create(ctx, fx.ProjectID, "Integration issue", "persisted in Postgres", fx.UserID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.IssueNumber != 1 {
		t.Fatalf("IssueNumber = %d, want 1", created.IssueNumber)
	}

	got, err := repo.GetByProjectAndNumber(ctx, fx.ProjectID, created.IssueNumber)
	if err != nil {
		t.Fatalf("GetByProjectAndNumber: %v", err)
	}
	if got.ID != created.ID || got.Title != "Integration issue" {
		t.Fatalf("unexpected issue: %+v", got)
	}

	byWorkspace, err := repo.GetByWorkspaceAndNumber(ctx, fx.WorkspaceID, created.IssueNumber)
	if err != nil {
		t.Fatalf("GetByWorkspaceAndNumber: %v", err)
	}
	if byWorkspace.ID != created.ID {
		t.Fatalf("workspace lookup id = %q, want %q", byWorkspace.ID, created.ID)
	}
}

func TestIssueRepositoryIsolatesWorkspaces(t *testing.T) {
	pool := Pool(t)
	owner := SeedWorkspace(t, pool)
	outsider := SeedWorkspace(t, pool)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	repo := issue.NewRepository(pool)
	created, err := repo.Create(ctx, owner.ProjectID, "Private to owner", "", owner.UserID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = repo.GetByWorkspaceAndNumber(ctx, outsider.WorkspaceID, created.IssueNumber)
	if !errors.Is(err, issue.ErrNotFound) {
		t.Fatalf("cross-workspace GetByWorkspaceAndNumber err = %v, want ErrNotFound", err)
	}
}
