import { describe, expect, it } from "bun:test";
import { createFetchHandler } from "../src/server.ts";
import { json, keys, request, testConfig, token } from "./helpers.ts";

function rpc(method: string, params?: Record<string, unknown>) {
  return { jsonrpc: "2.0", id: 1, method, params };
}

async function postMcp(scope: string, body: unknown, groups: string[] = []) {
  const { privateKey, resolver } = await keys();
  const accessToken = await token(privateKey, { scope, groups });
  return request(createFetchHandler(testConfig, resolver), "/mcp", {
    method: "POST",
    headers: { Authorization: `Bearer ${accessToken}`, "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}

describe("MCP tool authorization policy", () => {
  it("denies tool calls when mcp:tools:execute scope is missing", async () => {
    const res = await postMcp("mcp:tools:read", rpc("tools/call", { name: "echo", arguments: { message: "hello" } }));
    expect(res.status).toBe(403);
    expect((await json(res)).error.message).toContain("mcp:tools:execute");
  });

  it("allows tools/list with mcp:tools:read scope", async () => {
    const res = await postMcp("mcp:tools:read", rpc("tools/list"));
    const body = await json(res);
    expect(res.status).toBe(200);
    expect(body.result.tools.map((tool: { name: string }) => tool.name)).toEqual(["whoami", "echo", "admin_status"]);
  });

  it("allows echo with mcp:tools:execute scope", async () => {
    const res = await postMcp("mcp:tools:execute", rpc("tools/call", { name: "echo", arguments: { message: "hello" } }));
    expect(res.status).toBe(200);
    expect((await json(res)).result.content[0].text).toContain("hello");
  });

  it("denies admin_status without admin group or mcp:admin scope", async () => {
    const res = await postMcp("mcp:tools:execute", rpc("tools/call", { name: "admin_status", arguments: {} }));
    expect(res.status).toBe(403);
    expect((await json(res)).error.message).toContain("admin");
  });

  it("allows admin_status with admin group", async () => {
    const res = await postMcp("mcp:tools:execute", rpc("tools/call", { name: "admin_status", arguments: {} }), ["admin"]);
    expect(res.status).toBe(200);
    expect((await json(res)).result.content[0].text).toContain("admin");
  });
});
