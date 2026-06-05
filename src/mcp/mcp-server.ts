import type { Request, Response } from "express";
import { ForbiddenError, requireScope } from "../auth/policy.js";
import { asMcpContent, callTool, tools } from "./tools.js";

type JsonRpcRequest = {
  jsonrpc?: "2.0";
  id?: string | number | null;
  method?: string;
  params?: Record<string, unknown>;
};

function result(id: JsonRpcRequest["id"], value: unknown) {
  return { jsonrpc: "2.0", id: id ?? null, result: value };
}

function error(id: JsonRpcRequest["id"], code: number, message: string) {
  return { jsonrpc: "2.0", id: id ?? null, error: { code, message } };
}

export async function handleMcpRequest(req: Request, res: Response): Promise<void> {
  const auth = req.auth;
  if (!auth) {
    res.status(500).json(error(null, -32603, "Authentication context missing"));
    return;
  }

  const body = req.body as JsonRpcRequest;
  if (!body || body.jsonrpc !== "2.0" || typeof body.method !== "string") {
    res.status(400).json(error(null, -32600, "Invalid JSON-RPC 2.0 request"));
    return;
  }

  try {
    switch (body.method) {
      case "initialize":
        res.json(result(body.id, {
          protocolVersion: "2025-03-26",
          capabilities: { tools: {} },
          serverInfo: { name: "keycloak-google-oauth-mcp-demo", version: "0.1.0" },
        }));
        return;
      case "notifications/initialized":
        res.status(202).end();
        return;
      case "tools/list":
        requireScope(auth, "mcp:tools:read");
        res.json(result(body.id, { tools }));
        return;
      case "tools/call": {
        requireScope(auth, "mcp:tools:execute");
        const params = body.params ?? {};
        const name = typeof params.name === "string" ? params.name : undefined;
        if (!name) {
          res.status(400).json(error(body.id, -32602, "tools/call requires params.name"));
          return;
        }
        const output = await callTool(name, params.arguments, auth);
        res.json(result(body.id, { content: asMcpContent(output) }));
        return;
      }
      default:
        res.status(404).json(error(body.id, -32601, `Method not found: ${body.method}`));
    }
  } catch (caught) {
    if (caught instanceof ForbiddenError) {
      res.status(403).json(error(body.id, -32001, caught.message));
      return;
    }
    res.status(400).json(error(body.id, -32602, caught instanceof Error ? caught.message : "Invalid request"));
  }
}
