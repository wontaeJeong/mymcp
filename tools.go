package main

import "errors"

type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

var tools = []ToolDefinition{
	{
		Name:        "whoami",
		Description: "Return the authenticated Keycloak user attached to this MCP request.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
	},
	{
		Name:        "echo",
		Description: "Echo a message string back to the caller.",
		InputSchema: map[string]any{
			"type":                 "object",
			"properties":           map[string]any{"message": map[string]any{"type": "string"}},
			"required":             []string{"message"},
			"additionalProperties": false,
		},
	},
	{
		Name:        "admin_status",
		Description: "Return admin-only demo status. Requires the admin group or mcp:admin scope.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
	},
}

type forbiddenToolError struct{ message string }

func (e forbiddenToolError) Error() string { return e.message }

func callTool(name string, args any, auth AuthContext) (any, error) {
	switch name {
	case "whoami":
		return map[string]any{
			"subject":            auth.Subject,
			"email":              auth.Email,
			"preferred_username": auth.PreferredUsername,
			"groups":             auth.Groups,
			"scopes":             auth.Scopes,
		}, nil
	case "echo":
		arguments, _ := args.(map[string]any)
		message, ok := arguments["message"].(string)
		if !ok {
			return nil, errors.New("echo requires a string message argument")
		}
		return map[string]any{"message": message}, nil
	case "admin_status":
		if !canCallAdminTool(auth) {
			return nil, forbiddenToolError{"admin_status requires admin group or mcp:admin scope"}
		}
		return map[string]any{"status": "ok", "admin": true}, nil
	default:
		return nil, errors.New("Unknown tool: " + name)
	}
}
