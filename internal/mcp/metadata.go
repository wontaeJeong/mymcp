package mcp

import (
	"fmt"

	"mymcp/internal/config"
	"mymcp/internal/policy"
)

var MCPScopes = []string{policy.ScopeToolsRead, policy.ScopeToolsExecute, policy.ScopeAdmin}

type ProtectedResourceMetadata struct {
	Resource               string   `json:"resource"`
	AuthorizationServers   []string `json:"authorization_servers"`
	ScopesSupported        []string `json:"scopes_supported"`
	BearerMethodsSupported []string `json:"bearer_methods_supported"`
}

func ProtectedResource(cfg config.Config) ProtectedResourceMetadata {
	return ProtectedResourceMetadata{
		Resource:               cfg.BaseURL + cfg.MCPPath,
		AuthorizationServers:   []string{cfg.Issuer},
		ScopesSupported:        append([]string(nil), MCPScopes...),
		BearerMethodsSupported: []string{"header"},
	}
}

func WWWAuthenticateHeader(cfg config.Config) string {
	return fmt.Sprintf(`Bearer realm="%s", resource_metadata="%s/.well-known/oauth-protected-resource"`, cfg.Realm, cfg.BaseURL)
}
