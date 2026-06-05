# Remote MCP OAuth demo: Keycloak broker + Google upstream IdP

This repository is a local, runnable TypeScript demo for a **remote MCP server protected by OAuth**.

The important boundary is:

- Google OAuth is **only an upstream login provider**.
- Keycloak is the **OAuth authorization server / auth broker** for the MCP resource.
- The MCP server is an **OAuth protected resource / resource server**.
- The MCP server trusts only **Keycloak-issued MCP access tokens**. It does not trust Google access tokens directly.

Local demos work without Google credentials by logging in with built-in Keycloak users.

## Architecture

```text
opencode or another remote MCP client
  -> POST http://localhost:3000/mcp
  -> receives 401 WWW-Authenticate with protected resource metadata
  -> discovers http://localhost:8080/realms/mcp-demo
  -> performs OAuth login with Keycloak
  -> Keycloak may broker Google login, or local alice/admin-user login
  -> client retries /mcp with a Keycloak access token
  -> MCP server verifies iss, aud, exp, scope, and tool policy
```

## Prerequisites

- Node.js 22+
- Docker Compose
- `jq` for the shell examples

## Quick start

```bash
cp .env.example .env
docker compose up -d
npm install
npm run dev
```

Or use Make targets:

```bash
cp .env.example .env
make install
make up
```

`make up` starts Keycloak in Docker and then runs the MCP server in watch mode.

## Endpoints

- MCP endpoint: `POST http://localhost:3000/mcp`
- OAuth Protected Resource Metadata:
  - `GET http://localhost:3000/.well-known/oauth-protected-resource`
  - `GET http://localhost:3000/.well-known/oauth-protected-resource/mcp`
- Keycloak issuer: `http://localhost:8080/realms/mcp-demo`
- Keycloak JWKS: `http://localhost:8080/realms/mcp-demo/protocol/openid-connect/certs`

## Verify metadata

```bash
curl http://localhost:3000/.well-known/oauth-protected-resource | jq
```

Expected shape:

```json
{
  "resource": "http://localhost:3000/mcp",
  "authorization_servers": ["http://localhost:8080/realms/mcp-demo"],
  "scopes_supported": ["mcp:tools:read", "mcp:tools:execute", "mcp:admin"],
  "bearer_methods_supported": ["header"]
}
```

## Verify unauthenticated MCP requests return 401

```bash
curl -i http://localhost:3000/mcp
```

Expected header:

```http
WWW-Authenticate: Bearer realm="mcp-demo", resource_metadata="http://localhost:3000/.well-known/oauth-protected-resource"
```

## Local test tokens

The realm includes `mcp-demo-test`, a public client with password grant enabled solely for local reproducibility.

> Production warning: Resource Owner Password Credentials / password grant must not be used in production. Use authorization code + PKCE for public clients.

### Alice token: read + execute

```bash
export ALICE_TOKEN=$(curl -s -X POST http://localhost:8080/realms/mcp-demo/protocol/openid-connect/token \
  -H 'content-type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=password' \
  --data-urlencode 'client_id=mcp-demo-test' \
  --data-urlencode 'username=alice' \
  --data-urlencode 'password=alice' \
  --data-urlencode 'scope=openid profile email mcp:tools:read mcp:tools:execute' | jq -r .access_token)
```

### Admin token: read + execute + admin

```bash
export ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/realms/mcp-demo/protocol/openid-connect/token \
  -H 'content-type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=password' \
  --data-urlencode 'client_id=mcp-demo-test' \
  --data-urlencode 'username=admin-user' \
  --data-urlencode 'password=admin-user' \
  --data-urlencode 'scope=openid profile email mcp:tools:read mcp:tools:execute mcp:admin' | jq -r .access_token)
```

## Curl MCP examples

Initialize:

```bash
curl -s http://localhost:3000/mcp \
  -H "authorization: Bearer $ALICE_TOKEN" \
  -H 'content-type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | jq
```

List tools, requires `mcp:tools:read`:

```bash
curl -s http://localhost:3000/mcp \
  -H "authorization: Bearer $ALICE_TOKEN" \
  -H 'content-type: application/json' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | jq
```

Call `whoami`, requires `mcp:tools:execute`:

```bash
curl -s http://localhost:3000/mcp \
  -H "authorization: Bearer $ALICE_TOKEN" \
  -H 'content-type: application/json' \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"whoami","arguments":{}}}' | jq
```

Call `echo`, requires `mcp:tools:execute`:

```bash
curl -s http://localhost:3000/mcp \
  -H "authorization: Bearer $ALICE_TOKEN" \
  -H 'content-type: application/json' \
  -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hello from MCP"}}}' | jq
```

Try `admin_status` with Alice. It should be rejected:

```bash
curl -i http://localhost:3000/mcp \
  -H "authorization: Bearer $ALICE_TOKEN" \
  -H 'content-type: application/json' \
  -d '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"admin_status","arguments":{}}}'
```

