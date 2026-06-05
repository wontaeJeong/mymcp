package main

import "fmt"

var MCPScopes = []string{"mcp:tools:read", "mcp:tools:execute", "mcp:admin"}

type ProtectedResourceMetadata struct {
	Resource               string   `json:"resource"`
	AuthorizationServers   []string `json:"authorization_servers"`
	ScopesSupported        []string `json:"scopes_supported"`
	BearerMethodsSupported []string `json:"bearer_methods_supported"`
}

func protectedResourceMetadata(config AppConfig) ProtectedResourceMetadata {
	return ProtectedResourceMetadata{
		Resource:               config.BaseURL + config.MCPPath,
		AuthorizationServers:   []string{config.Issuer},
		ScopesSupported:        append([]string(nil), MCPScopes...),
		BearerMethodsSupported: []string{"header"},
	}
}

func wwwAuthenticateHeader(config AppConfig) string {
	return fmt.Sprintf(`Bearer realm="%s", resource_metadata="%s/.well-known/oauth-protected-resource"`, config.Realm, config.BaseURL)
}
