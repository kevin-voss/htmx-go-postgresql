package workspace

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestValidateCreateOK(t *testing.T) {
	t.Parallel()

	errs := ValidateCreate(CreateInput{
		Name: "Acme Corp",
		Slug: "acme",
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
		{name: "simple", slug: "acme", wantErr: false},
		{name: "hyphenated", slug: "acme-corp", wantErr: false},
		{name: "with digits", slug: "team-42", wantErr: false},
		{name: "min length", slug: "ab", wantErr: false},
		{name: "too short", slug: "a", wantErr: true},
		{name: "uppercase rejected after normalize expectation", slug: "Acme", wantErr: true},
		{name: "leading hyphen", slug: "-acme", wantErr: true},
		{name: "trailing hyphen", slug: "acme-", wantErr: true},
		{name: "double hyphen", slug: "acme--corp", wantErr: true},
		{name: "underscore", slug: "acme_corp", wantErr: true},
		{name: "space", slug: "acme corp", wantErr: true},
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
	bySlug map[string]Workspace
	create func(ctx context.Context, name, slug, createdBy string) (Workspace, error)
}

func (s *stubStore) Create(ctx context.Context, name, slug, createdBy string) (Workspace, error) {
	if s.create != nil {
		return s.create(ctx, name, slug, createdBy)
	}
	if s.bySlug == nil {
		s.bySlug = map[string]Workspace{}
	}
	if _, ok := s.bySlug[slug]; ok {
		return Workspace{}, ErrDuplicateSlug
	}
	w := Workspace{ID: "w1", Name: name, Slug: slug, CreatedBy: createdBy}
	s.bySlug[slug] = w
	return w, nil
}

func (s *stubStore) GetBySlug(ctx context.Context, slug string) (Workspace, error) {
	if s.bySlug == nil {
		return Workspace{}, ErrNotFound
	}
	w, ok := s.bySlug[slug]
	if !ok {
		return Workspace{}, ErrNotFound
	}
	return w, nil
}

func (s *stubStore) Onboard(ctx context.Context, name, slug, createdBy, projectName, projectSlug string) (OnboardResult, error) {
	w, err := s.Create(ctx, name, slug, createdBy)
	if err != nil {
		return OnboardResult{}, err
	}
	return OnboardResult{
		Workspace:   w,
		ProjectID:   "p1",
		ProjectName: projectName,
		ProjectSlug: projectSlug,
	}, nil
}

func TestCreateNormalizesAndPersists(t *testing.T) {
	t.Parallel()

	store := &stubStore{}
	svc := NewService(store)

	ws, errs, err := svc.Create(context.Background(), CreateInput{
		Name:      "  Acme Corp  ",
		Slug:      "  Acme-Corp  ",
		CreatedBy: "user-1",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if errs.Any() {
		t.Fatalf("unexpected field errors: %+v", errs)
	}
	if ws.Name != "Acme Corp" {
		t.Fatalf("Name = %q, want Acme Corp", ws.Name)
	}
	if ws.Slug != "acme-corp" {
		t.Fatalf("Slug = %q, want acme-corp", ws.Slug)
	}
	if ws.CreatedBy != "user-1" {
		t.Fatalf("CreatedBy = %q, want user-1", ws.CreatedBy)
	}
}

func TestCreateRejectsDuplicateSlug(t *testing.T) {
	t.Parallel()

	store := &stubStore{
		bySlug: map[string]Workspace{
			"acme": {ID: "existing", Name: "Existing", Slug: "acme"},
		},
	}
	svc := NewService(store)

	_, errs, err := svc.Create(context.Background(), CreateInput{
		Name:      "Another Acme",
		Slug:      "acme",
		CreatedBy: "user-2",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if errs.Slug == "" {
		t.Fatal("want slug uniqueness error")
	}
	if !strings.Contains(strings.ToLower(errs.Slug), "taken") {
		t.Fatalf("slug error = %q, want uniqueness message", errs.Slug)
	}
}

func TestCreatePropagatesStoreErrors(t *testing.T) {
	t.Parallel()

	boom := errors.New("db down")
	store := &stubStore{
		create: func(ctx context.Context, name, slug, createdBy string) (Workspace, error) {
			return Workspace{}, boom
		},
	}
	svc := NewService(store)

	_, errs, err := svc.Create(context.Background(), CreateInput{
		Name:      "Acme",
		Slug:      "acme",
		CreatedBy: "user-1",
	})
	if err == nil {
		t.Fatal("want error")
	}
	if errs.Any() {
		t.Fatalf("unexpected field errors: %+v", errs)
	}
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want wrapped %v", err, boom)
	}
}

func TestGetBySlugNotFound(t *testing.T) {
	t.Parallel()

	svc := NewService(&stubStore{})
	_, err := svc.GetBySlug(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestOnboardCreatesWorkspaceAndProject(t *testing.T) {
	t.Parallel()

	store := &stubStore{}
	svc := NewService(store)

	result, errs, err := svc.Onboard(context.Background(), OnboardInput{
		Name:        "  Acme Corp  ",
		Slug:        "  Acme-Corp  ",
		ProjectName: "  Launch Board  ",
		CreatedBy:   "user-1",
	})
	if err != nil {
		t.Fatalf("Onboard: %v", err)
	}
	if errs.Any() {
		t.Fatalf("unexpected field errors: %+v", errs)
	}
	if result.Workspace.Slug != "acme-corp" {
		t.Fatalf("workspace slug = %q, want acme-corp", result.Workspace.Slug)
	}
	if result.ProjectName != "Launch Board" {
		t.Fatalf("project name = %q, want Launch Board", result.ProjectName)
	}
	if result.ProjectSlug != "launch-board" {
		t.Fatalf("project slug = %q, want launch-board", result.ProjectSlug)
	}
}

func TestValidateOnboardRequiresProjectName(t *testing.T) {
	t.Parallel()

	errs := ValidateOnboard(OnboardInput{
		Name:        "Acme",
		Slug:        "acme",
		ProjectName: "A",
	})
	if errs.ProjectName == "" {
		t.Fatal("want project name error")
	}
}
