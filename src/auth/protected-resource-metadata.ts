import type { AppConfig } from "../config.js";

export type OAuthProtectedResourceMetadata = {
  resource: string;
  authorization_servers: string[];
  scopes_supported: string[];
  bearer_methods_supported: string[];
  resource_documentation?: string;
};

export function protectedResourceMetadata(config: AppConfig): OAuthProtectedResourceMetadata {
  return {
    resource: `${config.baseUrl}${config.resourcePath}`,
    authorization_servers: config.authorizationServers,
    scopes_supported: config.scopesSupported,
    bearer_methods_supported: ["header"],
    resource_documentation: `${config.baseUrl}/README.md`,
  };
}
