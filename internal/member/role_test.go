package member

import "testing"

func TestRoleValid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		role Role
		want bool
	}{
		{RoleOwner, true},
		{RoleMember, true},
		{RoleViewer, true},
		{Role("admin"), false},
		{Role(""), false},
	}
	for _, tc := range cases {
		if got := tc.role.Valid(); got != tc.want {
			t.Fatalf("Role(%q).Valid() = %v, want %v", tc.role, got, tc.want)
		}
	}
}

func TestRoleAtLeastAndCapabilities(t *testing.T) {
	t.Parallel()

	if !RoleOwner.AtLeast(RoleOwner) || !RoleOwner.AtLeast(RoleMember) || !RoleOwner.AtLeast(RoleViewer) {
		t.Fatal("owner should outrank all roles")
	}
	if RoleMember.AtLeast(RoleOwner) {
		t.Fatal("member must not outrank owner")
	}
	if !RoleMember.CanMutate() {
		t.Fatal("member should mutate")
	}
	if !RoleOwner.CanMutate() {
		t.Fatal("owner should mutate")
	}
	if RoleViewer.CanMutate() {
		t.Fatal("viewer must not mutate")
	}
	if !RoleOwner.CanManageSettings() {
		t.Fatal("owner should manage settings")
	}
	if RoleMember.CanManageSettings() || RoleViewer.CanManageSettings() {
		t.Fatal("non-owners must not manage settings")
	}
}
