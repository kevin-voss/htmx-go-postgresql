package activity_test

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/kevin-voss/htmx-go-postgresql/internal/activity"
)

// memWorld is an in-memory beginner + store that stages writes until Commit.
type memWorld struct {
	mu      sync.Mutex
	issues  []string
	events  []activity.Event
	seq     int
	failIns bool
}

type memTx struct {
	world         *memWorld
	pendingIssues []string
	pendingEvents []activity.Event
	committed     bool
	rolledBack    bool
}

func (w *memWorld) Begin(context.Context) (activity.Tx, error) {
	return &memTx{world: w}, nil
}

func (t *memTx) Commit(context.Context) error {
	t.world.mu.Lock()
	defer t.world.mu.Unlock()
	t.world.issues = append(t.world.issues, t.pendingIssues...)
	t.world.events = append(t.world.events, t.pendingEvents...)
	t.committed = true
	return nil
}

func (t *memTx) Rollback(context.Context) error {
	t.pendingIssues = nil
	t.pendingEvents = nil
	t.rolledBack = true
	return nil
}

func (w *memWorld) Insert(ctx context.Context, e activity.EventInput) (activity.Event, error) {
	return w.InsertTx(ctx, nil, e)
}

func (w *memWorld) InsertTx(_ context.Context, tx activity.Tx, e activity.EventInput) (activity.Event, error) {
	if w.failIns {
		return activity.Event{}, errors.New("activity insert failed")
	}
	w.mu.Lock()
	w.seq++
	id := "evt-" + strconv.Itoa(w.seq)
	w.mu.Unlock()

	event := activity.Event{
		ID:          id,
		WorkspaceID: e.WorkspaceID,
		ProjectID:   e.ProjectID,
		IssueID:     e.IssueID,
		ActorID:     e.ActorID,
		Type:        e.Type,
		Summary:     e.Summary,
	}

	if tx == nil {
		w.mu.Lock()
		defer w.mu.Unlock()
		w.events = append(w.events, event)
		return event, nil
	}
	mt, ok := tx.(*memTx)
	if !ok {
		return activity.Event{}, errors.New("unsupported tx")
	}
	mt.pendingEvents = append(mt.pendingEvents, event)
	return event, nil
}

func (w *memWorld) ListByProject(_ context.Context, projectID string, _ int) ([]activity.Event, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	var out []activity.Event
	for _, e := range w.events {
		if e.ProjectID == projectID {
			out = append(out, e)
		}
	}
	if out == nil {
		out = []activity.Event{}
	}
	return out, nil
}

func (w *memWorld) ListByWorkspace(_ context.Context, workspaceID string, _ int) ([]activity.Event, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	var out []activity.Event
	for _, e := range w.events {
		if e.WorkspaceID == workspaceID {
			out = append(out, e)
		}
	}
	if out == nil {
		out = []activity.Event{}
	}
	return out, nil
}

func stageIssue(tx activity.Tx, title string) error {
	mt, ok := tx.(*memTx)
	if !ok {
		return errors.New("unsupported tx")
	}
	mt.pendingIssues = append(mt.pendingIssues, title)
	return nil
}

func TestRecordAtomicCommitsDomainAndActivityTogether(t *testing.T) {
	t.Parallel()

	world := &memWorld{}
	svc := activity.NewService(world)
	ctx := context.Background()

	err := svc.RecordAtomic(ctx, world, func(ctx context.Context, tx activity.Tx) (activity.EventInput, error) {
		if err := stageIssue(tx, "Ship login"); err != nil {
			return activity.EventInput{}, err
		}
		return activity.EventInput{
			WorkspaceID: "ws-1",
			ProjectID:   "proj-1",
			IssueID:     "issue-1",
			ActorID:     "user-1",
			Type:        activity.TypeIssueCreated,
			Summary:     "Created issue Ship login",
		}, nil
	})
	if err != nil {
		t.Fatalf("RecordAtomic: %v", err)
	}

	if len(world.issues) != 1 || world.issues[0] != "Ship login" {
		t.Fatalf("issues after commit = %#v", world.issues)
	}
	if len(world.events) != 1 || world.events[0].Type != activity.TypeIssueCreated {
		t.Fatalf("events after commit = %#v", world.events)
	}
}

func TestRecordAtomicRollsBackDomainWhenActivityFails(t *testing.T) {
	t.Parallel()

	world := &memWorld{failIns: true}
	svc := activity.NewService(world)
	ctx := context.Background()

	err := svc.RecordAtomic(ctx, world, func(ctx context.Context, tx activity.Tx) (activity.EventInput, error) {
		if err := stageIssue(tx, "Should not persist"); err != nil {
			return activity.EventInput{}, err
		}
		return activity.EventInput{
			WorkspaceID: "ws-1",
			ProjectID:   "proj-1",
			IssueID:     "issue-1",
			ActorID:     "user-1",
			Type:        activity.TypeIssueCreated,
			Summary:     "Created issue",
		}, nil
	})
	if err == nil {
		t.Fatal("expected activity insert failure")
	}

	if len(world.issues) != 0 {
		t.Fatalf("domain row leaked after rollback: %#v", world.issues)
	}
	if len(world.events) != 0 {
		t.Fatalf("activity row leaked after rollback: %#v", world.events)
	}
}

func TestRecordAtomicRollsBackActivityWhenDomainFails(t *testing.T) {
	t.Parallel()

	world := &memWorld{}
	svc := activity.NewService(world)
	ctx := context.Background()
	boom := errors.New("domain write failed")

	err := svc.RecordAtomic(ctx, world, func(ctx context.Context, tx activity.Tx) (activity.EventInput, error) {
		if err := stageIssue(tx, "partial"); err != nil {
			return activity.EventInput{}, err
		}
		return activity.EventInput{}, boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want domain failure", err)
	}
	if len(world.issues) != 0 || len(world.events) != 0 {
		t.Fatalf("leaked state issues=%#v events=%#v", world.issues, world.events)
	}
}

func TestListByProjectScopesEvents(t *testing.T) {
	t.Parallel()

	world := &memWorld{}
	svc := activity.NewService(world)
	ctx := context.Background()

	if _, err := svc.Record(ctx, activity.EventInput{
		WorkspaceID: "ws-1",
		ProjectID:   "proj-a",
		ActorID:     "user-1",
		Type:        activity.TypeIssueCreated,
		Summary:     "A",
	}); err != nil {
		t.Fatalf("record a: %v", err)
	}
	if _, err := svc.Record(ctx, activity.EventInput{
		WorkspaceID: "ws-1",
		ProjectID:   "proj-b",
		ActorID:     "user-1",
		Type:        activity.TypeCommentCreated,
		Summary:     "B",
	}); err != nil {
		t.Fatalf("record b: %v", err)
	}

	got, err := svc.ListByProject(ctx, "proj-a", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 || got[0].Summary != "A" {
		t.Fatalf("got %#v", got)
	}
}
