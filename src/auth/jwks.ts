export type JwtHeader = Record<string, unknown>;
export type Jwk = JsonWebKey & { kid?: string; alg?: string; use?: string };
export type JwkResolver = (header: JwtHeader) => Promise<CryptoKey>;

async function importRsaPublicKey(jwk: Jwk): Promise<CryptoKey> {
  return crypto.subtle.importKey(
    "jwk",
    jwk,
    { name: "RSASSA-PKCS1-v1_5", hash: "SHA-256" },
    false,
    ["verify"],
  );
}

export function createStaticJwksResolver(jwks: { keys: Jwk[] }): JwkResolver {
  return async (header) => {
    const kid = typeof header.kid === "string" ? header.kid : undefined;
    const jwk = jwks.keys.find((candidate) => !kid || candidate.kid === kid);
    if (!jwk) {
      throw new Error("No matching JWK");
    }
    return importRsaPublicKey(jwk);
  };
}

export function createRemoteJwksResolver(jwksUri: string): JwkResolver {
  let cache: { keys: Jwk[] } | undefined;
  return async (header) => {
    if (!cache) {
      const response = await fetch(jwksUri);
      if (!response.ok) {
        throw new Error(`JWKS fetch failed: ${response.status}`);
      }
      cache = await response.json() as { keys: Jwk[] };
    }
    return createStaticJwksResolver(cache)(header);
  };
}
