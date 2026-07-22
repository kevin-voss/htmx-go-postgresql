package issue

import (
	"context"
	"errors"
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
	if got := PriorityLabel(PriorityUrgent); got != "Urgent" {
		t.Fatalf("PriorityLabel = %q, want Urgent", got)
	}
}

func TestUpdateStatusAndPriority(t *testing.T) {
	t.Parallel()

	svc := NewService(newMemoryStore())
	created, _, err := svc.Create(context.Background(), CreateInput{
		ProjectID: "proj-a",
		Title:     "Workflow",
		CreatedBy: "user-1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := svc.UpdateStatus(context.Background(), "ws-1", created.IssueNumber, StatusInProgress)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if updated.Status != StatusInProgress {
		t.Fatalf("status = %q, want %q", updated.Status, StatusInProgress)
	}

	updated, err = svc.UpdatePriority(context.Background(), "ws-1", created.IssueNumber, PriorityHigh)
	if err != nil {
		t.Fatalf("update priority: %v", err)
	}
	if updated.Priority != PriorityHigh {
		t.Fatalf("priority = %q, want %q", updated.Priority, PriorityHigh)
	}

	_, err = svc.UpdateStatus(context.Background(), "ws-1", created.IssueNumber, "nope")
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("invalid status err = %v, want ErrInvalidStatus", err)
	}
	_, err = svc.UpdatePriority(context.Background(), "ws-1", created.IssueNumber, "nope")
	if !errors.Is(err, ErrInvalidPriority) {
		t.Fatalf("invalid priority err = %v, want ErrInvalidPriority", err)
	}
}

func TestUpdateAssigneeRequiresWorkspaceMember(t *testing.T) {
	t.Parallel()

	svc := NewService(newMemoryStore()).WithMembershipChecker(stubMembers{
		members: map[string]bool{"member-1": true},
	})
	created, _, err := svc.Create(context.Background(), CreateInput{
		ProjectID: "proj-a",
		Title:     "Assign me",
		CreatedBy: "user-1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := svc.UpdateAssignee(context.Background(), "ws-1", created.IssueNumber, "member-1")
	if err != nil {
		t.Fatalf("assign member: %v", err)
	}
	if updated.AssigneeID != "member-1" {
		t.Fatalf("assignee = %q, want member-1", updated.AssigneeID)
	}

	_, err = svc.UpdateAssignee(context.Background(), "ws-1", created.IssueNumber, "outsider")
	if !errors.Is(err, ErrInvalidAssignee) {
		t.Fatalf("outsider err = %v, want ErrInvalidAssignee", err)
	}

	cleared, err := svc.UpdateAssignee(context.Background(), "ws-1", created.IssueNumber, "")
	if err != nil {
		t.Fatalf("clear assignee: %v", err)
	}
	if cleared.AssigneeID != "" {
		t.Fatalf("assignee = %q, want empty", cleared.AssigneeID)
	}
}

func TestArchiveHidesFromDefaultList(t *testing.T) {
	t.Parallel()

	svc := NewService(newMemoryStore())
	created, _, err := svc.Create(context.Background(), CreateInput{
		ProjectID: "proj-a",
		Title:     "Archive me",
		CreatedBy: "user-1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	archived, err := svc.Archive(context.Background(), "ws-1", created.IssueNumber)
	if err != nil {
		t.Fatalf("archive: %v", err)
	}
	if !archived.Archived {
		t.Fatal("want archived=true")
	}

	listed, err := svc.ListByProject(context.Background(), "proj-a", ListFilter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("list len = %d, want 0 (archived hidden)", len(listed))
	}

	got, err := svc.GetByProjectAndNumber(context.Background(), "proj-a", created.IssueNumber)
	if err != nil {
		t.Fatalf("get archived: %v", err)
	}
	if !got.Archived {
		t.Fatal("direct get should still return archived issue")
	}
}
