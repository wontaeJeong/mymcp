package policy

func CanCallAdminTool(scopes []string, groups []string) bool {
	return hasScope(scopes, ScopeAdmin) || hasGroup(groups, "admin")
}
