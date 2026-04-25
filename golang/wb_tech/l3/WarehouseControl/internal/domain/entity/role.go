package entity

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleViewer  Role = "viewer"
	RoleAuditor Role = "auditor"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleManager, RoleViewer, RoleAuditor:
		return true
	default:
		return false
	}
}

func (r Role) CanViewItems() bool {
	return r.IsValid()
}

func (r Role) CanManageItems() bool {
	return r == RoleAdmin || r == RoleManager
}

func (r Role) CanDeleteItems() bool {
	return r == RoleAdmin
}

func (r Role) CanViewHistory() bool {
	return r.IsValid()
}

func (r Role) Permissions() []string {
	permissions := []string{"items:read", "history:read"}

	if r.CanManageItems() {
		permissions = append(permissions, "items:create", "items:update")
	}
	if r.CanDeleteItems() {
		permissions = append(permissions, "items:delete")
	}

	return permissions
}
