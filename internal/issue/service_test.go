package issue

import (
	"context"
	"strings"
	"testing"
)

func TestValidateCreateRequiresTitle(t *testing.T) {
	t.Parallel()

	errs := ValidateCreate(CreateInput{Title: ""})
	if errs.Title == "" {
		t.Fatal("want title error for empty title")
	}

	errs = ValidateCreate(CreateInput{Title: "Ship it"})
	if errs.Any() {
		t.Fatalf("unexpected errors: %+v", errs)
	}
}

func TestCreateRejectsEmptyTitle(t *testing.T) {
	t.Parallel()

	svc := NewService(newMemoryStore())
	_, errs, err := svc.Create(context.Background(), CreateInput{
		ProjectID: "proj-a",
		Title:     "   ",
		CreatedBy: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errs.Title == "" {
		t.Fatal("want title validation error")
	}
}

func TestCreateRejectsTooLongTitle(t *testing.T) {
	t.Parallel()

	svc := NewService(newMemoryStore())
	_, errs, err := svc.Create(context.Background(), CreateInput{
		ProjectID: "proj-a",
		Title:     strings.Repeat("a", 201),
		CreatedBy: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errs.Title == "" {
		t.Fatal("want title length error")
	}
}

func TestGetByProjectAndNumber(t *testing.T) {
	t.Parallel()

	svc := NewService(newMemoryStore())
	created, _, err := svc.Create(context.Background(), CreateInput{
		ProjectID: "proj-a",
		Title:     "Lookup me",
		CreatedBy: "user-1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := svc.GetByProjectAndNumber(context.Background(), "proj-a", created.IssueNumber)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != created.ID || got.Title != "Lookup me" {
		t.Fatalf("got %+v", got)
	}

	_, err = svc.GetByProjectAndNumber(context.Background(), "proj-b", created.IssueNumber)
	if err != ErrNotFound {
		t.Fatalf("cross-project get err = %v, want ErrNotFound", err)
	}
}

func TestDisplayKeyAndStatusLabel(t *testing.T) {
	t.Parallel()

	if got := DisplayKey("platform", 3); got != "PLATFORM-3" {
		t.Fatalf("DisplayKey = %q, want PLATFORM-3", got)
	}
	if got := StatusLabel(StatusBacklog); got != "Backlog" {
		t.Fatalf("StatusLabel = %q, want Backlog", got)
	}
}
