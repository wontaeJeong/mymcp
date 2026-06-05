package main

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, body any, headers map[string]string) {
	for key, value := range headers {
		w.Header().Set(key, value)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func addCORS(w http.ResponseWriter, r *http.Request, config AppConfig) {
	if len(config.CORSOrigins) == 0 {
		return
	}
	origin := r.Header.Get("Origin")
	for _, allowed := range config.CORSOrigins {
		if origin == allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			return
		}
	}
}

func CreateHandler(config AppConfig, resolveJWK JWKResolver) http.Handler {
	if config.MCPPath == "" {
		config = LoadConfig(nil)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addCORS(w, r, config)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method == http.MethodGet && (r.URL.Path == "/.well-known/oauth-protected-resource" || r.URL.Path == "/.well-known/oauth-protected-resource/mcp") {
			writeJSON(w, http.StatusOK, protectedResourceMetadata(config), nil)
			return
		}

		if r.URL.Path == config.MCPPath {
			authResult := authenticateRequest(r, config, resolveJWK)
			if !authResult.OK {
				writeJSON(w, authResult.Status, authResult.Body, authResult.Headers)
				return
			}
			if r.Method != http.MethodPost {
				writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed", "message": "Use POST for MCP JSON-RPC requests."}, nil)
				return
			}
			var rpc JSONRPCRequest
			if err := json.NewDecoder(r.Body).Decode(&rpc); err != nil {
				writeJSON(w, http.StatusBadRequest, rpcError(nil, -32700, "Invalid JSON body"), nil)
				return
			}
			mcp := handleMCPJSONRPC(rpc, authResult.Auth)
			if mcp.Body == nil {
				w.WriteHeader(mcp.Status)
				return
			}
			writeJSON(w, mcp.Status, mcp.Body, nil)
			return
		}

		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"}, nil)
	})
}
