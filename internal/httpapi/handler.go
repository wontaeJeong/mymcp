package httpapi

import (
	"encoding/json"
	"net/http"

	"mymcp/internal/auth"
	"mymcp/internal/config"
	"mymcp/internal/mcp"
	"mymcp/internal/metadata"
)

func writeJSON(w http.ResponseWriter, status int, body any, headers map[string]string) {
	for key, value := range headers {
		w.Header().Set(key, value)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func addCORS(w http.ResponseWriter, r *http.Request, cfg config.Config) {
	if len(cfg.CORSOrigins) == 0 {
		return
	}
	origin := r.Header.Get("Origin")
	for _, allowed := range cfg.CORSOrigins {
		if origin == allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			return
		}
	}
}

func NewHandler(cfg config.Config, resolveJWK auth.JWKResolver) http.Handler {
	if cfg.MCPPath == "" {
		cfg = config.Load(nil)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addCORS(w, r, cfg)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method == http.MethodGet && (r.URL.Path == "/.well-known/oauth-protected-resource" || r.URL.Path == "/.well-known/oauth-protected-resource/mcp") {
			writeJSON(w, http.StatusOK, metadata.ProtectedResource(cfg), nil)
			return
		}

		if r.URL.Path == cfg.MCPPath {
			authResult := auth.AuthenticateRequest(r, cfg, resolveJWK)
			if !authResult.OK {
				writeJSON(w, authResult.Status, authResult.Body, authResult.Headers)
				return
			}
			if r.Method != http.MethodPost {
				writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed", "message": "Use POST for MCP JSON-RPC requests."}, nil)
				return
			}
			var rpc mcp.JSONRPCRequest
			if err := json.NewDecoder(r.Body).Decode(&rpc); err != nil {
				writeJSON(w, http.StatusBadRequest, mcp.RPCError(nil, -32700, "Invalid JSON body"), nil)
				return
			}
			mcpResponse := mcp.HandleJSONRPC(rpc, authResult.Auth)
			if mcpResponse.Body == nil {
				w.WriteHeader(mcpResponse.Status)
				return
			}
			writeJSON(w, mcpResponse.Status, mcpResponse.Body, nil)
			return
		}

		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"}, nil)
	})
}
