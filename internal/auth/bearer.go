package auth

import (
	"net/http"
	"strings"

	"mymcp/internal/config"
	"mymcp/internal/metadata"
)

func BearerTokenFromHeader(header string) string {
	parts := strings.Fields(strings.TrimSpace(header))
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

func AuthenticateRequest(r *http.Request, cfg config.Config, resolveJWK JWKResolver) AuthResult {
	if resolveJWK == nil {
		resolveJWK = CreateRemoteJwksResolver(cfg.JWKSURI)
	}
	challenge := metadata.WWWAuthenticateHeader(cfg)
	token := BearerTokenFromHeader(r.Header.Get("Authorization"))
	if token == "" {
		return AuthResult{OK: false, Status: http.StatusUnauthorized, Headers: map[string]string{"WWW-Authenticate": challenge}, Body: map[string]string{"error": "missing_bearer_token"}}
	}
	auth, err := VerifyAccessToken(token, cfg, resolveJWK)
	if err != nil {
		return AuthResult{OK: false, Status: http.StatusUnauthorized, Headers: map[string]string{"WWW-Authenticate": challenge}, Body: map[string]string{"error": "invalid_token"}}
	}
	return AuthResult{OK: true, Auth: auth}
}
