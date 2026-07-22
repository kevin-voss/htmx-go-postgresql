package project

import (
	"context"
	"strings"
	"testing"
)

func TestSlugFromName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "simple", in: "Launch Board", want: "launch-board"},
		{name: "trim", in: "  My App  ", want: "my-app"},
		{name: "punctuation", in: "Hello, World!", want: "hello-world"},
		{name: "empty", in: "   ", want: "project"},
		{name: "symbols only", in: "!!!", want: "project"},
		{name: "collapse hyphens", in: "a -- b", want: "a-b"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := SlugFromName(tc.in); got != tc.want {
				t.Fatalf("SlugFromName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestValidateCreateOK(t *testing.T) {
	t.Parallel()

	errs := ValidateCreate(CreateInput{
		Name: "Platform",
		Slug: "platform",
	})
	if errs.Any() {
		t.Fatalf("unexpected errors: %+v", errs)
	}
}

func TestValidateCreateFieldErrors(t *testing.T) {
	t.Parallel()

	errs := ValidateCreate(CreateInput{
		Name: "A",
		Slug: "BAD_SLUG",
	})
	if errs.Name == "" {
		t.Fatal("want name error")
	}
	if errs.Slug == "" {
		t.Fatal("want slug error")
	}
}

func TestValidateCreateSlugRules(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		slug    string
		wantErr bool
	}{
		{name: "simple", slug: "platform", wantErr: false},
		{name: "hyphenated", slug: "launch-board", wantErr: false},
		{name: "with digits", slug: "app-42", wantErr: false},
		{name: "min length", slug: "ab", wantErr: false},
		{name: "too short", slug: "a", wantErr: true},
		{name: "uppercase rejected", slug: "Platform", wantErr: true},
		{name: "leading hyphen", slug: "-app", wantErr: true},
		{name: "trailing hyphen", slug: "app-", wantErr: true},
		{name: "double hyphen", slug: "app--board", wantErr: true},
		{name: "underscore", slug: "app_board", wantErr: true},
		{name: "space", slug: "app board", wantErr: true},
		{name: "too long", slug: strings.Repeat("a", 49), wantErr: true},
		{name: "max length", slug: strings.Repeat("a", 48), wantErr: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			errs := ValidateCreate(CreateInput{Name: "Valid Name", Slug: tc.slug})
			gotErr := errs.Slug != ""
			if gotErr != tc.wantErr {
				t.Fatalf("slug %q: gotErr=%v wantErr=%v errs=%+v", tc.slug, gotErr, tc.wantErr, errs)
			}
		})
	}
}

type stubStore struct {
	byWorkspace map[string]map[string]Project
}

func (s *stubStore) ensure() {
	if s.byWorkspace == nil {
		s.byWorkspace = map[string]map[string]Project{}
	}
}

func (s *stubStore) Create(_ context.Context, workspaceID, name, slug, createdBy string) (Project, error) {
	s.ensure()
	if _, ok := s.byWorkspace[workspaceID]; !ok {
		s.byWorkspace[workspaceID] = map[string]Project{}
	}
	if _, ok := s.byWorkspace[workspaceID][slug]; ok {
		return Project{}, ErrDuplicateSlug
	}
	p := Project{
		ID:          "p-" + workspaceID + "-" + slug,
		WorkspaceID: workspaceID,
		Name:        name,
		Slug:        slug,
		CreatedBy:   createdBy,
	}
	s.byWorkspace[workspaceID][slug] = p
	return p, nil
}

func (s *stubStore) ListByWorkspace(_ context.Context, workspaceID string) ([]Project, error) {
	s.ensure()
	out := make([]Project, 0, len(s.byWorkspace[workspaceID]))
	for _, p := range s.byWorkspace[workspaceID] {
		out = append(out, p)
	}
	return out, nil
}

func (s *stubStore) GetByWorkspaceAndSlug(_ context.Context, workspaceID, slug string) (Project, error) {
	s.ensure()
	p, ok := s.byWorkspace[workspaceID][slug]
	if !ok {
		return Project{}, ErrNotFound
	}
	return p, nil
}

func TestCreateNormalizesAndPersists(t *testing.T) {
	t.Parallel()

	store := &stubStore{}
	svc := NewService(store)

	p, errs, err := svc.Create(context.Background(), CreateInput{
		WorkspaceID: "ws-1",
		Name:        "  Launch Board  ",
		Slug:        "  Launch-Board  ",
		CreatedBy:   "user-1",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if errs.Any() {
		t.Fatalf("unexpected field errors: %+v", errs)
	}
	if p.Name != "Launch Board" {
		t.Fatalf("name = %q, want Launch Board", p.Name)
	}
	if p.Slug != "launch-board" {
		t.Fatalf("slug = %q, want launch-board", p.Slug)
	}
	if p.WorkspaceID != "ws-1" {
		t.Fatalf("workspace = %q, want ws-1", p.WorkspaceID)
	}
}

func TestCreateRejectsDuplicateSlugInSameWorkspace(t *testing.T) {
	t.Parallel()

	store := &stubStore{}
	svc := NewService(store)

	_, errs, err := svc.Create(context.Background(), CreateInput{
		WorkspaceID: "ws-1",
		Name:        "Platform",
		Slug:        "platform",
		CreatedBy:   "user-1",
	})
	if err != nil || errs.Any() {
		t.Fatalf("first create: err=%v errs=%+v", err, errs)
	}

	_, errs, err = svc.Create(context.Background(), CreateInput{
		WorkspaceID: "ws-1",
		Name:        "Platform Two",
		Slug:        "platform",
		CreatedBy:   "user-2",
	})
	if err != nil {
		t.Fatalf("second create unexpected error: %v", err)
	}
	if errs.Slug == "" {
		t.Fatal("want duplicate slug field error in same workspace")
	}
}

func TestCreateAllowsSameSlugInDifferentWorkspaces(t *testing.T) {
	t.Parallel()

	store := &stubStore{}
	svc := NewService(store)

	_, errs, err := svc.Create(context.Background(), CreateInput{
		WorkspaceID: "ws-1",
		Name:        "Platform",
		Slug:        "platform",
		CreatedBy:   "user-1",
	})
	if err != nil || errs.Any() {
		t.Fatalf("first create: err=%v errs=%+v", err, errs)
	}

	p, errs, err := svc.Create(context.Background(), CreateInput{
		WorkspaceID: "ws-2",
		Name:        "Platform",
		Slug:        "platform",
		CreatedBy:   "user-1",
	})
	if err != nil {
		t.Fatalf("second create error: %v", err)
	}
	if errs.Any() {
		t.Fatalf("unexpected field errors across workspaces: %+v", errs)
	}
	if p.WorkspaceID != "ws-2" || p.Slug != "platform" {
		t.Fatalf("got workspace=%q slug=%q", p.WorkspaceID, p.Slug)
	}
}

func TestGetByWorkspaceAndSlug(t *testing.T) {
	t.Parallel()

	store := &stubStore{}
	svc := NewService(store)

	created, _, err := svc.Create(context.Background(), CreateInput{
		WorkspaceID: "ws-1",
		Name:        "Platform",
		Slug:        "platform",
		CreatedBy:   "user-1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := svc.GetByWorkspaceAndSlug(context.Background(), "ws-1", "PLATFORM")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("id = %q, want %q", got.ID, created.ID)
	}

	_, err = svc.GetByWorkspaceAndSlug(context.Background(), "ws-2", "platform")
	if err != ErrNotFound {
		t.Fatalf("cross-workspace get err = %v, want ErrNotFound", err)
	}
}
