import type { NextFunction, Request, Response } from "express";
import type { AppConfig } from "../config.js";
import { verifyAccessToken, UnauthorizedError } from "./jwks.js";
import type { AuthContext } from "./policy.js";

declare global {
  namespace Express {
    interface Request {
      auth?: AuthContext;
    }
  }
}

export type TokenVerifier = (token: string, config: AppConfig) => Promise<AuthContext>;

export function wwwAuthenticateHeader(config: AppConfig): string {
  return `Bearer realm="${config.realm}", resource_metadata="${config.baseUrl}/.well-known/oauth-protected-resource"`;
}

export function readBearerToken(req: Request): string | undefined {
  const authorization = req.header("authorization");
  if (!authorization) return undefined;
  const [scheme, token] = authorization.split(/\s+/, 2);
  if (scheme?.toLowerCase() !== "bearer" || !token) return undefined;
  return token;
}

export function requireBearerAuth(config: AppConfig, verifier: TokenVerifier = verifyAccessToken) {
  return async (req: Request, res: Response, next: NextFunction) => {
    const token = readBearerToken(req);
    if (!token) {
      res.setHeader("WWW-Authenticate", wwwAuthenticateHeader(config));
      res.status(401).json({ error: "unauthorized", error_description: "Missing Bearer access token" });
      return;
    }

    try {
      req.auth = await verifier(token, config);
      next();
    } catch (error) {
      const description = error instanceof Error ? error.message : "Invalid Bearer access token";
      res.setHeader("WWW-Authenticate", `${wwwAuthenticateHeader(config)}, error="invalid_token"`);
      res.status(error instanceof UnauthorizedError ? 401 : 401).json({
        error: "invalid_token",
        error_description: description,
      });
    }
  };
}
