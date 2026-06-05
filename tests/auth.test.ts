import { describe, expect, it } from "bun:test";
import { createFetchHandler } from "../src/server.ts";
import { json, keys, request, testConfig, token } from "./helpers.ts";

describe("bearer auth", () => {
  it("returns 401 and WWW-Authenticate when Authorization is missing", async () => {
    const { resolver } = await keys();
    const res = await request(createFetchHandler(testConfig, resolver), "/mcp");

    expect(res.status).toBe(401);
    expect(res.headers.get("WWW-Authenticate")).toContain("Bearer realm=\"mcp-demo\"");
    expect(res.headers.get("WWW-Authenticate")).toContain("resource_metadata=\"http://localhost:3000/.well-known/oauth-protected-resource\"");
  });

  it("rejects tokens with the wrong issuer", async () => {
    const { privateKey, resolver } = await keys();
    const bad = await token(privateKey, { iss: "http://localhost:8080/realms/other" });
    const res = await request(createFetchHandler(testConfig, resolver), "/mcp", {
      method: "POST",
      headers: { Authorization: `Bearer ${bad}`, "Content-Type": "application/json" },
      body: JSON.stringify({ jsonrpc: "2.0", id: 1, method: "tools/list" }),
    });

    expect(res.status).toBe(401);
    expect((await json(res)).error).toBe("invalid_token");
  });

  it("rejects tokens with the wrong audience", async () => {
    const { privateKey, resolver } = await keys();
    const bad = await token(privateKey, { aud: "wrong-audience" });
    const res = await request(createFetchHandler(testConfig, resolver), "/mcp", {
      method: "POST",
      headers: { Authorization: `Bearer ${bad}`, "Content-Type": "application/json" },
      body: JSON.stringify({ jsonrpc: "2.0", id: 1, method: "tools/list" }),
    });

    expect(res.status).toBe(401);
    expect((await json(res)).error).toBe("invalid_token");
  });
});
