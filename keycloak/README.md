# Keycloak realm notes

`realm-export.json` imports the `mcp-demo` realm used by the MCP resource server.

## Included clients

- `opencode-mcp-demo`: public authorization-code client for remote MCP clients such as opencode. It intentionally has no client secret.
- `mcp-demo-test`: public local test client with Direct Access Grants enabled so README curl examples can mint a token. This is for local demo use only.

## Included users

- `alice` / `alice`: local demo user. Request `mcp:tools:read` and `mcp:tools:execute` scopes for normal tool access.
- `admin-user` / `admin-user`: local demo admin user in the `admin` group. Request `mcp:admin` when you want the admin scope as well.

## Google upstream IdP

The Google IdP is imported disabled by default because real Google credentials must not be committed. Set `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET`, restart/import the realm, and enable the `google` IdP in the Admin Console.