Call `admin_status` with the admin token:

```bash
curl -s http://localhost:3000/mcp \
  -H "authorization: Bearer $ADMIN_TOKEN" \
  -H 'content-type: application/json' \
  -d '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"admin_status","arguments":{}}}' | jq
```

## Token validation rules

The MCP server validates Keycloak access tokens with `jose` and Keycloak JWKS.

Required claims and policy:

- `iss` must be `http://localhost:8080/realms/mcp-demo`.
- `aud` must include one configured MCP resource audience. This demo accepts either:
  - `mcp-demo-resource`
  - `http://localhost:3000/mcp`
- `exp` must be present and unexpired. `jose` enforces expiry during JWT verification.
- `scope` must include:
  - `mcp:tools:read` for `tools/list`
  - `mcp:tools:execute` for `tools/call`
  - `mcp:admin` or Keycloak group `admin` for `admin_status`
- `email`, `preferred_username`, and `groups` are copied into request context and are visible via `whoami`.

`MCP_REQUIRED_AUDIENCES` can be changed in `.env` if you want to accept only one audience value.

## opencode remote MCP configuration

Example `opencode` config:

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

Authenticate and inspect:

```bash
opencode mcp auth mcp-oauth-demo
opencode mcp list
opencode mcp logout mcp-oauth-demo
```

### Redirect URI warning

opencode versions may use different OAuth callback redirect URIs. If login fails with a redirect URI error:

1. Open the Keycloak Admin Console at `http://localhost:8080/admin`.
2. Login as `admin` / `admin`.
3. Check realm `mcp-demo` events, or inspect the browser URL during the failed authorization request.
4. Copy the exact `redirect_uri` value.
5. Add that value to client `opencode-mcp-demo` → **Valid Redirect URIs**.

The imported realm allows broad localhost redirect URIs for convenience. Avoid wildcard redirect URIs outside a local demo.

## Google upstream Identity Provider setup

The realm export includes a disabled Google IdP placeholder. Real secrets are not committed.

1. Create or select a project in Google Cloud Console.
2. Configure the OAuth consent screen.
3. Create an OAuth 2.0 Client ID for a Web application.
4. Add this Keycloak broker redirect URI to Google:

   ```text
   http://localhost:8080/realms/mcp-demo/broker/google/endpoint
   ```

5. Put credentials in `.env`:

   ```bash
   GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=your-google-client-secret
   ```

6. Restart Keycloak and re-import the realm if needed:

   ```bash
   docker compose down
   docker compose up -d
   ```

7. In Keycloak Admin Console, realm `mcp-demo` → Identity Providers → Google:
   - Enable the provider.
   - Confirm Client ID and Client Secret.
   - Save.

### Restricting Google Workspace domains

If only a Google Workspace domain should be allowed, do not enforce that in the MCP server by trusting Google tokens. Enforce it in Keycloak, for example:

- Configure the Google IdP with hosted-domain behavior where appropriate.
- Add a Keycloak IdP mapper or post-broker login flow that checks the Google `hd` claim.
- Map only approved users/groups/scopes into the Keycloak access token consumed by the MCP server.

## Security notes

- Do not trust a Google access token as the MCP resource token.
- Trust only Keycloak-issued access tokens intended for this MCP resource.
- Validate `iss`, `aud`, `exp`, and `scope` before serving MCP tools.
- Do not use an ID token as an API bearer token.
- Do not pass a user's downstream access token through to unrelated APIs.
- Do not put a client secret in public CLI clients.
- Use authorization code + PKCE for production public clients.
- Restrict CORS to known origins in production.
- Never log access tokens, refresh tokens, or authorization codes.

## Known limitations

- This demo implements a small JSON-RPC-over-HTTP MCP subset sufficient for `initialize`, `tools/list`, and `tools/call`. The `@modelcontextprotocol/sdk` dependency is included for MCP protocol compatibility work, but the server intentionally keeps the HTTP handler minimal so the OAuth protected-resource flow is easy to inspect. Replace this handler with the SDK Streamable HTTP transport when your target MCP client and SDK version agree on the production transport behavior you need.
- The imported Keycloak realm uses broad localhost redirect URI patterns for local testing. Tighten these before sharing an environment.
- The `mcp-demo-test` password-grant client exists only for curl-based local testing.
- Keycloak's Google IdP is disabled until real credentials are supplied and the provider is enabled.

## Production checklist

- Replace password-grant examples with authorization code + PKCE only.
- Pin and regularly update container images and npm dependencies.
- Serve Keycloak and the MCP server over HTTPS.
- Use exact redirect URIs; remove localhost wildcards.
- Restrict CORS to required origins.
- Rotate secrets with a real secret manager.
- Review Keycloak mappers so only intended audiences, scopes, and groups enter access tokens.
- Monitor authentication events without logging tokens or authorization codes.
