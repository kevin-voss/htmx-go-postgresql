package comment_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/comment"
	"github.com/kevin-voss/htmx-go-postgresql/internal/issue"
	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/render"
	"github.com/kevin-voss/htmx-go-postgresql/internal/project"
	"github.com/kevin-voss/htmx-go-postgresql/web"
)

type authzMemberStore struct {
	access map[string]member.Access
}

func (s *authzMemberStore) Create(context.Context, string, string, member.Role) (member.Membership, error) {
	return member.Membership{}, nil
}

func (s *authzMemberStore) GetByWorkspaceAndUser(_ context.Context, workspaceID, userID string) (member.Membership, error) {
	for _, a := range s.access {
		if a.WorkspaceID == workspaceID && a.Membership.UserID == userID {
			return a.Membership, nil
		}
	}
	return member.Membership{}, member.ErrNotFound
}

func (s *authzMemberStore) GetAccessBySlug(_ context.Context, slug, userID string) (member.Access, error) {
	a, ok := s.access[slug+"|"+userID]
	if !ok {
		return member.Access{}, member.ErrNotFound
	}
	return a, nil
}

func (s *authzMemberStore) HasAny(context.Context, string) (bool, error) { return true, nil }

func (s *authzMemberStore) ListByUser(context.Context, string) ([]member.UserWorkspace, error) {
	return nil, nil
}

func (s *authzMemberStore) ListByWorkspace(context.Context, string) ([]member.MemberView, error) {
	return nil, nil
}

func (s *authzMemberStore) UpdateRole(context.Context, string, string, member.Role) error {
	return nil
}

func (s *authzMemberStore) Delete(context.Context, string, string) error { return nil }

func (s *authzMemberStore) CreateInvitation(context.Context, string, string, member.Role, string, string, time.Time) (member.Invitation, error) {
	return member.Invitation{}, nil
}

func (s *authzMemberStore) GetInvitationByTokenHash(context.Context, string) (member.Invitation, error) {
	return member.Invitation{}, member.ErrNotFound
}

func (s *authzMemberStore) AcceptInvitation(context.Context, string, string, string, member.Role, time.Time) (member.Membership, error) {
	return member.Membership{}, nil
}

func (s *authzMemberStore) MarkInvitationAccepted(context.Context, string, time.Time) error {
	return nil
}

type memoryCommentStore struct {
	byID    map[string]comment.Comment
	byIssue map[string][]string
	next    int
}

func newMemoryCommentStore() *memoryCommentStore {
	return &memoryCommentStore{
		byID:    map[string]comment.Comment{},
		byIssue: map[string][]string{},
	}
}

func (s *memoryCommentStore) Create(_ context.Context, issueID, authorID, body string) (comment.Comment, error) {
	s.next++
	id := "c" + strconv.Itoa(s.next)
	c := comment.Comment{
		ID:         id,
		IssueID:    issueID,
		AuthorID:   authorID,
		AuthorName: "Author",
		Body:       body,
		CreatedAt:  time.Unix(0, 0).UTC(),
	}
	s.byID[id] = c
	s.byIssue[issueID] = append(s.byIssue[issueID], id)
	return c, nil
}

func (s *memoryCommentStore) ListByIssue(_ context.Context, issueID string) ([]comment.Comment, error) {
	ids := s.byIssue[issueID]
	out := make([]comment.Comment, 0, len(ids))
	for _, id := range ids {
		out = append(out, s.byID[id])
	}
	return out, nil
}

func (s *memoryCommentStore) CountByIssue(_ context.Context, issueID string) (int, error) {
	return len(s.byIssue[issueID]), nil
}

func (s *memoryCommentStore) GetByID(_ context.Context, id string) (comment.Comment, error) {
	c, ok := s.byID[id]
	if !ok {
		return comment.Comment{}, comment.ErrNotFound
	}
	return c, nil
}

func (s *memoryCommentStore) Delete(_ context.Context, id string) error {
	c, ok := s.byID[id]
	if !ok {
		return comment.ErrNotFound
	}
	delete(s.byID, id)
	ids := s.byIssue[c.IssueID]
	filtered := ids[:0]
	for _, existing := range ids {
		if existing != id {
			filtered = append(filtered, existing)
		}
	}
	s.byIssue[c.IssueID] = filtered
	return nil
}

type memoryIssueStore struct {
	issues map[string]issue.Issue // key: workspaceID|number
}

