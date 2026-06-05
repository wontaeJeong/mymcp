.PHONY: install up down dev test build typecheck token-alice token-admin

install:
	npm install

up:
	docker compose up -d
	npm run dev

down:
	docker compose down

dev:
	npm run dev

build:
	npm run build

typecheck:
	npm run typecheck

test:
	npm test

token-alice:
	curl -s -X POST http://localhost:8080/realms/mcp-demo/protocol/openid-connect/token \
	  -H 'content-type: application/x-www-form-urlencoded' \
	  --data-urlencode 'grant_type=password' \
	  --data-urlencode 'client_id=mcp-demo-test' \
	  --data-urlencode 'username=alice' \
	  --data-urlencode 'password=alice' \
	  --data-urlencode 'scope=openid profile email mcp:tools:read mcp:tools:execute' | jq -r .access_token

token-admin:
	curl -s -X POST http://localhost:8080/realms/mcp-demo/protocol/openid-connect/token \
	  -H 'content-type: application/x-www-form-urlencoded' \
	  --data-urlencode 'grant_type=password' \
	  --data-urlencode 'client_id=mcp-demo-test' \
	  --data-urlencode 'username=admin-user' \
	  --data-urlencode 'password=admin-user' \
	  --data-urlencode 'scope=openid profile email mcp:tools:read mcp:tools:execute mcp:admin' | jq -r .access_token
