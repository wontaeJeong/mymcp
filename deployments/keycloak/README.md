# Keycloak realm

`realm-export.json` imports the `mcp-demo` realm for the local Docker Compose demo.

Default console account: `admin` / `admin` at <http://localhost:8080/admin>.

Local users:

- `alice` / `alice` (`alice@example.com`)
- `admin-user` / `admin-user` (`admin@example.com`, member of `/admin`)

Clients:

- `opencode-mcp-demo`: public browser/CLI client for opencode Authorization Code + PKCE.
- `mcp-demo-cli`: public local test client with Direct Access Grants enabled for curl demos only.
- `mcp-demo-test`: public local automated test client.

The realm includes optional client scopes named `mcp:tools:read`, `mcp:tools:execute`, and `mcp:admin`, an audience mapper that emits `aud=mcp-demo-resource`, and a group mapper that emits the `groups` claim.

The optional upstream identity provider is a disabled generic OIDC broker named `internal-oidc`. Configure it with `INTERNAL_OIDC_*` environment variables, then enable it in the Keycloak admin console. No external SaaS provider is required for the local demo.
