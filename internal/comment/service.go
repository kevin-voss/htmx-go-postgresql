package comment

import (
	"context"
	"errors"
)

// ErrForbidden is returned when the actor may not perform the action.
var ErrForbidden = errors.New("comment: forbidden")

// Store is the persistence port used by Service.
type Store interface {
	Create(ctx context.Context, issueID, authorID, body string) (Comment, error)
	ListByIssue(ctx context.Context, issueID string) ([]Comment, error)
	CountByIssue(ctx context.Context, issueID string) (int, error)
	GetByID(ctx context.Context, id string) (Comment, error)
	Delete(ctx context.Context, id string) error
}

// Service implements comment business rules.
type Service struct {
	store Store
}

// NewService constructs a comment service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// ListByIssue returns comments for an issue ordered oldest-first.
func (s *Service) ListByIssue(ctx context.Context, issueID string) ([]Comment, error) {
	if issueID == "" {
		return []Comment{}, nil
	}
	return s.store.ListByIssue(ctx, issueID)
}

// CountByIssue returns how many comments an issue has.
func (s *Service) CountByIssue(ctx context.Context, issueID string) (int, error) {
	if issueID == "" {
		return 0, nil
	}
	return s.store.CountByIssue(ctx, issueID)
}

// Create validates and persists a comment.
// On validation failure it returns field errors and a zero Comment.
func (s *Service) Create(ctx context.Context, in CreateInput) (Comment, CreateErrors, error) {
	normalized := normalizeCreateInput(in)
	fieldErrs := ValidateCreate(normalized)
	if fieldErrs.Any() {
		return Comment{}, fieldErrs, nil
	}
	if normalized.IssueID == "" || normalized.AuthorID == "" {
		return Comment{}, CreateErrors{}, errors.New("comment: missing issue or author")
	}

	c, err := s.store.Create(ctx, normalized.IssueID, normalized.AuthorID, normalized.Body)
	if err != nil {
		return Comment{}, CreateErrors{}, err
	}
	return c, CreateErrors{}, nil
}

// GetByID returns a comment by id, or ErrNotFound.
func (s *Service) GetByID(ctx context.Context, id string) (Comment, error) {
	if id == "" {
		return Comment{}, ErrNotFound
	}
	return s.store.GetByID(ctx, id)
}

// Delete removes a comment when the actor is the author or an Owner.
func (s *Service) Delete(ctx context.Context, commentID, actorID, workspaceRole string) error {
	if commentID == "" {
		return ErrNotFound
	}
	c, err := s.store.GetByID(ctx, commentID)
	if err != nil {
		return err
	}
	if !c.CanDelete(actorID, workspaceRole) {
		return ErrForbidden
	}
	return s.store.Delete(ctx, commentID)
}
