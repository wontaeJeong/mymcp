import type { AppConfig } from "../src/config.ts";
import { createStaticJwksResolver, type JwkResolver } from "../src/auth/jwks.ts";

export const testConfig: AppConfig = {
  port: 3000,
  baseUrl: "http://localhost:3000",
  mcpPath: "/mcp",
  realm: "mcp-demo",
  realmName: "mcp-demo",
  issuer: "http://localhost:8080/realms/mcp-demo",
  jwksUri: "http://localhost:8080/realms/mcp-demo/protocol/openid-connect/certs",
  audience: "mcp-demo-resource",
  corsOrigins: [],
};

function b64url(input: string | Uint8Array): string {
  const bytes = typeof input === "string" ? new TextEncoder().encode(input) : input;
  return Buffer.from(bytes).toString("base64url");
}

export async function keys(): Promise<{ privateKey: CryptoKey; resolver: JwkResolver }> {
  const pair = await crypto.subtle.generateKey(
    { name: "RSASSA-PKCS1-v1_5", modulusLength: 2048, publicExponent: new Uint8Array([1, 0, 1]), hash: "SHA-256" },
    true,
    ["sign", "verify"],
  ) as CryptoKeyPair;
  const jwk = await crypto.subtle.exportKey("jwk", pair.publicKey);
  return { privateKey: pair.privateKey, resolver: createStaticJwksResolver({ keys: [{ ...jwk, kid: "test-key", alg: "RS256" }] }) };
}

export async function token(privateKey: CryptoKey, overrides: Record<string, unknown> = {}): Promise<string> {
  const now = Math.floor(Date.now() / 1000);
  const header = { alg: "RS256", typ: "JWT", kid: "test-key" };
  const payload = {
    sub: "alice-subject",
    iss: testConfig.issuer,
    aud: testConfig.audience,
    iat: now,
    exp: now + 3600,
    scope: "mcp:tools:read mcp:tools:execute",
    email: "alice@example.com",
    preferred_username: "alice",
    groups: [],
    ...overrides,
  };
  const signingInput = `${b64url(JSON.stringify(header))}.${b64url(JSON.stringify(payload))}`;
  const signature = await crypto.subtle.sign("RSASSA-PKCS1-v1_5", privateKey, new TextEncoder().encode(signingInput));
  return `${signingInput}.${b64url(new Uint8Array(signature))}`;
}

export async function request(handler: (request: Request) => Promise<Response>, path: string, init: RequestInit = {}): Promise<Response> {
  return handler(new Request(`http://localhost:3000${path}`, init));
}

export async function json(response: Response): Promise<any> {
  return response.json();
}
