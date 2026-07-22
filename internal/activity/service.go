package activity

import (
	"context"
	"fmt"
	"strings"
)

// Service records and lists activity events.
type Service struct {
	store Store
}

// NewService constructs an activity service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Record persists an activity event.
func (s *Service) Record(ctx context.Context, e EventInput) (Event, error) {
	if err := validateEventInput(e); err != nil {
		return Event{}, err
	}
	event, err := s.store.Insert(ctx, e)
	if err != nil {
		return Event{}, fmt.Errorf("activity: record: %w", err)
	}
	return event, nil
}

// RecordAtomic runs domainWrite and records the returned event in one transaction.
func (s *Service) RecordAtomic(ctx context.Context, begin Beginner, domainWrite func(ctx context.Context, tx Tx) (EventInput, error)) error {
	if err := RunAtomic(ctx, begin, s.store, func(ctx context.Context, tx Tx) (EventInput, error) {
		event, err := domainWrite(ctx, tx)
		if err != nil {
			return EventInput{}, err
		}
		if err := validateEventInput(event); err != nil {
			return EventInput{}, err
		}
		return event, nil
	}); err != nil {
		return fmt.Errorf("activity: record atomic: %w", err)
	}
	return nil
}

// ListByProject returns recent project-scoped events.
func (s *Service) ListByProject(ctx context.Context, projectID string, limit int) ([]Event, error) {
	return s.store.ListByProject(ctx, projectID, limit)
}

// ListByWorkspace returns recent workspace-scoped events.
func (s *Service) ListByWorkspace(ctx context.Context, workspaceID string, limit int) ([]Event, error) {
	return s.store.ListByWorkspace(ctx, workspaceID, limit)
}

func validateEventInput(e EventInput) error {
	if strings.TrimSpace(e.WorkspaceID) == "" {
		return fmt.Errorf("activity: workspace id is required")
	}
	if strings.TrimSpace(e.ActorID) == "" {
		return fmt.Errorf("activity: actor id is required")
	}
	if strings.TrimSpace(e.Summary) == "" {
		return fmt.Errorf("activity: summary is required")
	}
	switch strings.TrimSpace(e.Type) {
	case TypeIssueCreated, TypeIssueStatusChanged, TypeCommentCreated:
		return nil
	default:
		return fmt.Errorf("activity: unknown event type %q", e.Type)
	}
}
