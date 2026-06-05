import type { AuthContext } from "../auth/bearer-auth.ts";
import { canCallTools, canListTools } from "../auth/policy.ts";
import { callTool, tools } from "./tools.ts";

type JsonRpcRequest = {
  jsonrpc?: string;
  id?: string | number | null;
  method?: string;
  params?: Record<string, unknown>;
};

export type McpResponse = {
  status: number;
  body?: unknown;
};

function result(id: JsonRpcRequest["id"], value: unknown) {
  return { jsonrpc: "2.0", id: id ?? null, result: value };
}

function error(id: JsonRpcRequest["id"], code: number, message: string) {
  return { jsonrpc: "2.0", id: id ?? null, error: { code, message } };
}

function toolContent(value: unknown) {
  return { content: [{ type: "text", text: JSON.stringify(value, null, 2) }] };
}

export async function handleMcpJsonRpc(rpc: JsonRpcRequest, auth: AuthContext): Promise<McpResponse> {
  if (rpc.method === "initialize") {
    return {
      status: 200,
      body: result(rpc.id, {
        protocolVersion: "2025-03-26",
        capabilities: { tools: {} },
        serverInfo: { name: "keycloak-internal-oidc-mcp-demo", version: "0.2.0" },
      }),
    };
  }

  if (rpc.method === "notifications/initialized") {
    return { status: 202 };
  }

  if (rpc.method === "tools/list") {
    if (!canListTools(auth)) {
      return { status: 403, body: error(rpc.id, -32001, "tools/list requires mcp:tools:read scope") };
    }
    return { status: 200, body: result(rpc.id, { tools }) };
  }

  if (rpc.method === "tools/call") {
    if (!canCallTools(auth)) {
      return { status: 403, body: error(rpc.id, -32001, "tools/call requires mcp:tools:execute scope") };
    }

    const name = typeof rpc.params?.name === "string" ? rpc.params.name : undefined;
    const args = rpc.params?.arguments ?? {};
    if (!name) {
      return { status: 400, body: error(rpc.id, -32602, "tools/call requires params.name") };
    }

    try {
      return { status: 200, body: result(rpc.id, toolContent(callTool(name, args, auth))) };
    } catch (err) {
      const caught = err as Error;
      const status = caught.name === "ForbiddenToolError" ? 403 : 400;
      return { status, body: error(rpc.id, status === 403 ? -32001 : -32602, caught.message) };
    }
  }

  return { status: 404, body: error(rpc.id, -32601, `Unsupported MCP method: ${rpc.method ?? "<missing>"}`) };
}
