import type { AppConfig } from "../config.ts";
import { wwwAuthenticateHeader } from "./protected-resource-metadata.ts";
import { groupsFromClaim, scopesFromClaim } from "./policy.ts";
import { createRemoteJwksResolver, type JwkResolver } from "./jwks.ts";

export type AuthContext = {
  subject: string;
  email?: string;
  preferredUsername?: string;
  groups: string[];
  scopes: string[];
  claims: Record<string, unknown>;
};

export type AuthResult =
  | { ok: true; auth: AuthContext }
  | { ok: false; status: 401; headers: Record<string, string>; body: { error: string } };

export function bearerTokenFromHeader(header: string | null | undefined): string | undefined {
  if (!header) {
    return undefined;
  }
  const match = /^Bearer\s+(.+)$/i.exec(header.trim());
  return match?.[1];
}

function decodeBase64UrlJson(segment: string): Record<string, unknown> {
  return JSON.parse(Buffer.from(segment, "base64url").toString("utf8")) as Record<string, unknown>;
}

function audienceMatches(aud: unknown, expected: string): boolean {
  return typeof aud === "string" ? aud === expected : Array.isArray(aud) && aud.includes(expected);
}

export async function verifyAccessToken(token: string, config: AppConfig, resolveJwk: JwkResolver): Promise<AuthContext> {
  const parts = token.split(".");
  if (parts.length !== 3) {
    throw new Error("JWT must have three parts");
  }

  const header = decodeBase64UrlJson(parts[0]);
  const payload = decodeBase64UrlJson(parts[1]);
  if (header.alg !== "RS256") {
    throw new Error("Only RS256 access tokens are accepted");
  }
  if (typeof payload.iss !== "string" || payload.iss !== config.issuer) {
    throw new Error("Invalid issuer");
  }
  if (!audienceMatches(payload.aud, config.audience)) {
    throw new Error("Invalid audience");
  }
  if (typeof payload.exp !== "number" || payload.exp <= Math.floor(Date.now() / 1000)) {
    throw new Error("Token is expired or missing exp");
  }

  const publicKey = await resolveJwk(header);
  const signingInput = new TextEncoder().encode(`${parts[0]}.${parts[1]}`);
  const signature = Buffer.from(parts[2], "base64url");
  const valid = await crypto.subtle.verify("RSASSA-PKCS1-v1_5", publicKey, signature, signingInput);
  if (!valid) {
    throw new Error("Invalid signature");
  }

  return {
    subject: typeof payload.sub === "string" ? payload.sub : "",
    email: typeof payload.email === "string" ? payload.email : undefined,
    preferredUsername: typeof payload.preferred_username === "string" ? payload.preferred_username : undefined,
    groups: groupsFromClaim(payload.groups),
    scopes: scopesFromClaim(payload.scope),
    claims: payload,
  };
}

export async function authenticateRequest(
  request: Request,
  config: AppConfig,
  resolveJwk: JwkResolver = createRemoteJwksResolver(config.jwksUri),
): Promise<AuthResult> {
  const token = bearerTokenFromHeader(request.headers.get("Authorization"));
  const challenge = wwwAuthenticateHeader(config);
  if (!token) {
    return { ok: false, status: 401, headers: { "WWW-Authenticate": challenge }, body: { error: "missing_bearer_token" } };
  }
  try {
    return { ok: true, auth: await verifyAccessToken(token, config, resolveJwk) };
  } catch {
    return { ok: false, status: 401, headers: { "WWW-Authenticate": challenge }, body: { error: "invalid_token" } };
  }
}
