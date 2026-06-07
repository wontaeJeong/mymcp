# OAuth and OIDC Notes

The MCP server is an OAuth protected resource. It validates Keycloak-issued access tokens with JWKS, issuer, audience, expiration, and MCP tool scopes.

For local development, the Keycloak realm import lives under `deployments/keycloak/` and includes demo users plus MCP-specific client scopes.
