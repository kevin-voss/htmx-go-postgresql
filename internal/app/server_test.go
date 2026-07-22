package app_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/kevin-voss/htmx-go-postgresql/internal/app"
	"github.com/kevin-voss/htmx-go-postgresql/internal/config"
)

func TestRunGracefulShutdown(t *testing.T) {
	application := app.New(
		config.Config{Env: "test", Address: "127.0.0.1:0"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- application.Run(ctx)
	}()

	// Give ListenAndServe a moment to bind before requesting shutdown.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for graceful shutdown")
	}
}
