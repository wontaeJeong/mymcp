import { describe, expect, it } from "vitest";
import request from "supertest";
import { createApp } from "../src/server.js";
import { testConfig } from "./helpers.js";

describe("protected resource metadata", () => {
  it("returns authorization servers and supported scopes", async () => {
    const res = await request(createApp(testConfig)).get("/.well-known/oauth-protected-resource").expect(200);

    expect(res.body).toMatchObject({
      resource: "http://localhost:3000/mcp",
      authorization_servers: ["http://localhost:8080/realms/mcp-demo"],
      scopes_supported: ["mcp:tools:read", "mcp:tools:execute", "mcp:admin"],
      bearer_methods_supported: ["header"],
    });
  });

  it("also serves the MCP-specific metadata path", async () => {
    const res = await request(createApp(testConfig)).get("/.well-known/oauth-protected-resource/mcp").expect(200);
    expect(res.body.resource).toBe("http://localhost:3000/mcp");
  });
});
