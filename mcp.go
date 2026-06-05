package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

type JSONRPCRequest struct {
	JSONRPC string         `json:"jsonrpc,omitempty"`
	ID      any            `json:"id,omitempty"`
	Method  string         `json:"method,omitempty"`
	Params  map[string]any `json:"params,omitempty"`
}

type MCPResponse struct {
	Status int
	Body   any
}

func rpcResult(id any, value any) map[string]any {
	if id == nil {
		id = nil
	}
	return map[string]any{"jsonrpc": "2.0", "id": id, "result": value}
}

func rpcError(id any, code int, message string) map[string]any {
	return map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": code, "message": message}}
}

func toolContent(value any) map[string]any {
	encoded, _ := json.MarshalIndent(value, "", "  ")
	return map[string]any{"content": []map[string]string{{"type": "text", "text": string(encoded)}}}
}

func handleMCPJSONRPC(rpc JSONRPCRequest, auth AuthContext) MCPResponse {
	switch rpc.Method {
	case "initialize":
		return MCPResponse{Status: http.StatusOK, Body: rpcResult(rpc.ID, map[string]any{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "keycloak-internal-oidc-mcp-demo", "version": "0.2.0"},
		})}
	case "notifications/initialized":
		return MCPResponse{Status: http.StatusAccepted}
	case "tools/list":
		if !canListTools(auth) {
			return MCPResponse{Status: http.StatusForbidden, Body: rpcError(rpc.ID, -32001, "tools/list requires mcp:tools:read scope")}
		}
		return MCPResponse{Status: http.StatusOK, Body: rpcResult(rpc.ID, map[string]any{"tools": tools})}
	case "tools/call":
		if !canCallTools(auth) {
			return MCPResponse{Status: http.StatusForbidden, Body: rpcError(rpc.ID, -32001, "tools/call requires mcp:tools:execute scope")}
		}
		name, _ := rpc.Params["name"].(string)
		args := rpc.Params["arguments"]
		if args == nil {
			args = map[string]any{}
		}
		if name == "" {
			return MCPResponse{Status: http.StatusBadRequest, Body: rpcError(rpc.ID, -32602, "tools/call requires params.name")}
		}
		value, err := callTool(name, args, auth)
		if err != nil {
			status := http.StatusBadRequest
			code := -32602
			var forbidden forbiddenToolError
			if errors.As(err, &forbidden) {
				status = http.StatusForbidden
				code = -32001
			}
			return MCPResponse{Status: status, Body: rpcError(rpc.ID, code, err.Error())}
		}
		return MCPResponse{Status: http.StatusOK, Body: rpcResult(rpc.ID, toolContent(value))}
	default:
		method := rpc.Method
		if method == "" {
			method = "<missing>"
		}
		return MCPResponse{Status: http.StatusNotFound, Body: rpcError(rpc.ID, -32601, "Unsupported MCP method: "+method)}
	}
}
