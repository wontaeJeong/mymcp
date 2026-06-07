# Production Notes

This demo keeps the application small while separating internal packages for production growth.

Before production use, replace local wildcard redirect URIs, constrain scope issuance, configure explicit CORS origins, and use Authorization Code + PKCE rather than password grants.
