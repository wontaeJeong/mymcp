import { createRemoteJWKSet, jwtVerify, type JWTPayload } from "jose";
import type { AppConfig } from "../config.js";
import type { AuthContext } from "./policy.js";

const jwksCache = new Map<string, ReturnType<typeof createRemoteJWKSet>>();

export class UnauthorizedError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "UnauthorizedError";
  }
}

function getRemoteJwks(jwksUri: string): ReturnType<typeof createRemoteJWKSet> {
  const cached = jwksCache.get(jwksUri);
  if (cached) return cached;
  const jwks = createRemoteJWKSet(new URL(jwksUri));
  jwksCache.set(jwksUri, jwks);
  return jwks;
}

function parseScopes(payload: JWTPayload): string[] {
  const rawScope = payload.scope;
  if (typeof rawScope !== "string") return [];
  return rawScope.split(" ").map((scope) => scope.trim()).filter(Boolean);
}

function parseGroups(payload: JWTPayload): string[] {
  const rawGroups = payload.groups;
  return Array.isArray(rawGroups) ? rawGroups.filter((group): group is string => typeof group === "string") : [];
}

function parseAudience(payload: JWTPayload): string[] {
  if (Array.isArray(payload.aud)) return payload.aud;
  return typeof payload.aud === "string" ? [payload.aud] : [];
}

export function assertAcceptedAudience(payload: JWTPayload, acceptedAudiences: string[]): void {
  const audiences = parseAudience(payload);
  if (!audiences.some((audience) => acceptedAudiences.includes(audience))) {
    throw new UnauthorizedError(`Invalid audience. Expected one of: ${acceptedAudiences.join(", ")}`);
  }
}

export async function verifyAccessToken(token: string, config: AppConfig): Promise<AuthContext> {
  try {
    const { payload } = await jwtVerify(token, getRemoteJwks(config.jwksUri), {
      issuer: config.issuer,
    });

    assertAcceptedAudience(payload, config.audiences);

    const subject = payload.sub;
    if (!subject) throw new UnauthorizedError("Token is missing sub claim");
    if (!payload.exp) throw new UnauthorizedError("Token is missing exp claim");

    return {
      subject,
      issuer: payload.iss ?? config.issuer,
      audience: parseAudience(payload),
      scopes: parseScopes(payload),
      email: typeof payload.email === "string" ? payload.email : undefined,
      preferredUsername: typeof payload.preferred_username === "string" ? payload.preferred_username : undefined,
      groups: parseGroups(payload),
      expiresAt: payload.exp,
    };
  } catch (error) {
    if (error instanceof UnauthorizedError) throw error;
    throw new UnauthorizedError(error instanceof Error ? error.message : "Invalid bearer token");
  }
}
