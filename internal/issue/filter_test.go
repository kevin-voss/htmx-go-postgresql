package issue

import (
	"context"
	"testing"
)

func TestListByProjectFilterCombinations(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	svc := NewService(store).WithMembershipChecker(stubMembers{
		members: map[string]bool{"user-a": true, "user-b": true},
	})
	ctx := context.Background()

	mk := func(title, description string) Issue {
		t.Helper()
		issue, errs, err := svc.Create(ctx, CreateInput{
			ProjectID:   "proj-filter",
			Title:       title,
			Description: description,
			CreatedBy:   "user-a",
		})
		if err != nil || errs.Any() {
			t.Fatalf("create %q: err=%v errs=%+v", title, err, errs)
		}
		return issue
	}

	bug := mk("Fix login bug", "Password reset fails on Safari")
	feature := mk("Add dark mode", "Theme toggle for the app shell")
	chore := mk("Update docs", "Mention search filters in the handbook")

	if _, err := svc.UpdateStatus(ctx, "ws", bug.IssueNumber, StatusInProgress); err != nil {
		t.Fatalf("status bug: %v", err)
	}
	if _, err := svc.UpdateStatus(ctx, "ws", feature.IssueNumber, StatusTodo); err != nil {
		t.Fatalf("status feature: %v", err)
	}
	if _, err := svc.UpdatePriority(ctx, "ws", bug.IssueNumber, PriorityHigh); err != nil {
		t.Fatalf("priority bug: %v", err)
	}
	if _, err := svc.UpdateAssignee(ctx, "ws", bug.IssueNumber, "user-a"); err != nil {
		t.Fatalf("assignee bug: %v", err)
	}
	if _, err := svc.UpdateAssignee(ctx, "ws", feature.IssueNumber, "user-b"); err != nil {
		t.Fatalf("assignee feature: %v", err)
	}

	store.setLabels(bug.ID, "label-bug")
	store.setLabels(feature.ID, "label-feature")
	store.setLabels(chore.ID, "label-docs", "label-bug")

	cases := []struct {
		name   string
		filter ListFilter
		want   []int
	}{
		{
			name:   "no filters returns all",
			filter: ListFilter{},
			want:   []int{1, 2, 3},
		},
		{
			name:   "status only",
			filter: ListFilter{Status: StatusInProgress},
			want:   []int{1},
		},
		{
			name:   "assignee only",
			filter: ListFilter{AssigneeID: "user-b"},
			want:   []int{2},
		},
		{
			name:   "unassigned only",
			filter: ListFilter{AssigneeID: "none"},
			want:   []int{3},
		},
		{
			name:   "priority only",
			filter: ListFilter{Priority: PriorityHigh},
			want:   []int{1},
		},
		{
			name:   "label only",
			filter: ListFilter{LabelID: "label-bug"},
			want:   []int{1, 3},
		},
		{
			name:   "text search title",
			filter: ListFilter{Query: "dark"},
			want:   []int{2},
		},
		{
			name:   "text search description",
			filter: ListFilter{Query: "Safari"},
			want:   []int{1},
		},
		{
			name:   "text search case insensitive",
			filter: ListFilter{Query: "LOGIN"},
			want:   []int{1},
		},
		{
			name: "status and assignee AND",
			filter: ListFilter{
				Status:     StatusInProgress,
				AssigneeID: "user-a",
			},
			want: []int{1},
		},
		{
			name: "status and assignee no match",
			filter: ListFilter{
				Status:     StatusInProgress,
				AssigneeID: "user-b",
			},
			want: []int{},
		},
		{
			name: "query and label AND",
			filter: ListFilter{
				Query:   "docs",
				LabelID: "label-bug",
			},
			want: []int{3},
		},
		{
			name: "priority status label AND",
			filter: ListFilter{
				Status:   StatusInProgress,
				Priority: PriorityHigh,
				LabelID:  "label-bug",
			},
			want: []int{1},
		},
		{
			name: "injection-like query is literal substring",
			filter: ListFilter{
				Query: `%' OR 1=1 --`,
			},
			want: []int{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := svc.ListByProject(ctx, "proj-filter", tc.filter)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("len = %d, want %d (got numbers %v)", len(got), len(tc.want), numbers(got))
			}
			for i, wantNum := range tc.want {
				if got[i].IssueNumber != wantNum {
					t.Fatalf("got numbers %v, want %v", numbers(got), tc.want)
				}
			}
		})
	}
}

func TestNormalizeListFilterDropsInvalidEnums(t *testing.T) {
	t.Parallel()

	got := NormalizeListFilter(ListFilter{
		Status:   "not-a-status",
		Priority: "critical",
		Query:    "  ship  ",
	})
	if got.Status != "" {
		t.Fatalf("status = %q, want empty", got.Status)
	}
	if got.Priority != "" {
		t.Fatalf("priority = %q, want empty", got.Priority)
	}
	if got.Query != "ship" {
		t.Fatalf("query = %q, want ship", got.Query)
	}
}

func TestEscapeLike(t *testing.T) {
	t.Parallel()

	got := escapeLike(`100%_done\`)
	want := `100\%\_done\\`
	if got != want {
		t.Fatalf("escapeLike = %q, want %q", got, want)
	}
}

func numbers(issues []Issue) []int {
	out := make([]int, len(issues))
	for i, issue := range issues {
		out[i] = issue.IssueNumber
	}
	return out
}
