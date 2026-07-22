package issue

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"
)

// memoryLabelStore is an in-memory LabelStore for tests.
type memoryLabelStore struct {
	mu      sync.Mutex
	seq     int
	byID    map[string]Label
	byIssue map[string]map[string]bool // issueID -> labelID set
}

func newMemoryLabelStore() *memoryLabelStore {
	return &memoryLabelStore{
		byID:    map[string]Label{},
		byIssue: map[string]map[string]bool{},
	}
}

func (s *memoryLabelStore) CreateLabel(_ context.Context, workspaceID, name, color string) (Label, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.byID {
		if existing.WorkspaceID == workspaceID && strings.EqualFold(existing.Name, name) {
			return Label{}, ErrLabelExists
		}
	}
	s.seq++
	label := Label{
		ID:          "label-" + strconv.Itoa(s.seq),
		WorkspaceID: workspaceID,
		Name:        name,
		Color:       color,
		CreatedAt:   time.Now().UTC(),
	}
	s.byID[label.ID] = label
	return label, nil
}

func (s *memoryLabelStore) ListLabelsByWorkspace(_ context.Context, workspaceID string) ([]Label, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Label
	for _, label := range s.byID {
		if label.WorkspaceID == workspaceID {
			out = append(out, label)
		}
	}
	if out == nil {
		out = []Label{}
	}
	return out, nil
}

func (s *memoryLabelStore) GetLabelByID(_ context.Context, id string) (Label, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	label, ok := s.byID[id]
	if !ok {
		return Label{}, ErrLabelNotFound
	}
	return label, nil
}

func (s *memoryLabelStore) DeleteLabel(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byID[id]; !ok {
		return ErrLabelNotFound
	}
	delete(s.byID, id)
	for issueID, set := range s.byIssue {
		delete(set, id)
		if len(set) == 0 {
			delete(s.byIssue, issueID)
		}
	}
	return nil
}

func (s *memoryLabelStore) AttachLabel(_ context.Context, issueID, labelID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byID[labelID]; !ok {
		return ErrLabelNotFound
	}
	if s.byIssue[issueID] == nil {
		s.byIssue[issueID] = map[string]bool{}
	}
	s.byIssue[issueID][labelID] = true
	return nil
}

func (s *memoryLabelStore) DetachLabel(_ context.Context, issueID, labelID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	set := s.byIssue[issueID]
	if set == nil || !set[labelID] {
		return ErrLabelNotFound
	}
	delete(set, labelID)
	if len(set) == 0 {
		delete(s.byIssue, issueID)
	}
	return nil
}

func (s *memoryLabelStore) ListLabelsForIssue(_ context.Context, issueID string) ([]Label, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Label
	for labelID := range s.byIssue[issueID] {
		if label, ok := s.byID[labelID]; ok {
			out = append(out, label)
		}
	}
	if out == nil {
		out = []Label{}
	}
	return out, nil
}

func (s *memoryLabelStore) ListLabelsForIssues(_ context.Context, issueIDs []string) (map[string][]Label, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string][]Label, len(issueIDs))
	for _, issueID := range issueIDs {
		var labels []Label
		for labelID := range s.byIssue[issueID] {
			if label, ok := s.byID[labelID]; ok {
				labels = append(labels, label)
			}
		}
		if labels == nil {
			labels = []Label{}
		}
		out[issueID] = labels
	}
	return out, nil
}
