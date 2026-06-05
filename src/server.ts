import type { JwkResolver } from "./auth/jwks.ts";
import type { AppConfig } from "./config.ts";
import { config as defaultConfig } from "./config.ts";
import { authenticateRequest } from "./auth/bearer-auth.ts";
import { protectedResourceMetadata } from "./auth/protected-resource-metadata.ts";
import { handleMcpJsonRpc } from "./mcp/mcp-server.ts";

function json(body: unknown, init: ResponseInit = {}): Response {
  const headers = new Headers(init.headers);
  headers.set("Content-Type", "application/json");
  return new Response(JSON.stringify(body), { ...init, headers });
}

function addCors(response: Response, request: Request, config: AppConfig): Response {
  if (config.corsOrigins.length === 0) {
    return response;
  }
  const origin = request.headers.get("Origin");
  if (origin && config.corsOrigins.includes(origin)) {
    response.headers.set("Access-Control-Allow-Origin", origin);
    response.headers.set("Access-Control-Allow-Headers", "Authorization, Content-Type");
    response.headers.set("Access-Control-Allow-Methods", "GET, POST, OPTIONS");
  }
  return response;
}

export function createFetchHandler(config: AppConfig = defaultConfig, resolveJwk?: JwkResolver) {
  return async (request: Request): Promise<Response> => {
    const url = new URL(request.url);
    let response: Response;

    if (request.method === "OPTIONS") {
      response = new Response(null, { status: 204 });
      return addCors(response, request, config);
    }

    if (request.method === "GET" && (url.pathname === "/.well-known/oauth-protected-resource" || url.pathname === "/.well-known/oauth-protected-resource/mcp")) {
      response = json(protectedResourceMetadata(config));
      return addCors(response, request, config);
    }

    if (url.pathname === config.mcpPath) {
      const authResult = await authenticateRequest(request, config, resolveJwk);
      if (!authResult.ok) {
        response = json(authResult.body, { status: authResult.status, headers: authResult.headers });
        return addCors(response, request, config);
      }

      if (request.method !== "POST") {
        response = json({ error: "method_not_allowed", message: "Use POST for MCP JSON-RPC requests." }, { status: 405 });
        return addCors(response, request, config);
      }

      const rpc = await request.json().catch(() => undefined);
      if (!rpc || typeof rpc !== "object") {
        response = json({ jsonrpc: "2.0", id: null, error: { code: -32700, message: "Invalid JSON body" } }, { status: 400 });
        return addCors(response, request, config);
      }

      const mcp = await handleMcpJsonRpc(rpc as Record<string, unknown>, authResult.auth);
      response = mcp.body === undefined ? new Response(null, { status: mcp.status }) : json(mcp.body, { status: mcp.status });
      return addCors(response, request, config);
    }

    response = json({ error: "not_found" }, { status: 404 });
    return addCors(response, request, config);
  };
}

if (import.meta.main) {
  Bun.serve({ port: defaultConfig.port, fetch: createFetchHandler() });
  console.log(`MCP OAuth demo listening on ${defaultConfig.baseUrl}${defaultConfig.mcpPath}`);
}
