package issue_test

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
	"github.com/kevin-voss/htmx-go-postgresql/internal/issue"
	"github.com/kevin-voss/htmx-go-postgresql/internal/member"
	"github.com/kevin-voss/htmx-go-postgresql/internal/project"
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

func TestViewerCannotChangeIssueFields(t *testing.T) {
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

	issueSvc := issue.NewService(newHandlerMemoryStore()).WithMembershipChecker(memberSvc)
	_, _, err := issueSvc.Create(context.Background(), issue.CreateInput{
		ProjectID: "proj-a",
		Title:     "Locked for viewers",
		CreatedBy: "owner1",
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	h := issue.NewHandler(
		issueSvc,
		project.NewService(nil),
		memberSvc,
		nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	mux := http.NewServeMux()
	h.Mount(mux)

	cases := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/w/acme/issues/1/status", "status=todo"},
		{http.MethodPatch, "/w/acme/issues/1/status", "status=todo"},
		{http.MethodPost, "/w/acme/issues/1/priority", "priority=high"},
		{http.MethodPatch, "/w/acme/issues/1/priority", "priority=high"},
		{http.MethodPost, "/w/acme/issues/1/assignee", "assignee_id="},
		{http.MethodPatch, "/w/acme/issues/1/assignee", "assignee_id="},
		{http.MethodPost, "/w/acme/issues/1/archive", ""},
		{http.MethodPost, "/w/acme/labels", "name=bug&color=%2364748b"},
		{http.MethodPost, "/w/acme/labels/label-1/delete", ""},
		{http.MethodPost, "/w/acme/issues/1/labels", "label_id=label-1"},
		{http.MethodPost, "/w/acme/issues/1/labels/label-1/remove", ""},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
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
		})
	}
}

func TestNonMemberCannotManageLabels(t *testing.T) {
	t.Parallel()

	memberSvc := member.NewService(&authzMemberStore{access: map[string]member.Access{}})
	h := issue.NewHandler(
		issue.NewService(newHandlerMemoryStore()),
		project.NewService(nil),
		memberSvc,
		nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	mux := http.NewServeMux()
	h.Mount(mux)

	cases := []string{
		"/w/acme/labels",
		"/w/acme/labels",
	}
	methods := []string{http.MethodGet, http.MethodPost}
	for i, path := range cases {
		req := httptest.NewRequest(methods[i], path, strings.NewReader("name=bug"))
		if methods[i] == http.MethodPost {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{
			ID:          "outsider",
			Email:       "out@example.com",
			DisplayName: "Outsider",
		}))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("%s %s status = %d, want %d", methods[i], path, rr.Code, http.StatusNotFound)
		}
	}
}

func TestMemberCanChangeIssueStatus(t *testing.T) {
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

	store := newHandlerMemoryStore()
	issueSvc := issue.NewService(store).WithMembershipChecker(memberSvc)
	created, _, err := issueSvc.Create(context.Background(), issue.CreateInput{
		ProjectID: "proj-a",
		Title:     "Editable",
		CreatedBy: "member1",
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	h := issue.NewHandler(
		issueSvc,
		project.NewService(nil),
		memberSvc,
		nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	mux := http.NewServeMux()
	h.Mount(mux)

	req := httptest.NewRequest(
		http.MethodPost,
		"/w/acme/issues/1/status",
		strings.NewReader("status=in_progress&redirect_to=/w/acme/projects/platform/issues"),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{
		ID:          "member1",
		Email:       "member@example.com",
		DisplayName: "Member",
	}))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d body=%q", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	got, err := store.GetByProjectAndNumber(context.Background(), "proj-a", created.IssueNumber)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != issue.StatusInProgress {
		t.Fatalf("status = %q, want %q", got.Status, issue.StatusInProgress)
	}
}

// handler memory store (duplicated for issue_test package).
type handlerMemoryStore struct {
	byProject map[string][]issue.Issue
	seq       int
}

func newHandlerMemoryStore() *handlerMemoryStore {
	return &handlerMemoryStore{byProject: map[string][]issue.Issue{}}
}

func (s *handlerMemoryStore) Create(_ context.Context, projectID, title, description, createdBy string) (issue.Issue, error) {
	next := 1
	for _, existing := range s.byProject[projectID] {
		if existing.IssueNumber >= next {
			next = existing.IssueNumber + 1
		}
	}
	s.seq++
	iss := issue.Issue{
		ID:          "issue-" + strconv.Itoa(s.seq),
		ProjectID:   projectID,
		IssueNumber: next,
		Title:       title,
		Description: description,
		Status:      issue.StatusBacklog,
		Priority:    issue.PriorityMedium,
		CreatedBy:   createdBy,
	}
	s.byProject[projectID] = append(s.byProject[projectID], iss)
	return iss, nil
}

func (s *handlerMemoryStore) ListByProject(_ context.Context, projectID string) ([]issue.Issue, error) {
	var out []issue.Issue
	for _, iss := range s.byProject[projectID] {
		if !iss.Archived {
			out = append(out, iss)
		}
	}
	if out == nil {
		out = []issue.Issue{}
	}
	return out, nil
}

func (s *handlerMemoryStore) GetByProjectAndNumber(_ context.Context, projectID string, issueNumber int) (issue.Issue, error) {
	for _, iss := range s.byProject[projectID] {
		if iss.IssueNumber == issueNumber {
			return iss, nil
		}
	}
	return issue.Issue{}, issue.ErrNotFound
}

func (s *handlerMemoryStore) GetByWorkspaceAndNumber(_ context.Context, _ string, issueNumber int) (issue.Issue, error) {
	var matches []issue.Issue
	for _, issues := range s.byProject {
		for _, iss := range issues {
			if iss.IssueNumber == issueNumber {
				matches = append(matches, iss)
			}
		}
	}
	if len(matches) != 1 {
		return issue.Issue{}, issue.ErrNotFound
	}
	return matches[0], nil
}

func (s *handlerMemoryStore) UpdateStatus(_ context.Context, id, status string) (issue.Issue, error) {
	return s.update(id, func(iss *issue.Issue) { iss.Status = status })
}

func (s *handlerMemoryStore) UpdatePriority(_ context.Context, id, priority string) (issue.Issue, error) {
	return s.update(id, func(iss *issue.Issue) { iss.Priority = priority })
}

func (s *handlerMemoryStore) UpdateAssignee(_ context.Context, id, assigneeID string) (issue.Issue, error) {
	return s.update(id, func(iss *issue.Issue) { iss.AssigneeID = assigneeID })
}

func (s *handlerMemoryStore) Archive(_ context.Context, id string) (issue.Issue, error) {
	return s.update(id, func(iss *issue.Issue) { iss.Archived = true })
}

func (s *handlerMemoryStore) update(id string, apply func(*issue.Issue)) (issue.Issue, error) {
	for projectID, issues := range s.byProject {
		for i := range issues {
			if issues[i].ID == id {
				apply(&issues[i])
				s.byProject[projectID] = issues
				return issues[i], nil
			}
		}
	}
	return issue.Issue{}, issue.ErrNotFound
}
