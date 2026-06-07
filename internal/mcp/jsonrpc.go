package mcp

type JSONRPCRequest struct {
	JSONRPC string         `json:"jsonrpc,omitempty"`
	ID      any            `json:"id,omitempty"`
	Method  string         `json:"method,omitempty"`
	Params  map[string]any `json:"params,omitempty"`
}

type Response struct {
	Status int
	Body   any
}

func rpcResult(id any, value any) map[string]any {
	if id == nil {
		id = nil
	}
	return map[string]any{"jsonrpc": "2.0", "id": id, "result": value}
}
