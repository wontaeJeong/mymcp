package auth

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"mymcp/internal/config"
	"mymcp/internal/policy"
)

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

func VerifyAccessToken(token string, cfg config.Config, resolveJWK JWKResolver) (AuthContext, error) {
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
	if iss, _ := payload["iss"].(string); iss != cfg.Issuer {
		return AuthContext{}, errors.New("invalid issuer")
	}
	if !audienceMatches(payload["aud"], cfg.Audience) {
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
		Groups:            policy.GroupsFromClaim(payload["groups"]),
		Scopes:            policy.ScopesFromClaim(payload["scope"]),
		Claims:            payload,
	}, nil
}
