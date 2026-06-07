package config

import "strconv"

type Config struct {
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

func Load(env map[string]string) Config {
	if env == nil {
		env = environ()
	}

	baseURL := trimTrailingSlash(getenv(env, "MCP_BASE_URL", "http://localhost:3000"))
	keycloakBaseURL := trimTrailingSlash(getenv(env, "KEYCLOAK_BASE_URL", "http://localhost:8080"))
	realmName := getenv(env, "KEYCLOAK_REALM", "mcp-demo")
	issuer := getenv(env, "KEYCLOAK_ISSUER", keycloakBaseURL+"/realms/"+realmName)
	port, err := strconv.Atoi(getenv(env, "PORT", "3000"))
	if err != nil {
		port = 3000
	}

	return Config{
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
