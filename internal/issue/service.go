package issue

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidStatus is returned when a status value is not allowed.
var ErrInvalidStatus = errors.New("issue: invalid status")

// ErrInvalidPriority is returned when a priority value is not allowed.
var ErrInvalidPriority = errors.New("issue: invalid priority")

// ErrInvalidAssignee is returned when the assignee is not a workspace member.
var ErrInvalidAssignee = errors.New("issue: invalid assignee")

// Store is the persistence port used by Service.
type Store interface {
	Create(ctx context.Context, projectID, title, description, createdBy string) (Issue, error)
	ListByProject(ctx context.Context, projectID string, filter ListFilter) ([]Issue, error)
	GetByProjectAndNumber(ctx context.Context, projectID string, issueNumber int) (Issue, error)
	GetByWorkspaceAndNumber(ctx context.Context, workspaceID string, issueNumber int) (Issue, error)
	UpdateStatus(ctx context.Context, id, status string) (Issue, error)
	UpdatePriority(ctx context.Context, id, priority string) (Issue, error)
	UpdateAssignee(ctx context.Context, id, assigneeID string) (Issue, error)
	Archive(ctx context.Context, id string) (Issue, error)
}

// MembershipChecker verifies a user belongs to a workspace (for assignee).
type MembershipChecker interface {
	IsMember(ctx context.Context, workspaceID, userID string) (bool, error)
}

// LabelStore is the persistence port for workspace labels and issue tagging.
type LabelStore interface {
	CreateLabel(ctx context.Context, workspaceID, name, color string) (Label, error)
	ListLabelsByWorkspace(ctx context.Context, workspaceID string) ([]Label, error)
	GetLabelByID(ctx context.Context, id string) (Label, error)
	DeleteLabel(ctx context.Context, id string) error
	AttachLabel(ctx context.Context, issueID, labelID string) error
	DetachLabel(ctx context.Context, issueID, labelID string) error
	ListLabelsForIssue(ctx context.Context, issueID string) ([]Label, error)
	ListLabelsForIssues(ctx context.Context, issueIDs []string) (map[string][]Label, error)
}

// Service implements issue business rules.
type Service struct {
	store   Store
	labels  LabelStore
	members MembershipChecker
}

// NewService constructs an issue service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// WithMembershipChecker attaches workspace membership checks for assignees.
func (s *Service) WithMembershipChecker(members MembershipChecker) *Service {
	s.members = members
	return s
}

// WithLabelStore attaches label persistence for tagging.
func (s *Service) WithLabelStore(labels LabelStore) *Service {
	s.labels = labels
	return s
}

// ListByProject returns non-archived issues for a project ordered by number.
// Optional filter fields combine with AND; empty fields are ignored.
func (s *Service) ListByProject(ctx context.Context, projectID string, filter ListFilter) ([]Issue, error) {
	if projectID == "" {
		return nil, nil
	}
	return s.store.ListByProject(ctx, projectID, NormalizeListFilter(filter))
}

// GetByProjectAndNumber returns an issue, or ErrNotFound.
func (s *Service) GetByProjectAndNumber(ctx context.Context, projectID string, issueNumber int) (Issue, error) {
	if projectID == "" || issueNumber < 1 {
		return Issue{}, ErrNotFound
	}
	return s.store.GetByProjectAndNumber(ctx, projectID, issueNumber)
}

// GetByWorkspaceAndNumber returns an issue unique within a workspace, or ErrNotFound.
func (s *Service) GetByWorkspaceAndNumber(ctx context.Context, workspaceID string, issueNumber int) (Issue, error) {
	if workspaceID == "" || issueNumber < 1 {
		return Issue{}, ErrNotFound
	}
	return s.store.GetByWorkspaceAndNumber(ctx, workspaceID, issueNumber)
}

