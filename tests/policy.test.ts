import { describe, expect, it } from "vitest";
import request from "supertest";
import { createApp } from "../src/server.js";
import { auth, testConfig } from "./helpers.js";

describe("MCP tool authorization policy", () => {
  it("denies tool calls without the execute scope", async () => {
    const app = createApp(testConfig, async () => auth(["mcp:tools:read"]));
    const res = await request(app)
      .post("/mcp")
      .set("authorization", "Bearer read-only")
      .send({ jsonrpc: "2.0", id: 1, method: "tools/call", params: { name: "echo", arguments: { message: "hi" } } })
      .expect(403);

    expect(res.body.error.message).toContain("mcp:tools:execute");
  });

  it("allows tools/list with mcp:tools:read", async () => {
    const app = createApp(testConfig, async () => auth(["mcp:tools:read"]));
    const res = await request(app)
      .post("/mcp")
      .set("authorization", "Bearer read")
      .send({ jsonrpc: "2.0", id: 1, method: "tools/list" })
      .expect(200);

    expect(res.body.result.tools.map((tool: { name: string }) => tool.name)).toEqual(["whoami", "echo", "admin_status"]);
  });

  it("allows echo with mcp:tools:execute", async () => {
    const app = createApp(testConfig, async () => auth(["mcp:tools:execute"]));
    const res = await request(app)
      .post("/mcp")
      .set("authorization", "Bearer execute")
      .send({ jsonrpc: "2.0", id: 1, method: "tools/call", params: { name: "echo", arguments: { message: "hello" } } })
      .expect(200);

    expect(res.body.result.content[0].text).toContain("hello");
  });

  it("denies admin_status without admin group or mcp:admin scope", async () => {
    const app = createApp(testConfig, async () => auth(["mcp:tools:execute"]));
    const res = await request(app)
      .post("/mcp")
      .set("authorization", "Bearer execute")
      .send({ jsonrpc: "2.0", id: 1, method: "tools/call", params: { name: "admin_status", arguments: {} } })
      .expect(403);

    expect(res.body.error.message).toContain("admin");
  });

  it("allows admin_status for the admin group", async () => {
    const app = createApp(testConfig, async () => auth(["mcp:tools:execute"], ["admin"]));
    const res = await request(app)
      .post("/mcp")
      .set("authorization", "Bearer admin")
      .send({ jsonrpc: "2.0", id: 1, method: "tools/call", params: { name: "admin_status", arguments: {} } })
      .expect(200);

    expect(res.body.result.content[0].text).toContain("admin access");
  });
});
