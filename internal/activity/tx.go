package activity

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Tx is a commit/rollback handle shared by domain writes and activity inserts.
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Beginner starts transactions for atomic domain + activity writes.
type Beginner interface {
	Begin(ctx context.Context) (Tx, error)
}

// PoolBeginner adapts *pgxpool.Pool to Beginner.
type PoolBeginner struct {
	Pool *pgxpool.Pool
}

// Begin starts a PostgreSQL transaction.
func (b PoolBeginner) Begin(ctx context.Context) (Tx, error) {
	if b.Pool == nil {
		return nil, fmt.Errorf("activity: beginner: nil pool")
	}
	tx, err := b.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("activity: begin: %w", err)
	}
	return &pgxTx{Tx: tx}, nil
}

type pgxTx struct {
	pgx.Tx
}

func (t *pgxTx) Commit(ctx context.Context) error {
	return t.Tx.Commit(ctx)
}

func (t *pgxTx) Rollback(ctx context.Context) error {
	return t.Tx.Rollback(ctx)
}

// AsPgx returns the underlying pgx transaction when tx was started by PoolBeginner.
func AsPgx(tx Tx) (pgx.Tx, bool) {
	if t, ok := tx.(*pgxTx); ok && t != nil {
		return t.Tx, true
	}
	return nil, false
}

// RunAtomic runs domainWrite then inserts the returned event in one transaction.
// On any error the transaction is rolled back so neither side persists.
func RunAtomic(ctx context.Context, begin Beginner, store Store, domainWrite func(ctx context.Context, tx Tx) (EventInput, error)) error {
	if begin == nil {
		return fmt.Errorf("activity: run atomic: nil beginner")
	}
	if store == nil {
		return fmt.Errorf("activity: run atomic: nil store")
	}
	if domainWrite == nil {
		return fmt.Errorf("activity: run atomic: nil domain write")
	}

	tx, err := begin.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	event, err := domainWrite(ctx, tx)
	if err != nil {
		return err
	}
	if _, err := store.InsertTx(ctx, tx, event); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("activity: commit: %w", err)
	}
	return nil
}
