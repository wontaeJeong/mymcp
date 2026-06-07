package policy

func GroupsFromClaim(groups any) []string {
	entries, ok := groups.([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(entries))
	for _, entry := range entries {
		if value, ok := entry.(string); ok {
			values = append(values, value)
		}
	}
	return values
}

func hasGroup(groups []string, group string) bool {
	for _, candidate := range groups {
		if candidate == group || candidate == "/"+group {
			return true
		}
	}
	return false
}
