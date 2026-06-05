import cors from "cors";
import express from "express";
import { loadConfig, type AppConfig } from "./config.js";
import { requireBearerAuth, type TokenVerifier } from "./auth/bearer-auth.js";
import { protectedResourceMetadata } from "./auth/protected-resource-metadata.js";
import { handleMcpRequest } from "./mcp/mcp-server.js";

export function createApp(config: AppConfig = loadConfig(), verifier?: TokenVerifier) {
  const app = express();

  app.use(express.json({ limit: "1mb" }));
  if (config.corsOrigins.length > 0) {
    app.use(cors({ origin: config.corsOrigins, credentials: false }));
  }

  app.get("/.well-known/oauth-protected-resource", (_req, res) => {
    res.json(protectedResourceMetadata(config));
  });

  app.get("/.well-known/oauth-protected-resource/mcp", (_req, res) => {
    res.json(protectedResourceMetadata(config));
  });

  app.all(config.resourcePath, requireBearerAuth(config, verifier));
  app.post(config.resourcePath, handleMcpRequest);

  return app;
}

if (process.env.NODE_ENV !== "test") {
  const config = loadConfig();
  createApp(config).listen(config.port, () => {
    console.log(`MCP OAuth protected resource listening on ${config.baseUrl}${config.resourcePath}`);
    console.log(`OAuth protected resource metadata: ${config.baseUrl}/.well-known/oauth-protected-resource`);
  });
}
