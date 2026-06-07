package mcp

import "encoding/json"

func RPCError(id any, code int, message string) map[string]any {
	return map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": code, "message": message}}
}

func toolContent(value any) map[string]any {
	encoded, _ := json.MarshalIndent(value, "", "  ")
	return map[string]any{"content": []map[string]string{{"type": "text", "text": string(encoded)}}}
}
