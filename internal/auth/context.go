package auth

type AuthContext struct {
	Subject           string
	Email             string
	PreferredUsername string
	Groups            []string
	Scopes            []string
	Claims            map[string]any
}

type AuthResult struct {
	OK      bool
	Auth    AuthContext
	Status  int
	Headers map[string]string
	Body    map[string]string
}
