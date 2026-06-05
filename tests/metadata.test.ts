import { describe, expect, it } from "bun:test";
import { createFetchHandler } from "../src/server.ts";
import { json, request, testConfig } from "./helpers.ts";

describe("protected resource metadata", () => {
  it("returns MCP OAuth protected resource metadata", async () => {
    const res = await request(createFetchHandler(testConfig), "/.well-known/oauth-protected-resource");
    const body = await json(res);

    expect(res.status).toBe(200);
    expect(body).toMatchObject({
      resource: "http://localhost:3000/mcp",
      authorization_servers: ["http://localhost:8080/realms/mcp-demo"],
      scopes_supported: ["mcp:tools:read", "mcp:tools:execute", "mcp:admin"],
    });
  });
});
