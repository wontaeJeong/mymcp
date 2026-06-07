package mcp

import (
	"errors"
	"net/http"

	"mymcp/internal/auth"
	"mymcp/internal/policy"
	"mymcp/internal/tools"
)

func HandleJSONRPC(rpc JSONRPCRequest, authContext auth.AuthContext) Response {
	switch rpc.Method {
	case "initialize":
		return Response{Status: http.StatusOK, Body: rpcResult(rpc.ID, map[string]any{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "keycloak-internal-oidc-mcp-demo", "version": "0.2.0"},
		})}
	case "notifications/initialized":
		return Response{Status: http.StatusAccepted}
	case "tools/list":
		if !policy.CanListTools(authContext.Scopes) {
			return Response{Status: http.StatusForbidden, Body: RPCError(rpc.ID, -32001, "tools/list requires mcp:tools:read scope")}
		}
		return Response{Status: http.StatusOK, Body: rpcResult(rpc.ID, map[string]any{"tools": tools.List()})}
	case "tools/call":
		if !policy.CanCallTools(authContext.Scopes) {
			return Response{Status: http.StatusForbidden, Body: RPCError(rpc.ID, -32001, "tools/call requires mcp:tools:execute scope")}
		}
		name, _ := rpc.Params["name"].(string)
		args := rpc.Params["arguments"]
		if args == nil {
			args = map[string]any{}
		}
		if name == "" {
			return Response{Status: http.StatusBadRequest, Body: RPCError(rpc.ID, -32602, "tools/call requires params.name")}
		}
		value, err := tools.Call(name, args, authContext)
		if err != nil {
			status := http.StatusBadRequest
			code := -32602
			var forbidden tools.ForbiddenError
			if errors.As(err, &forbidden) {
				status = http.StatusForbidden
				code = -32001
			}
			return Response{Status: status, Body: RPCError(rpc.ID, code, err.Error())}
		}
		return Response{Status: http.StatusOK, Body: rpcResult(rpc.ID, toolContent(value))}
	default:
		method := rpc.Method
		if method == "" {
			method = "<missing>"
		}
		return Response{Status: http.StatusNotFound, Body: RPCError(rpc.ID, -32601, "Unsupported MCP method: "+method)}
	}
}
