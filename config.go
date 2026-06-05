package main

import (
	"os"
	"strconv"
	"strings"
)

type AppConfig struct {
	Port        int
	BaseURL     string
	MCPPath     string
	Realm       string
	Issuer      string
	JWKSURI     string
	Audience    string
	RealmName   string
	CORSOrigins []string
}

func trimTrailingSlash(value string) string {
	return strings.TrimRight(value, "/")
}

func csv(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	entries := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			entries = append(entries, trimmed)
		}
	}
	return entries
}

func getenv(env map[string]string, key, fallback string) string {
	if value, ok := env[key]; ok && value != "" {
		return value
	}
	return fallback
}

func LoadConfig(env map[string]string) AppConfig {
	if env == nil {
		env = map[string]string{}
		for _, entry := range os.Environ() {
			key, value, ok := strings.Cut(entry, "=")
			if ok {
				env[key] = value
			}
		}
	}

	baseURL := trimTrailingSlash(getenv(env, "MCP_BASE_URL", "http://localhost:3000"))
	keycloakBaseURL := trimTrailingSlash(getenv(env, "KEYCLOAK_BASE_URL", "http://localhost:8080"))
	realmName := getenv(env, "KEYCLOAK_REALM", "mcp-demo")
	issuer := getenv(env, "KEYCLOAK_ISSUER", keycloakBaseURL+"/realms/"+realmName)
	port, err := strconv.Atoi(getenv(env, "PORT", "3000"))
	if err != nil {
		port = 3000
	}

	return AppConfig{
		Port:        port,
		BaseURL:     baseURL,
		MCPPath:     getenv(env, "MCP_PATH", "/mcp"),
		Realm:       getenv(env, "AUTH_REALM", "mcp-demo"),
		RealmName:   realmName,
		Issuer:      issuer,
		JWKSURI:     getenv(env, "KEYCLOAK_JWKS_URI", issuer+"/protocol/openid-connect/certs"),
		Audience:    getenv(env, "MCP_AUDIENCE", "mcp-demo-resource"),
		CORSOrigins: csv(env["CORS_ORIGINS"]),
	}
}
