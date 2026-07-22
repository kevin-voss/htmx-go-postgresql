package member

// Role is a workspace-scoped authorization level.
type Role string

const (
	RoleOwner  Role = "owner"
	RoleMember Role = "member"
	RoleViewer Role = "viewer"
)

// Valid reports whether r is one of the three v1 roles.
func (r Role) Valid() bool {
	switch r {
	case RoleOwner, RoleMember, RoleViewer:
		return true
	default:
		return false
	}
}

// rank returns a comparable privilege level (higher is more privileged).
func (r Role) rank() int {
	switch r {
	case RoleOwner:
		return 3
	case RoleMember:
		return 2
	case RoleViewer:
		return 1
	default:
		return 0
	}
}

// AtLeast reports whether r is at least as privileged as min.
func (r Role) AtLeast(min Role) bool {
	return r.rank() >= min.rank()
}

// CanMutate reports whether r may change workspace data (Owner or Member).
func (r Role) CanMutate() bool {
	return r.AtLeast(RoleMember)
}

// CanManageSettings reports whether r may access workspace settings (Owner).
func (r Role) CanManageSettings() bool {
	return r == RoleOwner
}

// CanInvite reports whether r may invite members (Owner).
func (r Role) CanInvite() bool {
	return r == RoleOwner
}

// CanManageMembers reports whether r may change roles or remove members (Owner).
func (r Role) CanManageMembers() bool {
	return r == RoleOwner
}

// Invitable reports whether r is a valid invitation role (Member or Viewer).
func (r Role) Invitable() bool {
	return r == RoleMember || r == RoleViewer
}
