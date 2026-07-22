package comment_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kevin-voss/htmx-go-postgresql/internal/comment"
	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
)

func TestValidateCreate(t *testing.T) {
	t.Parallel()

	if errs := comment.ValidateCreate(comment.CreateInput{Body: ""}); errs.Body == "" {
		t.Fatal("expected body required error")
	}
	if errs := comment.ValidateCreate(comment.CreateInput{Body: strings.Repeat("x", 10001)}); errs.Body == "" {
		t.Fatal("expected body too long error")
	}
	if errs := comment.ValidateCreate(comment.CreateInput{Body: "ok"}); errs.Any() {
		t.Fatalf("unexpected errors: %+v", errs)
	}
}

func TestDeleteRequiresAuthorOrOwner(t *testing.T) {
	t.Parallel()

	store := newMemoryCommentStore()
	svc := comment.NewService(store)
	created, _, err := svc.Create(context.Background(), comment.CreateInput{
		IssueID:  "issue-1",
		AuthorID: "author1",
		Body:     "Hello",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	err = svc.Delete(context.Background(), created.ID, "other", string(member.RoleMember))
	if !errors.Is(err, comment.ErrForbidden) {
		t.Fatalf("member non-author delete err = %v, want ErrForbidden", err)
	}

	err = svc.Delete(context.Background(), created.ID, "owner1", string(member.RoleOwner))
	if err != nil {
		t.Fatalf("owner delete: %v", err)
	}

	created2, _, err := svc.Create(context.Background(), comment.CreateInput{
		IssueID:  "issue-1",
		AuthorID: "author1",
		Body:     "Again",
	})
	if err != nil {
		t.Fatalf("create 2: %v", err)
	}
	if err := svc.Delete(context.Background(), created2.ID, "author1", string(member.RoleMember)); err != nil {
		t.Fatalf("author delete: %v", err)
	}
}
