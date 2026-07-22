package issue

import (
	"context"
	"strconv"
	"sync"
	"testing"
)

// memoryStore mirrors Repository numbering: next = max(issue_number)+1 per project.
type memoryStore struct {
	mu        sync.Mutex
	byProject map[string][]Issue
	seq       int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{byProject: map[string][]Issue{}}
}

func (s *memoryStore) Create(_ context.Context, projectID, title, description, createdBy string) (Issue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := 1
	for _, existing := range s.byProject[projectID] {
		if existing.IssueNumber >= next {
			next = existing.IssueNumber + 1
		}
	}
	s.seq++
	issue := Issue{
		ID:          "issue-" + strconv.Itoa(s.seq),
		ProjectID:   projectID,
		IssueNumber: next,
		Title:       title,
		Description: description,
		Status:      StatusBacklog,
		CreatedBy:   createdBy,
	}
	s.byProject[projectID] = append(s.byProject[projectID], issue)
	return issue, nil
}

func (s *memoryStore) ListByProject(_ context.Context, projectID string) ([]Issue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := append([]Issue{}, s.byProject[projectID]...)
	if out == nil {
		out = []Issue{}
	}
	return out, nil
}

func (s *memoryStore) GetByProjectAndNumber(_ context.Context, projectID string, issueNumber int) (Issue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, issue := range s.byProject[projectID] {
		if issue.IssueNumber == issueNumber {
			return issue, nil
		}
	}
	return Issue{}, ErrNotFound
}

func (s *memoryStore) GetByWorkspaceAndNumber(_ context.Context, _ string, issueNumber int) (Issue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var matches []Issue
	for _, issues := range s.byProject {
		for _, issue := range issues {
			if issue.IssueNumber == issueNumber {
				matches = append(matches, issue)
			}
		}
	}
	if len(matches) != 1 {
		return Issue{}, ErrNotFound
	}
	return matches[0], nil
}

func TestRepositoryNumberingIncrementsPerProject(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	svc := NewService(store)
	ctx := context.Background()

	a1, errs, err := svc.Create(ctx, CreateInput{
		ProjectID: "proj-a",
		Title:     "First A",
		CreatedBy: "user-1",
	})
	if err != nil || errs.Any() {
		t.Fatalf("create a1: err=%v errs=%+v", err, errs)
	}
	if a1.IssueNumber != 1 {
		t.Fatalf("a1 number = %d, want 1", a1.IssueNumber)
	}
	if a1.Status != StatusBacklog {
		t.Fatalf("a1 status = %q, want %q", a1.Status, StatusBacklog)
	}

	a2, errs, err := svc.Create(ctx, CreateInput{
		ProjectID: "proj-a",
		Title:     "Second A",
		CreatedBy: "user-1",
	})
	if err != nil || errs.Any() {
		t.Fatalf("create a2: err=%v errs=%+v", err, errs)
	}
	if a2.IssueNumber != 2 {
		t.Fatalf("a2 number = %d, want 2", a2.IssueNumber)
	}

	b1, errs, err := svc.Create(ctx, CreateInput{
		ProjectID: "proj-b",
		Title:     "First B",
		CreatedBy: "user-1",
	})
	if err != nil || errs.Any() {
		t.Fatalf("create b1: err=%v errs=%+v", err, errs)
	}
	if b1.IssueNumber != 1 {
		t.Fatalf("b1 number = %d, want 1 (independent per project)", b1.IssueNumber)
	}

	listed, err := svc.ListByProject(ctx, "proj-a")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("list len = %d, want 2", len(listed))
	}
	if listed[0].IssueNumber != 1 || listed[1].IssueNumber != 2 {
		t.Fatalf("list numbers = %d,%d want 1,2", listed[0].IssueNumber, listed[1].IssueNumber)
	}
}

func TestRepositoryNumberingConcurrentPerProject(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	svc := NewService(store)
	ctx := context.Background()

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make(chan error, n)
	for range n {
		go func() {
			defer wg.Done()
			_, fieldErrs, err := svc.Create(ctx, CreateInput{
				ProjectID: "proj-concurrent",
				Title:     "Issue",
				CreatedBy: "user-1",
			})
			if err != nil {
				errs <- err
				return
			}
			if fieldErrs.Any() {
				errs <- errField
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent create: %v", err)
	}

	listed, err := svc.ListByProject(ctx, "proj-concurrent")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != n {
		t.Fatalf("list len = %d, want %d", len(listed), n)
	}
	seen := map[int]bool{}
	for _, issue := range listed {
		if seen[issue.IssueNumber] {
			t.Fatalf("duplicate issue_number %d", issue.IssueNumber)
		}
		seen[issue.IssueNumber] = true
	}
	for want := 1; want <= n; want++ {
		if !seen[want] {
			t.Fatalf("missing issue_number %d", want)
		}
	}
}

var errField = errString("field errors")

type errString string

func (e errString) Error() string { return string(e) }
