// Package integration holds Postgres-backed integration tests for Forgeboard.
//
// Run via `make test` (Compose sets DATABASE_URL and applies migrations).
package integration

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/database"
	"github.com/kevin-voss/htmx-go-postgresql/internal/project"
	"github.com/kevin-voss/htmx-go-postgresql/internal/workspace"
)

const defaultDatabaseURL = "postgres://forgeboard:forgeboard@localhost:5432/forgeboard?sslmode=disable"

var (
	poolOnce sync.Once
	shared   *pgxpool.Pool
	poolErr  error
)

// Pool returns a process-wide pgx pool connected to DATABASE_URL (or the local Compose default).
// Migrations are applied once before the first successful open.
func Pool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	poolOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		url := databaseURL()
		if err := migrateUp(url); err != nil {
			poolErr = fmt.Errorf("migrate: %w", err)
			return
		}
		shared, poolErr = database.Open(ctx, url)
	})
	if poolErr != nil {
		t.Fatalf("integration database: %v", poolErr)
	}
	return shared
}

func databaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return defaultDatabaseURL
}

func migrateUp(databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, filepath.Join(moduleRoot(), "db", "migrations"))
}

func moduleRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

// uniqueSuffix returns a short lowercase hex token safe for emails and slugs.
func uniqueSuffix(t *testing.T) string {
	t.Helper()
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return hex.EncodeToString(b[:])
}

// Fixture is a disposable user + workspace + project for integration tests.
type Fixture struct {
	t *testing.T

	UserID        string
	Email         string
	Password      string
	DisplayName   string
	WorkspaceID   string
	WorkspaceSlug string
	ProjectID     string
	ProjectSlug   string
}

// SeedWorkspace creates a user (argon2 hash), Owner workspace, and project, and registers cleanup.
func SeedWorkspace(t *testing.T, pool *pgxpool.Pool) Fixture {
	t.Helper()

	suffix := uniqueSuffix(t)
	password := "integration-password"
	hash, err := auth.Hash(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	authRepo := auth.NewRepository(pool)
	user, err := authRepo.Create(ctx, "it-"+suffix+"@example.com", "IT User", hash)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	wsSlug := "it-" + suffix
	wsRepo := workspace.NewRepository(pool)
	ws, err := wsRepo.Create(ctx, "IT Workspace", wsSlug, user.ID)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	projSlug := "proj"
	projRepo := project.NewRepository(pool)
	proj, err := projRepo.Create(ctx, ws.ID, "IT Project", projSlug, user.ID)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	f := Fixture{
		t:             t,
		UserID:        user.ID,
		Email:         user.Email,
		Password:      password,
		DisplayName:   user.DisplayName,
		WorkspaceID:   ws.ID,
		WorkspaceSlug: ws.Slug,
		ProjectID:     proj.ID,
		ProjectSlug:   proj.Slug,
	}
	t.Cleanup(func() { f.cleanup(pool) })
	return f
}

func (f Fixture) cleanup(pool *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if f.WorkspaceID != "" {
		if _, err := pool.Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, f.WorkspaceID); err != nil {
			f.t.Logf("cleanup workspace: %v", err)
		}
	}
	if f.UserID != "" {
		if _, err := pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, f.UserID); err != nil {
			f.t.Logf("cleanup user: %v", err)
		}
	}
}
