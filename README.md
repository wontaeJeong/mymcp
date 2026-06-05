# Remote MCP OAuth demo: opencode → MCP resource server → Keycloak broker → internal OIDC

This repository is a runnable **Go** example of a remote MCP server protected by OAuth 2.0.

The intended production shape is an internal identity platform, not a public SaaS IdP:

```text
opencode / remote MCP client
        │
        │ Bearer access token issued by Keycloak
        ▼
MCP server = OAuth protected resource / resource server
        │
        │ validates Keycloak JWT with JWKS, iss, aud, exp, scope
        ▼
Keycloak = authorization server + auth broker
        │
        │ optional upstream federation
        ▼
Internal OIDC Identity Provider
```

For the local demo, upstream OIDC credentials are optional. Keycloak imports local users so the full resource-server flow works without any external service account.

## What is included

- Go server that uses only the standard library at runtime.
- Minimal HTTP JSON-RPC MCP endpoint at `POST /mcp`.
- OAuth Protected Resource Metadata at:
  - `GET /.well-known/oauth-protected-resource`
  - `GET /.well-known/oauth-protected-resource/mcp`
- Keycloak in Docker Compose with realm import.
- JWT validation using Go crypto libraries and Keycloak JWKS.
- Validation of `iss`, `aud`, `exp`, and tool scopes.
- Tool authorization by scope/group:
  - `tools/list` requires `mcp:tools:read`.
  - `tools/call` requires `mcp:tools:execute`.
  - `admin_status` additionally requires `mcp:admin` scope or `admin` group.
- Go unit tests for metadata, authentication, JWT rejection, and policy enforcement.

## Environment variables

`.env.example` is split into two sections:

### Public / non-secret values

These values may appear in local documentation or client configuration:

- `PORT`
- `MCP_BASE_URL`
- `MCP_PATH`
- `KEYCLOAK_BASE_URL`
- `KEYCLOAK_REALM`
- `KEYCLOAK_ISSUER`
- `KEYCLOAK_JWKS_URI`
- `MCP_AUDIENCE`
- `CORS_ORIGINS`
- `KEYCLOAK_ADMIN` for the local demo account name
- `INTERNAL_OIDC_ISSUER`
- `INTERNAL_OIDC_AUTHORIZATION_URL`
- `INTERNAL_OIDC_TOKEN_URL`
- `INTERNAL_OIDC_USERINFO_URL`
- `INTERNAL_OIDC_JWKS_URL`

### Secret values

Never commit real values for these variables:

- `KEYCLOAK_ADMIN_PASSWORD`
- `INTERNAL_OIDC_CLIENT_ID` if your internal IdP treats client identifiers as confidential
- `INTERNAL_OIDC_CLIENT_SECRET`

The variable names are stable and intentionally descriptive; only the grouping in `.env.example` distinguishes public configuration from secrets.

## Quick start

```bash
cp .env.example .env
docker compose up -d
go mod download
go run .
```

Equivalent Make targets:

```bash
make up
make install
make dev
```

Keycloak starts at <http://localhost:8080>. The admin console account defaults to `admin` / `admin` for local development only.

## Verify metadata

```bash
curl http://localhost:3000/.well-known/oauth-protected-resource
```

Expected shape:

```json
{
  "resource": "http://localhost:3000/mcp",
  "authorization_servers": ["http://localhost:8080/realms/mcp-demo"],
  "scopes_supported": ["mcp:tools:read", "mcp:tools:execute", "mcp:admin"]
}
```

## Verify the unauthenticated challenge

```bash
curl -i http://localhost:3000/mcp
```

Expected response includes `401 Unauthorized` and this header:

```http
WWW-Authenticate: Bearer realm="mcp-demo", resource_metadata="http://localhost:3000/.well-known/oauth-protected-resource"
```

## Local users

The realm import creates these users:

| Username | Password | Email | Intended privileges |
| --- | --- | --- | --- |
| `alice` | `alice` | `alice@example.com` | read/execute demo user |
| `admin-user` | `admin-user` | `admin@example.com` | read/execute/admin demo user, member of `admin` group |

## Get a local test token

The demo includes a public client named `mcp-demo-cli` with Direct Access Grants enabled so curl examples are easy to reproduce.

> **Production warning:** password grant / Direct Access Grants are for this local demo only. Production clients should use Authorization Code + PKCE.

Alice token:

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/realms/mcp-demo/protocol/openid-connect/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'client_id=mcp-demo-cli' \
  -d 'grant_type=password' \
  -d 'username=alice' \
  -d 'password=alice' \
  -d 'scope=openid profile email mcp:tools:read mcp:tools:execute' | jq -r .access_token)
