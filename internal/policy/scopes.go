package policy

import "strings"

const (
	ScopeToolsRead    = "mcp:tools:read"
	ScopeToolsExecute = "mcp:tools:execute"
	ScopeAdmin        = "mcp:admin"
)

func ScopesFromClaim(scope any) []string {
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

func hasScope(scopes []string, scope string) bool {
	for _, candidate := range scopes {
		if candidate == scope {
			return true
		}
	}
	return false
}

func CanListTools(scopes []string) bool { return hasScope(scopes, ScopeToolsRead) }

func CanCallTools(scopes []string) bool { return hasScope(scopes, ScopeToolsExecute) }
