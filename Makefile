.PHONY: up down dev build test install docker-build docker-up docker-down token-alice token-admin

up:
	docker compose up -d

down:
	docker compose down

install:
	go mod download

dev:
	go run .

build:
	go build ./...

test:
	go test ./...

docker-build:
	docker build -t mymcp .

docker-up:
	docker compose --profile app up -d --build

docker-down:
	docker compose --profile app down

token-alice:
	curl -s -X POST http://localhost:8080/realms/mcp-demo/protocol/openid-connect/token \
	  -H 'Content-Type: application/x-www-form-urlencoded' \
	  -d 'client_id=mcp-demo-cli' \
	  -d 'grant_type=password' \
	  -d 'username=alice' \
	  -d 'password=alice' \
	  -d 'scope=openid profile email mcp:tools:read mcp:tools:execute' | jq -r .access_token

token-admin:
	curl -s -X POST http://localhost:8080/realms/mcp-demo/protocol/openid-connect/token \
	  -H 'Content-Type: application/x-www-form-urlencoded' \
	  -d 'client_id=mcp-demo-cli' \
	  -d 'grant_type=password' \
	  -d 'username=admin-user' \
	  -d 'password=admin-user' \
	  -d 'scope=openid profile email mcp:tools:read mcp:tools:execute mcp:admin' | jq -r .access_token
