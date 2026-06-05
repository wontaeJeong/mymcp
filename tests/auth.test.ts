import { describe, expect, it } from "vitest";
import request from "supertest";
import { createApp } from "../src/server.js";
import { UnauthorizedError } from "../src/auth/jwks.js";
import { auth, testConfig } from "./helpers.js";

describe("bearer authentication", () => {
  it("returns 401 and WWW-Authenticate when Authorization is missing", async () => {
    const res = await request(createApp(testConfig)).post("/mcp").send({ jsonrpc: "2.0", id: 1, method: "initialize" }).expect(401);

    expect(res.headers["www-authenticate"]).toBe(
      'Bearer realm="mcp-demo", resource_metadata="http://localhost:3000/.well-known/oauth-protected-resource"',
    );
    expect(res.body.error).toBe("unauthorized");
  });

  it("rejects a token with the wrong issuer", async () => {
    const app = createApp(testConfig, async () => {
      throw new UnauthorizedError("unexpected iss claim value");
    });

    const res = await request(app)
      .post("/mcp")
      .set("authorization", "Bearer invalid-issuer")
      .send({ jsonrpc: "2.0", id: 1, method: "initialize" })
      .expect(401);

    expect(res.body.error).toBe("invalid_token");
    expect(res.body.error_description).toContain("iss");
  });

  it("rejects a token with the wrong audience", async () => {
    const app = createApp(testConfig, async () => {
      throw new UnauthorizedError("Invalid audience. Expected one of: mcp-demo-resource, http://localhost:3000/mcp");
    });

    const res = await request(app)
      .post("/mcp")
      .set("authorization", "Bearer invalid-audience")
      .send({ jsonrpc: "2.0", id: 1, method: "initialize" })
      .expect(401);

    expect(res.body.error).toBe("invalid_token");
    expect(res.body.error_description).toContain("Invalid audience");
  });

  it("allows initialize with any valid Keycloak access token", async () => {
    const app = createApp(testConfig, async () => auth([]));
    const res = await request(app)
      .post("/mcp")
      .set("authorization", "Bearer valid")
      .send({ jsonrpc: "2.0", id: 1, method: "initialize" })
      .expect(200);

    expect(res.body.result.serverInfo.name).toBe("keycloak-google-oauth-mcp-demo");
  });
});
