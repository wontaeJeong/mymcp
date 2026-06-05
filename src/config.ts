import dotenv from "dotenv";

dotenv.config();

export type AppConfig = {
  port: number;
  baseUrl: string;
  resourcePath: string;
  resource: string;
  realm: string;
  authorizationServers: string[];
  scopesSupported: string[];
  issuer: string;
  jwksUri: string;
  audiences: string[];
  corsOrigins: string[];
};

const splitCsv = (value: string | undefined, fallback: string[]): string[] =>
  (value ?? fallback.join(","))
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);

export function loadConfig(env: NodeJS.ProcessEnv = process.env): AppConfig {
  const port = Number(env.PORT ?? 3000);
  const baseUrl = env.MCP_BASE_URL ?? `http://localhost:${port}`;
  const resourcePath = env.MCP_RESOURCE_PATH ?? "/mcp";
  const keycloakRealm = env.KEYCLOAK_REALM ?? "mcp-demo";
  const keycloakBaseUrl = env.KEYCLOAK_BASE_URL ?? "http://localhost:8080";
  const issuer = env.KEYCLOAK_ISSUER ?? `${keycloakBaseUrl}/realms/${keycloakRealm}`;
  const resourceUrl = `${baseUrl}${resourcePath}`;

  return {
    port,
    baseUrl,
    resourcePath,
    resource: env.MCP_RESOURCE_ID ?? "mcp-demo-resource",
    realm: env.MCP_REALM ?? keycloakRealm,
    authorizationServers: splitCsv(env.AUTHORIZATION_SERVERS, [issuer]),
    scopesSupported: ["mcp:tools:read", "mcp:tools:execute", "mcp:admin"],
    issuer,
    jwksUri: env.KEYCLOAK_JWKS_URI ?? `${issuer}/protocol/openid-connect/certs`,
    audiences: splitCsv(env.MCP_REQUIRED_AUDIENCES, [env.MCP_RESOURCE_ID ?? "mcp-demo-resource", resourceUrl]),
    corsOrigins: splitCsv(env.CORS_ORIGINS, []),
  };
}
