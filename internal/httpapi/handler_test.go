package httpapi_test

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mymcp/internal/auth"
	"mymcp/internal/config"
	"mymcp/internal/httpapi"
)

var testConfig = config.Config{
	Port:        3000,
	BaseURL:     "http://localhost:3000",
	MCPPath:     "/mcp",
	Realm:       "mcp-demo",
	RealmName:   "mcp-demo",
	Issuer:      "http://localhost:8080/realms/mcp-demo",
	JWKSURI:     "http://localhost:8080/realms/mcp-demo/protocol/openid-connect/certs",
	Audience:    "mcp-demo-resource",
	CORSOrigins: nil,
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
	jwk := auth.JWK{
		Kid: "test-key",
		Alg: "RS256",
		Kty: "RSA",
		N:   b64url(privateKey.PublicKey.N.Bytes()),
		E:   b64url(exponent),
	}
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
		"groups":             []string{},
	}
	for key, value := range overrides {
		payload[key] = value
	}
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

func request(handler http.Handler, method, path string, body any, accessToken string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		encoded, _ := json.Marshal(body)
		reader = bytes.NewReader(encoded)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func decodeJSON(t *testing.T, res *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response JSON: %v\nbody: %s", err, res.Body.String())
	}
	return body
}

func TestProtectedResourceMetadata(t *testing.T) {
	res := request(httpapi.NewHandler(testConfig, nil), http.MethodGet, "/.well-known/oauth-protected-resource", nil, "")
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := decodeJSON(t, res)
	if body["resource"] != "http://localhost:3000/mcp" {
		t.Fatalf("resource = %v", body["resource"])
	}
	servers := body["authorization_servers"].([]any)
	if servers[0] != "http://localhost:8080/realms/mcp-demo" {
		t.Fatalf("authorization_servers = %v", servers)
	}
	scopes := body["scopes_supported"].([]any)
	if len(scopes) != 3 || scopes[0] != "mcp:tools:read" || scopes[1] != "mcp:tools:execute" || scopes[2] != "mcp:admin" {
		t.Fatalf("scopes_supported = %v", scopes)
	}
}

func TestBearerAuthMissingAuthorization(t *testing.T) {
	keys := makeKeys(t)
	res := request(httpapi.NewHandler(testConfig, keys.resolver), http.MethodGet, "/mcp", nil, "")
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
	challenge := res.Header().Get("WWW-Authenticate")
	if !bytes.Contains([]byte(challenge), []byte(`Bearer realm="mcp-demo"`)) || !bytes.Contains([]byte(challenge), []byte(`resource_metadata="http://localhost:3000/.well-known/oauth-protected-resource"`)) {
		t.Fatalf("unexpected challenge: %s", challenge)
	}
}

func TestBearerAuthRejectsWrongIssuer(t *testing.T) {
	keys := makeKeys(t)
	bad := token(t, keys.privateKey, map[string]any{"iss": "http://localhost:8080/realms/other"})
	res := request(httpapi.NewHandler(testConfig, keys.resolver), http.MethodPost, "/mcp", map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}, bad)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
	if body := decodeJSON(t, res); body["error"] != "invalid_token" {
		t.Fatalf("body = %v", body)
	}
}

func TestBearerAuthRejectsWrongAudience(t *testing.T) {
	keys := makeKeys(t)
	bad := token(t, keys.privateKey, map[string]any{"aud": "wrong-audience"})
	res := request(httpapi.NewHandler(testConfig, keys.resolver), http.MethodPost, "/mcp", map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}, bad)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
	if body := decodeJSON(t, res); body["error"] != "invalid_token" {
		t.Fatalf("body = %v", body)
	}
}

func rpc(method string, params map[string]any) map[string]any {
	body := map[string]any{"jsonrpc": "2.0", "id": 1, "method": method}
	if params != nil {
		body["params"] = params
	}
	return body
}

func postMCP(t *testing.T, scope string, body any, groups []string) *httptest.ResponseRecorder {
	t.Helper()
	keys := makeKeys(t)
	accessToken := token(t, keys.privateKey, map[string]any{"scope": scope, "groups": groups})
	return request(httpapi.NewHandler(testConfig, keys.resolver), http.MethodPost, "/mcp", body, accessToken)
}

func TestPolicyDeniesToolCallsWithoutExecuteScope(t *testing.T) {
	res := postMCP(t, "mcp:tools:read", rpc("tools/call", map[string]any{"name": "echo", "arguments": map[string]any{"message": "hello"}}), nil)
	if res.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusForbidden)
	}
	body := decodeJSON(t, res)
	errorBody := body["error"].(map[string]any)
	if !bytes.Contains([]byte(errorBody["message"].(string)), []byte("mcp:tools:execute")) {
		t.Fatalf("body = %v", body)
	}
}

func TestPolicyAllowsToolsListWithReadScope(t *testing.T) {
	res := postMCP(t, "mcp:tools:read", rpc("tools/list", nil), nil)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := decodeJSON(t, res)
	result := body["result"].(map[string]any)
	listedTools := result["tools"].([]any)
	names := []string{listedTools[0].(map[string]any)["name"].(string), listedTools[1].(map[string]any)["name"].(string), listedTools[2].(map[string]any)["name"].(string)}
	want := []string{"whoami", "echo", "admin_status"}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("tools = %v, want %v", names, want)
		}
	}
}

func TestPolicyAllowsEchoWithExecuteScope(t *testing.T) {
	res := postMCP(t, "mcp:tools:execute", rpc("tools/call", map[string]any{"name": "echo", "arguments": map[string]any{"message": "hello"}}), nil)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := decodeJSON(t, res)
	content := body["result"].(map[string]any)["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if !bytes.Contains([]byte(text), []byte("hello")) {
		t.Fatalf("text = %s", text)
	}
}

func TestPolicyDeniesAdminStatusWithoutAdmin(t *testing.T) {
	res := postMCP(t, "mcp:tools:execute", rpc("tools/call", map[string]any{"name": "admin_status", "arguments": map[string]any{}}), nil)
	if res.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusForbidden)
	}
	body := decodeJSON(t, res)
	errorBody := body["error"].(map[string]any)
	if !bytes.Contains([]byte(errorBody["message"].(string)), []byte("admin")) {
		t.Fatalf("body = %v", body)
	}
}

func TestPolicyAllowsAdminStatusWithAdminGroup(t *testing.T) {
	res := postMCP(t, "mcp:tools:execute", rpc("tools/call", map[string]any{"name": "admin_status", "arguments": map[string]any{}}), []string{"admin"})
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := decodeJSON(t, res)
	content := body["result"].(map[string]any)["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if !bytes.Contains([]byte(text), []byte("admin")) {
		t.Fatalf("text = %s", text)
	}
}
