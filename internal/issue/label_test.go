package issue

import (
	"context"
	"errors"
	"testing"
)

func TestCreateAndAttachDetachLabel(t *testing.T) {
	t.Parallel()

	labels := newMemoryLabelStore()
	svc := NewService(newMemoryStore()).WithLabelStore(labels)
	ctx := context.Background()

	created, _, err := svc.Create(ctx, CreateInput{
		ProjectID: "proj-a",
		Title:     "Tagged",
		CreatedBy: "user-1",
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	label, fieldErrs, err := svc.CreateLabel(ctx, CreateLabelInput{
		WorkspaceID: "ws-1",
		Name:        "bug",
		Color:       "#ef4444",
	})
	if err != nil || fieldErrs.Any() {
		t.Fatalf("create label: err=%v errs=%+v", err, fieldErrs)
	}
	if label.Name != "bug" || label.Color != "#ef4444" {
		t.Fatalf("label = %+v", label)
	}

	issue, err := svc.AttachLabel(ctx, "ws-1", created.IssueNumber, label.ID)
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	if issue.ID != created.ID {
		t.Fatalf("attach returned wrong issue")
	}

	attached, err := svc.LabelsForIssue(ctx, created.ID)
	if err != nil {
		t.Fatalf("labels for issue: %v", err)
	}
	if len(attached) != 1 || attached[0].ID != label.ID {
		t.Fatalf("attached = %+v, want [%s]", attached, label.ID)
	}

	_, err = svc.DetachLabel(ctx, "ws-1", created.IssueNumber, label.ID)
	if err != nil {
		t.Fatalf("detach: %v", err)
	}
	attached, err = svc.LabelsForIssue(ctx, created.ID)
	if err != nil {
		t.Fatalf("labels after detach: %v", err)
	}
	if len(attached) != 0 {
		t.Fatalf("attached len = %d, want 0", len(attached))
	}
}

func TestCreateLabelValidation(t *testing.T) {
	t.Parallel()

	svc := NewService(newMemoryStore()).WithLabelStore(newMemoryLabelStore())
	_, errs, err := svc.CreateLabel(context.Background(), CreateLabelInput{
		WorkspaceID: "ws-1",
		Name:        "",
		Color:       "red",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if errs.Name == "" || errs.Color == "" {
		t.Fatalf("want name and color errors, got %+v", errs)
	}
}

func TestCreateLabelRejectsDuplicateName(t *testing.T) {
	t.Parallel()

	svc := NewService(newMemoryStore()).WithLabelStore(newMemoryLabelStore())
	ctx := context.Background()
	_, errs, err := svc.CreateLabel(ctx, CreateLabelInput{
		WorkspaceID: "ws-1",
		Name:        "bug",
	})
	if err != nil || errs.Any() {
		t.Fatalf("first create: err=%v errs=%+v", err, errs)
	}
	_, errs, err = svc.CreateLabel(ctx, CreateLabelInput{
		WorkspaceID: "ws-1",
		Name:        "BUG",
	})
	if err != nil {
		t.Fatalf("second create err: %v", err)
	}
	if errs.Name == "" {
		t.Fatal("want duplicate name error")
	}
}

func TestAttachLabelRejectsOtherWorkspace(t *testing.T) {
	t.Parallel()

	labels := newMemoryLabelStore()
	svc := NewService(newMemoryStore()).WithLabelStore(labels)
	ctx := context.Background()

	created, _, err := svc.Create(ctx, CreateInput{
		ProjectID: "proj-a",
		Title:     "Cross workspace",
		CreatedBy: "user-1",
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	label, _, err := svc.CreateLabel(ctx, CreateLabelInput{
		WorkspaceID: "other-ws",
		Name:        "foreign",
	})
	if err != nil {
		t.Fatalf("create label: %v", err)
	}

	_, err = svc.AttachLabel(ctx, "ws-1", created.IssueNumber, label.ID)
	if !errors.Is(err, ErrLabelNotInWorkspace) {
		t.Fatalf("err = %v, want ErrLabelNotInWorkspace", err)
	}
}

func TestValidLabelColor(t *testing.T) {
	t.Parallel()

	if !ValidLabelColor("#64748b") || !ValidLabelColor("#ABCDEF") {
		t.Fatal("expected valid hex colors")
	}
	if ValidLabelColor("#fff") || ValidLabelColor("64748b") || ValidLabelColor("#gggggg") {
		t.Fatal("expected invalid colors to fail")
	}
}
