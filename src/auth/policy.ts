import type { AuthContext } from "./bearer-auth.ts";

export function scopesFromClaim(scope: unknown): string[] {
  if (typeof scope === "string") {
    return scope.split(/\s+/).filter(Boolean);
  }
  if (Array.isArray(scope)) {
    return scope.filter((entry): entry is string => typeof entry === "string");
  }
  return [];
}

export function groupsFromClaim(groups: unknown): string[] {
  if (!Array.isArray(groups)) {
    return [];
  }
  return groups.filter((entry): entry is string => typeof entry === "string");
}

export function hasScope(auth: Pick<AuthContext, "scopes">, scope: string): boolean {
  return auth.scopes.includes(scope);
}

export function hasGroup(auth: Pick<AuthContext, "groups">, group: string): boolean {
  return auth.groups.includes(group) || auth.groups.includes(`/${group}`);
}

export function canListTools(auth: Pick<AuthContext, "scopes">): boolean {
  return hasScope(auth, "mcp:tools:read");
}

export function canCallTools(auth: Pick<AuthContext, "scopes">): boolean {
  return hasScope(auth, "mcp:tools:execute");
}

export function canCallAdminTool(auth: Pick<AuthContext, "scopes" | "groups">): boolean {
  return hasScope(auth, "mcp:admin") || hasGroup(auth, "admin");
}
