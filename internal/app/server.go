package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const shutdownTimeout = 10 * time.Second

func (a *Application) newServer() *http.Server {
	return &http.Server{
		Addr:              a.Config.Address,
		Handler:           a.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// Run listens on APP_ADDRESS until ctx is cancelled or SIGINT/SIGTERM,
// then shuts the HTTP server down gracefully.
func (a *Application) Run(ctx context.Context) error {
	srv := a.newServer()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		a.Logger.Info("http server listening", "address", a.Config.Address)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		a.Logger.Info("shutting down http server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return <-errCh
	}
}
