import type { AuthContext } from "../auth/bearer-auth.ts";
import { canCallAdminTool } from "../auth/policy.ts";

export type ToolDefinition = {
  name: string;
  description: string;
  inputSchema: Record<string, unknown>;
};

export const tools: ToolDefinition[] = [
  {
    name: "whoami",
    description: "Return the authenticated Keycloak user attached to this MCP request.",
    inputSchema: { type: "object", properties: {}, additionalProperties: false },
  },
  {
    name: "echo",
    description: "Echo a message string back to the caller.",
    inputSchema: {
      type: "object",
      properties: { message: { type: "string" } },
      required: ["message"],
      additionalProperties: false,
    },
  },
  {
    name: "admin_status",
    description: "Return admin-only demo status. Requires the admin group or mcp:admin scope.",
    inputSchema: { type: "object", properties: {}, additionalProperties: false },
  },
];

export function callTool(name: string, args: unknown, auth: AuthContext): unknown {
  if (name === "whoami") {
    return {
      subject: auth.subject,
      email: auth.email,
      preferred_username: auth.preferredUsername,
      groups: auth.groups,
      scopes: auth.scopes,
    };
  }

  if (name === "echo") {
    const message = typeof args === "object" && args !== null && "message" in args ? (args as { message?: unknown }).message : undefined;
    if (typeof message !== "string") {
      throw new Error("echo requires a string message argument");
    }
    return { message };
  }

  if (name === "admin_status") {
    if (!canCallAdminTool(auth)) {
      const error = new Error("admin_status requires admin group or mcp:admin scope");
      error.name = "ForbiddenToolError";
      throw error;
    }
    return { status: "ok", admin: true };
  }

  throw new Error(`Unknown tool: ${name}`);
}
