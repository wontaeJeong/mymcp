import type { AppConfig } from "../config.ts";

export type ProtectedResourceMetadata = {
  resource: string;
  authorization_servers: string[];
  scopes_supported: string[];
  bearer_methods_supported: string[];
};

export const MCP_SCOPES = ["mcp:tools:read", "mcp:tools:execute", "mcp:admin"] as const;

export function protectedResourceMetadata(config: AppConfig): ProtectedResourceMetadata {
  return {
    resource: `${config.baseUrl}${config.mcpPath}`,
    authorization_servers: [config.issuer],
    scopes_supported: [...MCP_SCOPES],
    bearer_methods_supported: ["header"],
  };
}

export function wwwAuthenticateHeader(config: AppConfig): string {
  return `Bearer realm="${config.realm}", resource_metadata="${config.baseUrl}/.well-known/oauth-protected-resource"`;
}
