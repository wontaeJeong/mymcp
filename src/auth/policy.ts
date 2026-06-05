export type AuthContext = {
  subject: string;
  issuer: string;
  audience: string[];
  scopes: string[];
  email?: string;
  preferredUsername?: string;
  groups: string[];
  expiresAt: number;
};

export class ForbiddenError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "ForbiddenError";
  }
}

export function hasScope(auth: AuthContext, scope: string): boolean {
  return auth.scopes.includes(scope);
}

export function hasGroup(auth: AuthContext, group: string): boolean {
  return auth.groups.includes(group) || auth.groups.includes(`/${group}`);
}

export function requireScope(auth: AuthContext, scope: string): void {
  if (!hasScope(auth, scope)) {
    throw new ForbiddenError(`Missing required scope: ${scope}`);
  }
}

export function requireAdmin(auth: AuthContext): void {
  if (!hasScope(auth, "mcp:admin") && !hasGroup(auth, "admin")) {
    throw new ForbiddenError("admin_status requires the admin group or mcp:admin scope");
  }
}
