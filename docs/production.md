# Production Notes

This demo keeps the application small while separating internal packages for production growth.

Before production use, replace local wildcard redirect URIs, constrain scope issuance, configure explicit CORS origins, and use Authorization Code + PKCE rather than password grants.

## Corporate network, proxy, and CA trust

This application is expected to be developed and built from inside a corporate network, but served only on the internal network in production.

- Development and image-build steps accept `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY` (plus lowercase equivalents) so package installation and Go module download can pass through the corporate proxy.
- Do not configure proxy variables on the runtime MCP service unless the deployment platform requires outbound proxying. The Docker Compose `mcp` service intentionally passes proxy values only as build arguments, not runtime environment variables.
- Place the approved corporate public CA certificate at `certs/internal-ca.crt` when internal TLS or TLS inspection is required. Additional `.crt` files may also be added under `certs/`.
- Docker builds merge `certs/*.crt` into the image CA bundle before dependency download, then copy the resulting CA bundle into the final scratch runtime image so the service can validate internal TLS endpoints such as Keycloak JWKS.
- Local Make targets set `SSL_CERT_FILE` automatically when `certs/internal-ca.crt` exists, allowing Go commands to trust the same corporate CA during development.

Never commit proxy credentials, private keys, or secret certificate material. Only commit public CA certificates approved for distribution with the service image.
