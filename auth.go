package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type AuthContext struct {
	Subject           string
	Email             string
	PreferredUsername string
	Groups            []string
	Scopes            []string
	Claims            map[string]any
}

type AuthResult struct {
	OK      bool
	Auth    AuthContext
	Status  int
	Headers map[string]string
	Body    map[string]string
}

func bearerTokenFromHeader(header string) string {
	parts := strings.Fields(strings.TrimSpace(header))
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

func decodeBase64URLJSON(segment string) (map[string]any, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return nil, err
	}
	var value map[string]any
	if err := json.Unmarshal(decoded, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func audienceMatches(aud any, expected string) bool {
	switch typed := aud.(type) {
	case string:
		return typed == expected
	case []any:
		for _, entry := range typed {
			if value, ok := entry.(string); ok && value == expected {
				return true
			}
		}
	}
	return false
}

func verifyAccessToken(token string, config AppConfig, resolveJWK JWKResolver) (AuthContext, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return AuthContext{}, errors.New("JWT must have three parts")
	}
	header, err := decodeBase64URLJSON(parts[0])
	if err != nil {
		return AuthContext{}, err
	}
	payload, err := decodeBase64URLJSON(parts[1])
	if err != nil {
		return AuthContext{}, err
	}
	if alg, _ := header["alg"].(string); alg != "RS256" {
		return AuthContext{}, errors.New("only RS256 access tokens are accepted")
	}
	if iss, _ := payload["iss"].(string); iss != config.Issuer {
		return AuthContext{}, errors.New("invalid issuer")
	}
	if !audienceMatches(payload["aud"], config.Audience) {
		return AuthContext{}, errors.New("invalid audience")
	}
	exp, ok := payload["exp"].(float64)
	if !ok || int64(exp) <= time.Now().Unix() {
		return AuthContext{}, errors.New("token is expired or missing exp")
	}

	publicKey, err := resolveJWK(header)
	if err != nil {
		return AuthContext{}, err
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return AuthContext{}, err
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature); err != nil {
		return AuthContext{}, err
	}

	subject, _ := payload["sub"].(string)
	email, _ := payload["email"].(string)
	preferredUsername, _ := payload["preferred_username"].(string)
	return AuthContext{
		Subject:           subject,
		Email:             email,
		PreferredUsername: preferredUsername,
		Groups:            groupsFromClaim(payload["groups"]),
		Scopes:            scopesFromClaim(payload["scope"]),
		Claims:            payload,
	}, nil
}

func authenticateRequest(r *http.Request, config AppConfig, resolveJWK JWKResolver) AuthResult {
	if resolveJWK == nil {
		resolveJWK = createRemoteJwksResolver(config.JWKSURI)
	}
	challenge := wwwAuthenticateHeader(config)
	token := bearerTokenFromHeader(r.Header.Get("Authorization"))
	if token == "" {
		return AuthResult{OK: false, Status: http.StatusUnauthorized, Headers: map[string]string{"WWW-Authenticate": challenge}, Body: map[string]string{"error": "missing_bearer_token"}}
	}
	auth, err := verifyAccessToken(token, config, resolveJWK)
	if err != nil {
		return AuthResult{OK: false, Status: http.StatusUnauthorized, Headers: map[string]string{"WWW-Authenticate": challenge}, Body: map[string]string{"error": "invalid_token"}}
	}
	return AuthResult{OK: true, Auth: auth}
}
