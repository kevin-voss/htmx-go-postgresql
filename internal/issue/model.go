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

// Priority values for v1 (default Medium).
const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

// Issue is a project-scoped work item with a per-project sequential number.
type Issue struct {
	ID          string
	ProjectID   string
	IssueNumber int
	Title       string
	Description string
	Status      string
	Priority    string
	AssigneeID  string // empty when unassigned
	Archived    bool
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Statuses returns the ordered v1 status values.
func Statuses() []string {
	return []string{StatusBacklog, StatusTodo, StatusInProgress, StatusDone}
}

// Priorities returns the ordered v1 priority values.
func Priorities() []string {
	return []string{PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent}
}

// ValidStatus reports whether status is one of the v1 workflow values.
func ValidStatus(status string) bool {
	switch status {
	case StatusBacklog, StatusTodo, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// ValidPriority reports whether priority is one of the v1 values.
func ValidPriority(priority string) bool {
	switch priority {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
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

// PriorityLabel returns a human-readable priority label.
func PriorityLabel(priority string) string {
	switch priority {
	case PriorityLow:
		return "Low"
	case PriorityMedium:
		return "Medium"
	case PriorityHigh:
		return "High"
	case PriorityUrgent:
		return "Urgent"
	default:
		return priority
	}
}