```

Admin token:

```bash
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/realms/mcp-demo/protocol/openid-connect/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'client_id=mcp-demo-cli' \
  -d 'grant_type=password' \
  -d 'username=admin-user' \
  -d 'password=admin-user' \
  -d 'scope=openid profile email mcp:tools:read mcp:tools:execute mcp:admin' | jq -r .access_token)
```

The MCP server expects:

- issuer: `http://localhost:8080/realms/mcp-demo`
- audience: `mcp-demo-resource`
- resource URL: `http://localhost:3000/mcp`

The audience is a stable logical API audience (`mcp-demo-resource`). If your internal platform models resource indicators as URLs, set `MCP_AUDIENCE=http://localhost:3000/mcp` and update the Keycloak audience mapper accordingly.

## Call MCP methods with curl

Initialize:

```bash
curl -s http://localhost:3000/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | jq
```

List tools:

```bash
curl -s http://localhost:3000/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | jq
```

Call `whoami`:

```bash
curl -s http://localhost:3000/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"whoami","arguments":{}}}' | jq
```

Call `echo`:

```bash
curl -s http://localhost:3000/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hello from MCP"}}}' | jq
```

Admin-only tool with a non-admin token should be denied:

```bash
curl -i http://localhost:3000/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"admin_status","arguments":{}}}'
```

Admin token should be allowed:

```bash
curl -s http://localhost:3000/mcp \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"admin_status","arguments":{}}}' | jq
```

## opencode remote MCP configuration

Example opencode config:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "mcp-oauth-demo": {
      "type": "remote",
      "url": "http://localhost:3000/mcp",
      "enabled": true,
      "oauth": {
        "clientId": "opencode-mcp-demo",
        "scope": "openid profile email mcp:tools:read mcp:tools:execute"
      }
    }
  }
}
```

Run:

```bash
opencode mcp auth mcp-oauth-demo
opencode mcp list
opencode mcp logout mcp-oauth-demo
```

### Redirect URI warning

opencode versions may use different local OAuth callback redirect URIs. If login fails with a redirect URI error:

1. Open the Keycloak admin console.
2. Check realm events or inspect the browser URL during the failed authorization request.
3. Copy the exact `redirect_uri` value.
4. Add it to **Clients → opencode-mcp-demo → Valid Redirect URIs**.

The imported realm includes localhost wildcards for local demo convenience. Avoid wildcard redirect URIs outside local development.

## Configure an internal OIDC upstream Identity Provider

The realm import includes a disabled Keycloak broker named `internal-oidc`. To connect it to an internal OIDC system:

1. Register a confidential OIDC client in the internal identity system.
2. Add this Keycloak broker callback URL to the internal IdP's allowed redirect URIs:

   ```text
   http://localhost:8080/realms/mcp-demo/broker/internal-oidc/endpoint
   ```

3. Copy `.env.example` to `.env` and set the `INTERNAL_OIDC_*` values. Keep real client credentials out of git.
4. Restart Keycloak:

   ```bash
   docker compose down
   docker compose up -d
   ```

5. In the Keycloak admin console, go to **Identity providers → internal-oidc**, verify endpoints and credentials, and enable the provider.

If the internal IdP enforces corporate domain, department, device posture, or network claims, enforce those conditions in the internal IdP and/or Keycloak broker mappers/flows. Do not rely on client-side hints.

## Security notes

- The MCP server trusts only Keycloak-issued MCP access tokens.
- The MCP server does **not** trust upstream internal OIDC access tokens directly as resource-server tokens.
- The MCP server validates `iss`, `aud`, `exp`, and tool scopes.
- Do not use ID tokens as API bearer tokens.
- Do not pass a user's downstream access token through to unrelated APIs without an explicit token exchange/delegation design.
- Public CLI clients must not contain client secrets.
- Use Authorization Code + PKCE in production.
- Configure CORS with explicit origins via `CORS_ORIGINS`; do not use broad origins in production.
- Do not log access tokens, refresh tokens, or authorization codes.

## Known limitations

- This server intentionally implements the minimum HTTP JSON-RPC behavior needed for local MCP `initialize`, `tools/list`, and `tools/call` demonstrations. It does not expose a full Streamable HTTP transport lifecycle.
- The password grant examples are for local reproducibility only and should be disabled in production.
- The realm import includes localhost wildcard redirect URIs for opencode discovery convenience. Replace them with exact callback URIs outside local demos.
- Keycloak client scopes in the local demo make requested MCP scopes appear in the access token. For production, constrain scope issuance with client policies, role mappers, consent, and/or authorization services that match your authorization model.

## Tests

```bash
go test ./...
go build ./...
```