func (s *memoryIssueStore) Create(context.Context, string, string, string, string) (issue.Issue, error) {
	return issue.Issue{}, nil
}

func (s *memoryIssueStore) ListByProject(context.Context, string, issue.ListFilter) ([]issue.Issue, error) {
	return nil, nil
}

func (s *memoryIssueStore) GetByProjectAndNumber(context.Context, string, int) (issue.Issue, error) {
	return issue.Issue{}, issue.ErrNotFound
}

func (s *memoryIssueStore) GetByWorkspaceAndNumber(_ context.Context, workspaceID string, issueNumber int) (issue.Issue, error) {
	key := workspaceID + "|" + itoa(issueNumber)
	iss, ok := s.issues[key]
	if !ok {
		return issue.Issue{}, issue.ErrNotFound
	}
	return iss, nil
}

func (s *memoryIssueStore) UpdateStatus(context.Context, string, string) (issue.Issue, error) {
	return issue.Issue{}, nil
}

func (s *memoryIssueStore) UpdatePriority(context.Context, string, string) (issue.Issue, error) {
	return issue.Issue{}, nil
}

func (s *memoryIssueStore) UpdateAssignee(context.Context, string, string) (issue.Issue, error) {
	return issue.Issue{}, nil
}

func (s *memoryIssueStore) Archive(context.Context, string) (issue.Issue, error) {
	return issue.Issue{}, nil
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

func TestViewerCannotCreateComment(t *testing.T) {
	t.Parallel()

	memberSvc := member.NewService(&authzMemberStore{
		access: map[string]member.Access{
			"acme|viewer1": {
				WorkspaceID:   "w1",
				WorkspaceName: "Acme",
				WorkspaceSlug: "acme",
				Membership: member.Membership{
					ID:          "m1",
					WorkspaceID: "w1",
					UserID:      "viewer1",
					Role:        member.RoleViewer,
				},
			},
		},
	})

	issueSvc := issue.NewService(&memoryIssueStore{
		issues: map[string]issue.Issue{
			"w1|1": {
				ID:          "issue-1",
				ProjectID:   "proj-a",
				IssueNumber: 1,
				Title:       "Demo",
			},
		},
	})

	h := comment.NewHandler(
		comment.NewService(newMemoryCommentStore()),
		issueSvc,
		project.NewService(nil),
		memberSvc,
		nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	mux := http.NewServeMux()
	h.Mount(mux)

	req := httptest.NewRequest(http.MethodPost, "/w/acme/issues/1/comments", strings.NewReader("body=Hello"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{
		ID:          "viewer1",
		Email:       "viewer@example.com",
		DisplayName: "Viewer",
	}))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestMemberCanCreateCommentPartial(t *testing.T) {
	t.Parallel()

	memberSvc := member.NewService(&authzMemberStore{
		access: map[string]member.Access{
			"acme|member1": {
				WorkspaceID:   "w1",
				WorkspaceName: "Acme",
				WorkspaceSlug: "acme",
				Membership: member.Membership{
					ID:          "m2",
					WorkspaceID: "w1",
					UserID:      "member1",
					Role:        member.RoleMember,
				},
			},
		},
	})

	issueSvc := issue.NewService(&memoryIssueStore{
		issues: map[string]issue.Issue{
			"w1|1": {
				ID:          "issue-1",
				ProjectID:   "proj-a",
				IssueNumber: 1,
				Title:       "Demo",
			},
		},
	})

	renderer, err := render.New(web.Templates)
	if err != nil {
		t.Fatalf("render.New: %v", err)
	}

	h := comment.NewHandler(
		comment.NewService(newMemoryCommentStore()),
		issueSvc,
		project.NewService(nil),
		memberSvc,
		renderer,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	mux := http.NewServeMux()
	h.Mount(mux)

	req := httptest.NewRequest(http.MethodPost, "/w/acme/issues/1/comments", strings.NewReader("body=Hello+world"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request-Type", "partial")
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{
		ID:          "member1",
		Email:       "member@example.com",
		DisplayName: "Member",
	}))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	body := rr.Body.String()
	for _, want := range []string{
		`<hx-partial hx-target="#comment-list"`,
		`<hx-partial hx-target="#comment-count"`,
		`<hx-partial hx-target="#comment-form"`,
		"Hello world",
		"1 comment",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q; body=%s", want, body)
		}
	}
}