// Create validates and persists an issue with the next per-project number.
// On validation failure it returns field errors and a zero Issue.
func (s *Service) Create(ctx context.Context, in CreateInput) (Issue, CreateErrors, error) {
	normalized := normalizeCreateInput(in)
	fieldErrs := ValidateCreate(normalized)
	if fieldErrs.Any() {
		return Issue{}, fieldErrs, nil
	}
	if normalized.ProjectID == "" || normalized.CreatedBy == "" {
		return Issue{}, CreateErrors{}, fmt.Errorf("issue: create: missing project or creator")
	}

	issue, err := s.store.Create(ctx, normalized.ProjectID, normalized.Title, normalized.Description, normalized.CreatedBy)
	if err != nil {
		return Issue{}, CreateErrors{}, fmt.Errorf("issue: create: %w", err)
	}
	return issue, CreateErrors{}, nil
}

// UpdateStatus changes an issue's workflow status.
func (s *Service) UpdateStatus(ctx context.Context, workspaceID string, issueNumber int, status string) (Issue, error) {
	status = strings.TrimSpace(status)
	if !ValidStatus(status) {
		return Issue{}, ErrInvalidStatus
	}
	issue, err := s.GetByWorkspaceAndNumber(ctx, workspaceID, issueNumber)
	if err != nil {
		return Issue{}, err
	}
	updated, err := s.store.UpdateStatus(ctx, issue.ID, status)
	if err != nil {
		return Issue{}, fmt.Errorf("issue: update status: %w", err)
	}
	return updated, nil
}

// UpdatePriority changes an issue's priority.
func (s *Service) UpdatePriority(ctx context.Context, workspaceID string, issueNumber int, priority string) (Issue, error) {
	priority = strings.TrimSpace(priority)
	if !ValidPriority(priority) {
		return Issue{}, ErrInvalidPriority
	}
	issue, err := s.GetByWorkspaceAndNumber(ctx, workspaceID, issueNumber)
	if err != nil {
		return Issue{}, err
	}
	updated, err := s.store.UpdatePriority(ctx, issue.ID, priority)
	if err != nil {
		return Issue{}, fmt.Errorf("issue: update priority: %w", err)
	}
	return updated, nil
}

// UpdateAssignee sets or clears the assignee (must be a workspace member when set).
func (s *Service) UpdateAssignee(ctx context.Context, workspaceID string, issueNumber int, assigneeID string) (Issue, error) {
	assigneeID = strings.TrimSpace(assigneeID)
	if assigneeID != "" {
		if s.members == nil {
			return Issue{}, fmt.Errorf("issue: update assignee: membership checker not configured")
		}
		ok, err := s.members.IsMember(ctx, workspaceID, assigneeID)
		if err != nil {
			return Issue{}, fmt.Errorf("issue: update assignee: %w", err)
		}
		if !ok {
			return Issue{}, ErrInvalidAssignee
		}
	}

	issue, err := s.GetByWorkspaceAndNumber(ctx, workspaceID, issueNumber)
	if err != nil {
		return Issue{}, err
	}
	updated, err := s.store.UpdateAssignee(ctx, issue.ID, assigneeID)
	if err != nil {
		return Issue{}, fmt.Errorf("issue: update assignee: %w", err)
	}
	return updated, nil
}

// Archive soft-archives an issue so it is hidden from default lists.
func (s *Service) Archive(ctx context.Context, workspaceID string, issueNumber int) (Issue, error) {
	issue, err := s.GetByWorkspaceAndNumber(ctx, workspaceID, issueNumber)
	if err != nil {
		return Issue{}, err
	}
	updated, err := s.store.Archive(ctx, issue.ID)
	if err != nil {
		return Issue{}, fmt.Errorf("issue: archive: %w", err)
	}
	return updated, nil
}

// ListLabels returns all labels for a workspace.
func (s *Service) ListLabels(ctx context.Context, workspaceID string) ([]Label, error) {
	if workspaceID == "" || s.labels == nil {
		return []Label{}, nil
	}
	return s.labels.ListLabelsByWorkspace(ctx, workspaceID)
}

