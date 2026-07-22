package database_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kevin-voss/htmx-go-postgresql/internal/database"
)

func testDatabaseURL(t *testing.T) string {
	t.Helper()
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://forgeboard:forgeboard@localhost:5432/forgeboard?sslmode=disable"
}

func TestOpenClose(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := database.Open(ctx, testDatabaseURL(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close(pool)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestOpenRequiresURL(t *testing.T) {
	_, err := database.Open(context.Background(), "")
	if err == nil {
		t.Fatal("Open with empty URL: want error, got nil")
	}
}

func TestCloseNil(t *testing.T) {
	database.Close(nil)
}
