package issue

import "strings"

// ListFilter holds optional query-param filters for a project issue list.
// Empty fields are ignored. When multiple fields are set, they combine with AND.
type ListFilter struct {
	Status     string // exact status match
	Priority   string // exact priority match
	AssigneeID string // exact assignee user id; use "none" for unassigned
	LabelID    string // issue must have this label
	Query      string // case-insensitive substring match on title or description
}

// Active reports whether any filter field is set.
func (f ListFilter) Active() bool {
	return f.Status != "" || f.Priority != "" || f.AssigneeID != "" || f.LabelID != "" || f.Query != ""
}

// NormalizeListFilter trims fields and drops invalid status/priority values.
func NormalizeListFilter(f ListFilter) ListFilter {
	f.Status = strings.TrimSpace(f.Status)
	f.Priority = strings.TrimSpace(f.Priority)
	f.AssigneeID = strings.TrimSpace(f.AssigneeID)
	f.LabelID = strings.TrimSpace(f.LabelID)
	f.Query = strings.TrimSpace(f.Query)

	if f.Status != "" && !ValidStatus(f.Status) {
		f.Status = ""
	}
	if f.Priority != "" && !ValidPriority(f.Priority) {
		f.Priority = ""
	}
	return f
}

// escapeLike escapes %, _, and \ for use inside a SQL ILIKE pattern.
func escapeLike(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\\', '%', '_':
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// matchesListFilter reports whether issue matches filter (AND semantics).
// labelIDs are the label ids attached to the issue (may be nil).
func matchesListFilter(issue Issue, labelIDs map[string]bool, filter ListFilter) bool {
	if issue.Archived {
		return false
	}
	if filter.Status != "" && issue.Status != filter.Status {
		return false
	}
	if filter.Priority != "" && issue.Priority != filter.Priority {
		return false
	}
	if filter.AssigneeID != "" {
		if filter.AssigneeID == "none" {
			if issue.AssigneeID != "" {
				return false
			}
		} else if issue.AssigneeID != filter.AssigneeID {
			return false
		}
	}
	if filter.LabelID != "" && !labelIDs[filter.LabelID] {
		return false
	}
	if filter.Query != "" {
		q := strings.ToLower(filter.Query)
		title := strings.ToLower(issue.Title)
		desc := strings.ToLower(issue.Description)
		if !strings.Contains(title, q) && !strings.Contains(desc, q) {
			return false
		}
	}
	return true
}
