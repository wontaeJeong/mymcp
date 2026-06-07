package auth_test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"maps"
	"math/big"
	"testing"
	"time"

	"mymcp/internal/auth"
	"mymcp/internal/config"
)

var testConfig = config.Config{
	Issuer:   "http://localhost:8080/realms/mcp-demo",
	Audience: "mcp-demo-resource",
}

type testKeys struct {
	privateKey *rsa.PrivateKey
	resolver   auth.JWKResolver
}

func b64url(input []byte) string {
	return base64.RawURLEncoding.EncodeToString(input)
}

func makeKeys(t *testing.T) testKeys {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	exponent := big.NewInt(int64(privateKey.PublicKey.E)).Bytes()
	jwk := auth.JWK{Kid: "test-key", Alg: "RS256", Kty: "RSA", N: b64url(privateKey.PublicKey.N.Bytes()), E: b64url(exponent)}
	return testKeys{privateKey: privateKey, resolver: auth.CreateStaticJwksResolver(auth.JWKS{Keys: []auth.JWK{jwk}})}
}

func token(t *testing.T, privateKey *rsa.PrivateKey, overrides map[string]any) string {
	t.Helper()
	now := time.Now().Unix()
	header := map[string]any{"alg": "RS256", "typ": "JWT", "kid": "test-key"}
	payload := map[string]any{
		"sub":                "alice-subject",
		"iss":                testConfig.Issuer,
		"aud":                testConfig.Audience,
		"iat":                now,
		"exp":                now + 3600,
		"scope":              "mcp:tools:read mcp:tools:execute",
		"email":              "alice@example.com",
		"preferred_username": "alice",
		"groups":             []string{"admin"},
	}
	maps.Copy(payload, overrides)
	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)
	signingInput := b64url(headerJSON) + "." + b64url(payloadJSON)
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return signingInput + "." + b64url(signature)
}

func TestVerifyAccessTokenAcceptsValidToken(t *testing.T) {
	keys := makeKeys(t)
	ctx, err := auth.VerifyAccessToken(token(t, keys.privateKey, nil), testConfig, keys.resolver)
	if err != nil {
		t.Fatalf("VerifyAccessToken error = %v", err)
	}
	if ctx.Subject != "alice-subject" || ctx.Email != "alice@example.com" || ctx.PreferredUsername != "alice" {
		t.Fatalf("auth context = %#v", ctx)
	}
	if len(ctx.Scopes) != 2 || ctx.Scopes[0] != "mcp:tools:read" || ctx.Scopes[1] != "mcp:tools:execute" {
		t.Fatalf("scopes = %v", ctx.Scopes)
	}
	if len(ctx.Groups) != 1 || ctx.Groups[0] != "admin" {
		t.Fatalf("groups = %v", ctx.Groups)
	}
}

func TestVerifyAccessTokenRejectsWrongIssuer(t *testing.T) {
	keys := makeKeys(t)
	bad := token(t, keys.privateKey, map[string]any{"iss": "http://localhost:8080/realms/other"})
	if _, err := auth.VerifyAccessToken(bad, testConfig, keys.resolver); err == nil {
		t.Fatal("VerifyAccessToken accepted token with wrong issuer")
	}
}

func TestVerifyAccessTokenRejectsWrongAudience(t *testing.T) {
	keys := makeKeys(t)
	bad := token(t, keys.privateKey, map[string]any{"aud": "wrong-audience"})
	if _, err := auth.VerifyAccessToken(bad, testConfig, keys.resolver); err == nil {
		t.Fatal("VerifyAccessToken accepted token with wrong audience")
	}
}
