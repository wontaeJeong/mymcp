package mcp_test

import (
	"testing"

	"mymcp/internal/config"
	"mymcp/internal/mcp"
)

func TestProtectedResourceMetadata(t *testing.T) {
	cfg := config.Config{
		BaseURL: "http://localhost:3000",
		MCPPath: "/mcp",
		Realm:   "mcp-demo",
		Issuer:  "http://localhost:8080/realms/mcp-demo",
	}
	metadata := mcp.ProtectedResource(cfg)
	if metadata.Resource != "http://localhost:3000/mcp" {
		t.Fatalf("resource = %v", metadata.Resource)
	}
	if len(metadata.AuthorizationServers) != 1 || metadata.AuthorizationServers[0] != "http://localhost:8080/realms/mcp-demo" {
		t.Fatalf("authorization_servers = %v", metadata.AuthorizationServers)
	}
	if len(metadata.ScopesSupported) != 3 || metadata.ScopesSupported[0] != "mcp:tools:read" || metadata.ScopesSupported[1] != "mcp:tools:execute" || metadata.ScopesSupported[2] != "mcp:admin" {
		t.Fatalf("scopes_supported = %v", metadata.ScopesSupported)
	}
	if len(metadata.BearerMethodsSupported) != 1 || metadata.BearerMethodsSupported[0] != "header" {
		t.Fatalf("bearer_methods_supported = %v", metadata.BearerMethodsSupported)
	}
}

func TestWWWAuthenticateHeader(t *testing.T) {
	cfg := config.Config{BaseURL: "http://localhost:3000", Realm: "mcp-demo"}
	want := `Bearer realm="mcp-demo", resource_metadata="http://localhost:3000/.well-known/oauth-protected-resource"`
	if got := mcp.WWWAuthenticateHeader(cfg); got != want {
		t.Fatalf("WWWAuthenticateHeader = %q, want %q", got, want)
	}
}
