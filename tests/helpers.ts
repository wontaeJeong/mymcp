import type { AppConfig } from "../src/config.js";
import type { AuthContext } from "../src/auth/policy.js";

export const testConfig: AppConfig = {
  port: 3000,
  baseUrl: "http://localhost:3000",
  resourcePath: "/mcp",
  resource: "mcp-demo-resource",
  realm: "mcp-demo",
  authorizationServers: ["http://localhost:8080/realms/mcp-demo"],
  scopesSupported: ["mcp:tools:read", "mcp:tools:execute", "mcp:admin"],
  issuer: "http://localhost:8080/realms/mcp-demo",
  jwksUri: "http://localhost:8080/realms/mcp-demo/protocol/openid-connect/certs",
  audiences: ["mcp-demo-resource", "http://localhost:3000/mcp"],
  corsOrigins: [],
};

export function auth(scopes: string[], groups: string[] = []): AuthContext {
  return {
    subject: "user-123",
    issuer: testConfig.issuer,
    audience: ["mcp-demo-resource"],
    scopes,
    groups,
    email: "alice@example.com",
    preferredUsername: "alice",
    expiresAt: Math.floor(Date.now() / 1000) + 3600,
  };
}
