# MVP Demo Script

This script is a 5-10 minute walkthrough for the local remote MCP OAuth demo.

## 1. Start Keycloak

```bash
cp .env.example .env
docker compose -f deployments/docker-compose.yml up -d
```

Wait until Keycloak is available at <http://localhost:8080>. The local admin account is `admin` / `admin`.

## 2. Start the MCP Server

In a second terminal:

```bash
go run ./cmd/mymcp
```

The server listens on <http://localhost:3000/mcp> by default.

## 3. Verify OAuth Protected Resource Metadata

```bash
curl -s http://localhost:3000/.well-known/oauth-protected-resource | jq
```

Expected highlights:

- `resource` is `http://localhost:3000/mcp`.
- `authorization_servers` contains `http://localhost:8080/realms/mcp-demo`.
- `scopes_supported` contains `mcp:tools:read`, `mcp:tools:execute`, and `mcp:admin`.

## 4. Verify the Unauthenticated Challenge

```bash
curl -i http://localhost:3000/mcp
```

Expected highlights:

- Status is `401 Unauthorized`.
- `WWW-Authenticate` is `Bearer realm="mcp-demo", resource_metadata="http://localhost:3000/.well-known/oauth-protected-resource"`.

## 5. Issue an Alice Token

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/realms/mcp-demo/protocol/openid-connect/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'client_id=mcp-demo-cli' \
  -d 'grant_type=password' \
  -d 'username=alice' \
  -d 'password=alice' \
  -d 'scope=mcp:tools:read mcp:tools:execute' | jq -r .access_token)
```

## 6. Initialize the MCP Session

```bash
curl -s http://localhost:3000/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | jq
```

Expected highlights: protocol version, tool capabilities, and server info.

## 7. Call `tools/list`

```bash
curl -s http://localhost:3000/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | jq
```

Expected tools: `whoami`, `echo`, and `admin_status`.

## 8. Call `whoami`

```bash
curl -s http://localhost:3000/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"whoami","arguments":{}}}' | jq
```

Expected highlights: the MCP scopes attached to Alice's token. User identity fields appear when the Keycloak token includes those claims.

## 9. Call `echo`

```bash
curl -s http://localhost:3000/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hello from MCP"}}}' | jq
```

Expected result content includes `"message": "hello from MCP"`.

## 10. Verify Alice Cannot Call `admin_status`

```bash
curl -i http://localhost:3000/mcp \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"admin_status","arguments":{}}}'
```

Expected highlights:

- Status is `403 Forbidden`.
- The JSON-RPC error mentions the required admin group or `mcp:admin` scope.

## 11. Issue an Admin Token

```bash
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/realms/mcp-demo/protocol/openid-connect/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'client_id=mcp-demo-cli' \
  -d 'grant_type=password' \
  -d 'username=admin-user' \
  -d 'password=admin-user' \
  -d 'scope=mcp:tools:read mcp:tools:execute mcp:admin' | jq -r .access_token)
```

## 12. Call `admin_status`

```bash
curl -s http://localhost:3000/mcp \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"admin_status","arguments":{}}}' | jq
```

Expected result content includes `"status": "ok"` and `"admin": true`.

## 13. Stop the Demo

Stop the MCP server with `Ctrl-C`, then stop Keycloak:

```bash
docker compose -f deployments/docker-compose.yml down
```
