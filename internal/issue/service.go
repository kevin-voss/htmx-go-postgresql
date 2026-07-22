package issue

import (
	"context"
	"fmt"
)

// Store is the persistence port used by Service.
type Store interface {
	Create(ctx context.Context, projectID, title, description, createdBy string) (Issue, error)
	ListByProject(ctx context.Context, projectID string) ([]Issue, error)
	GetByProjectAndNumber(ctx context.Context, projectID string, issueNumber int) (Issue, error)
	GetByWorkspaceAndNumber(ctx context.Context, workspaceID string, issueNumber int) (Issue, error)
}

// Service implements issue business rules.
type Service struct {
	store Store
}

// NewService constructs an issue service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// ListByProject returns issues for a project ordered by number.
func (s *Service) ListByProject(ctx context.Context, projectID string) ([]Issue, error) {
	if projectID == "" {
		return nil, nil
	}
	return s.store.ListByProject(ctx, projectID)
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
