export type AppConfig = {
  port: number;
  baseUrl: string;
  mcpPath: string;
  realm: string;
  issuer: string;
  jwksUri: string;
  audience: string;
  realmName: string;
  corsOrigins: string[];
};

function trimTrailingSlash(value: string): string {
  return value.replace(/\/$/, "");
}

function csv(value: string | undefined): string[] {
  return value?.split(",").map((entry) => entry.trim()).filter(Boolean) ?? [];
}

export function loadConfig(env: Record<string, string | undefined> = process.env): AppConfig {
  const baseUrl = trimTrailingSlash(env.MCP_BASE_URL ?? "http://localhost:3000");
  const keycloakBaseUrl = trimTrailingSlash(env.KEYCLOAK_BASE_URL ?? "http://localhost:8080");
  const realm = env.KEYCLOAK_REALM ?? "mcp-demo";
  const issuer = env.KEYCLOAK_ISSUER ?? `${keycloakBaseUrl}/realms/${realm}`;

  return {
    port: Number(env.PORT ?? 3000),
    baseUrl,
    mcpPath: env.MCP_PATH ?? "/mcp",
    realm: env.AUTH_REALM ?? "mcp-demo",
    realmName: realm,
    issuer,
    jwksUri: env.KEYCLOAK_JWKS_URI ?? `${issuer}/protocol/openid-connect/certs`,
    audience: env.MCP_AUDIENCE ?? "mcp-demo-resource",
    corsOrigins: csv(env.CORS_ORIGINS),
  };
}

export const config = loadConfig();
