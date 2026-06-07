package policy_test

import (
	"testing"

	"mymcp/internal/policy"
)

func TestScopesFromClaim(t *testing.T) {
	scopes := policy.ScopesFromClaim("mcp:tools:read mcp:tools:execute")
	if len(scopes) != 2 || scopes[0] != "mcp:tools:read" || scopes[1] != "mcp:tools:execute" {
		t.Fatalf("ScopesFromClaim string = %v", scopes)
	}

	scopes = policy.ScopesFromClaim([]any{"mcp:admin"})
	if len(scopes) != 1 || scopes[0] != "mcp:admin" {
		t.Fatalf("ScopesFromClaim array = %v", scopes)
	}
}

func TestCanCallAdminTool(t *testing.T) {
	if !policy.CanCallAdminTool([]string{"mcp:admin"}, nil) {
		t.Fatal("mcp:admin scope should allow admin tool")
	}
	if !policy.CanCallAdminTool(nil, []string{"/admin"}) {
		t.Fatal("/admin group should allow admin tool")
	}
	if policy.CanCallAdminTool([]string{"mcp:tools:execute"}, []string{"user"}) {
		t.Fatal("non-admin scope/group should not allow admin tool")
	}
}