// CreateLabel validates and persists a workspace label.
func (s *Service) CreateLabel(ctx context.Context, in CreateLabelInput) (Label, CreateLabelErrors, error) {
	if s.labels == nil {
		return Label{}, CreateLabelErrors{}, fmt.Errorf("issue: create label: label store not configured")
	}
	normalized := normalizeCreateLabelInput(in)
	fieldErrs := ValidateCreateLabel(normalized)
	if fieldErrs.Any() {
		return Label{}, fieldErrs, nil
	}
	if normalized.WorkspaceID == "" {
		return Label{}, CreateLabelErrors{}, fmt.Errorf("issue: create label: missing workspace")
	}

	label, err := s.labels.CreateLabel(ctx, normalized.WorkspaceID, normalized.Name, normalized.Color)
	if err != nil {
		if errors.Is(err, ErrLabelExists) {
			return Label{}, CreateLabelErrors{Name: "A label with this name already exists."}, nil
		}
		return Label{}, CreateLabelErrors{}, fmt.Errorf("issue: create label: %w", err)
	}
	return label, CreateLabelErrors{}, nil
}

// DeleteLabel removes a label from a workspace (must belong to that workspace).
func (s *Service) DeleteLabel(ctx context.Context, workspaceID, labelID string) error {
	if s.labels == nil {
		return fmt.Errorf("issue: delete label: label store not configured")
	}
	label, err := s.labels.GetLabelByID(ctx, labelID)
	if err != nil {
		return err
	}
	if label.WorkspaceID != workspaceID {
		return ErrLabelNotFound
	}
	if err := s.labels.DeleteLabel(ctx, labelID); err != nil {
		return fmt.Errorf("issue: delete label: %w", err)
	}
	return nil
}

// LabelsForIssue returns labels attached to an issue.
func (s *Service) LabelsForIssue(ctx context.Context, issueID string) ([]Label, error) {
	if issueID == "" || s.labels == nil {
		return []Label{}, nil
	}
	return s.labels.ListLabelsForIssue(ctx, issueID)
}

// LabelsForIssues returns attached labels keyed by issue id.
func (s *Service) LabelsForIssues(ctx context.Context, issueIDs []string) (map[string][]Label, error) {
	if len(issueIDs) == 0 || s.labels == nil {
		return map[string][]Label{}, nil
	}
	return s.labels.ListLabelsForIssues(ctx, issueIDs)
}

// AttachLabel tags an issue with a workspace label.
func (s *Service) AttachLabel(ctx context.Context, workspaceID string, issueNumber int, labelID string) (Issue, error) {
	if s.labels == nil {
		return Issue{}, fmt.Errorf("issue: attach label: label store not configured")
	}
	labelID = strings.TrimSpace(labelID)
	if labelID == "" {
		return Issue{}, ErrLabelNotFound
	}

	issue, err := s.GetByWorkspaceAndNumber(ctx, workspaceID, issueNumber)
	if err != nil {
		return Issue{}, err
	}
	label, err := s.labels.GetLabelByID(ctx, labelID)
	if err != nil {
		return Issue{}, err
	}
	if label.WorkspaceID != workspaceID {
		return Issue{}, ErrLabelNotInWorkspace
	}
	if err := s.labels.AttachLabel(ctx, issue.ID, label.ID); err != nil {
		return Issue{}, fmt.Errorf("issue: attach label: %w", err)
	}
	return issue, nil
}

// DetachLabel removes a label from an issue.
func (s *Service) DetachLabel(ctx context.Context, workspaceID string, issueNumber int, labelID string) (Issue, error) {
	if s.labels == nil {
		return Issue{}, fmt.Errorf("issue: detach label: label store not configured")
	}
	labelID = strings.TrimSpace(labelID)
	if labelID == "" {
		return Issue{}, ErrLabelNotFound
	}

	issue, err := s.GetByWorkspaceAndNumber(ctx, workspaceID, issueNumber)
	if err != nil {
		return Issue{}, err
	}
	label, err := s.labels.GetLabelByID(ctx, labelID)
	if err != nil {
		return Issue{}, err
	}
	if label.WorkspaceID != workspaceID {
		return Issue{}, ErrLabelNotInWorkspace
	}
	if err := s.labels.DetachLabel(ctx, issue.ID, label.ID); err != nil {
		return Issue{}, err
	}
	return issue, nil
}
