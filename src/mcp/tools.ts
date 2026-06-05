import type { AuthContext } from "../auth/policy.js";
import { requireAdmin } from "../auth/policy.js";

export type ToolDefinition = {
  name: string;
  description: string;
  inputSchema: Record<string, unknown>;
};

export const tools: ToolDefinition[] = [
  {
    name: "whoami",
    description: "Return the authenticated Keycloak user context seen by the MCP resource server.",
    inputSchema: { type: "object", additionalProperties: false, properties: {} },
  },
  {
    name: "echo",
    description: "Return the provided message.",
    inputSchema: {
      type: "object",
      additionalProperties: false,
      properties: { message: { type: "string" } },
      required: ["message"],
    },
  },
  {
    name: "admin_status",
    description: "Return admin-only status. Requires the admin group or mcp:admin scope.",
    inputSchema: { type: "object", additionalProperties: false, properties: {} },
  },
];

export async function callTool(name: string, args: unknown, auth: AuthContext): Promise<unknown> {
  switch (name) {
    case "whoami":
      return {
        subject: auth.subject,
        email: auth.email,
        preferred_username: auth.preferredUsername,
        groups: auth.groups,
        scopes: auth.scopes,
        issuer: auth.issuer,
        audience: auth.audience,
      };
    case "echo": {
      const message = typeof args === "object" && args !== null && "message" in args ? (args as { message?: unknown }).message : undefined;
      if (typeof message !== "string") {
        throw new Error("echo requires a string input property named message");
      }
      return { message };
    }
    case "admin_status":
      requireAdmin(auth);
      return { ok: true, message: "You have admin access to this MCP resource." };
    default:
      throw new Error(`Unknown tool: ${name}`);
  }
}

export function asMcpContent(result: unknown) {
  return [{ type: "text", text: typeof result === "string" ? result : JSON.stringify(result, null, 2) }];
}
