package main

import "strings"

func scopesFromClaim(scope any) []string {
	switch typed := scope.(type) {
	case string:
		return strings.Fields(typed)
	case []any:
		values := make([]string, 0, len(typed))
		for _, entry := range typed {
			if value, ok := entry.(string); ok {
				values = append(values, value)
			}
		}
		return values
	default:
		return nil
	}
}

func groupsFromClaim(groups any) []string {
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

func hasScope(auth AuthContext, scope string) bool {
	for _, candidate := range auth.Scopes {
		if candidate == scope {
			return true
		}
	}
	return false
}

func hasGroup(auth AuthContext, group string) bool {
	for _, candidate := range auth.Groups {
		if candidate == group || candidate == "/"+group {
			return true
		}
	}
	return false
}

func canListTools(auth AuthContext) bool { return hasScope(auth, "mcp:tools:read") }

func canCallTools(auth AuthContext) bool { return hasScope(auth, "mcp:tools:execute") }

func canCallAdminTool(auth AuthContext) bool {
	return hasScope(auth, "mcp:admin") || hasGroup(auth, "admin")
}
