package issue

import (
	"strconv"
	"strings"
	"time"
)

// Status values for v1 workflow (default Backlog).
const (
	StatusBacklog    = "backlog"
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusDone       = "done"
)

// Issue is a project-scoped work item with a per-project sequential number.
type Issue struct {
	ID          string
	ProjectID   string
	IssueNumber int
	Title       string
	Description string
	Status      string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DisplayKey returns a human-readable key like PLATFORM-1.
func DisplayKey(projectSlug string, issueNumber int) string {
	if projectSlug == "" {
		return ""
	}
	return strings.ToUpper(projectSlug) + "-" + strconv.Itoa(issueNumber)
}

// StatusLabel returns a human-readable status label.
func StatusLabel(status string) string {
	switch status {
	case StatusBacklog:
		return "Backlog"
	case StatusTodo:
		return "Todo"
	case StatusInProgress:
		return "In Progress"
	case StatusDone:
		return "Done"
	default:
		return status
	}
}
