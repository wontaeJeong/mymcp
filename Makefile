COMPOSE_FILE := deployments/docker-compose.yml

INTERNAL_CA_CERT ?= certs/internal-ca.crt
CA_CERT_ENV = $(if $(wildcard $(INTERNAL_CA_CERT)),SSL_CERT_FILE=$(INTERNAL_CA_CERT),)
PROXY_ENV = HTTP_PROXY="$(HTTP_PROXY)" HTTPS_PROXY="$(HTTPS_PROXY)" NO_PROXY="$(NO_PROXY)" http_proxy="$(http_proxy)" https_proxy="$(https_proxy)" no_proxy="$(no_proxy)"
DOCKER_PROXY_BUILD_ARGS = --build-arg HTTP_PROXY="$(HTTP_PROXY)" --build-arg HTTPS_PROXY="$(HTTPS_PROXY)" --build-arg NO_PROXY="$(NO_PROXY)" --build-arg http_proxy="$(http_proxy)" --build-arg https_proxy="$(https_proxy)" --build-arg no_proxy="$(no_proxy)"

.PHONY: up down dev build test install docker-build docker-up docker-down token-alice token-admin

up:
	docker compose -f $(COMPOSE_FILE) up -d

down:
	docker compose -f $(COMPOSE_FILE) down

install:
	$(PROXY_ENV) $(CA_CERT_ENV) go mod download

dev:
	$(PROXY_ENV) $(CA_CERT_ENV) go run ./cmd/mymcp

build:
	$(PROXY_ENV) $(CA_CERT_ENV) go build ./...

test:
	$(PROXY_ENV) $(CA_CERT_ENV) go test ./...

docker-build:
	docker build $(DOCKER_PROXY_BUILD_ARGS) -t mymcp .

docker-up:
	docker compose -f $(COMPOSE_FILE) --profile app up -d --build

docker-down:
	docker compose -f $(COMPOSE_FILE) --profile app down

token-alice:
	curl -s -X POST http://localhost:8080/realms/mcp-demo/protocol/openid-connect/token \
	  -H 'Content-Type: application/x-www-form-urlencoded' \
	  -d 'client_id=mcp-demo-cli' \
	  -d 'grant_type=password' \
	  -d 'username=alice' \
	  -d 'password=alice' \
	  -d 'scope=mcp:tools:read mcp:tools:execute' | jq -r .access_token

token-admin:
	curl -s -X POST http://localhost:8080/realms/mcp-demo/protocol/openid-connect/token \
	  -H 'Content-Type: application/x-www-form-urlencoded' \
	  -d 'client_id=mcp-demo-cli' \
	  -d 'grant_type=password' \
	  -d 'username=admin-user' \
	  -d 'password=admin-user' \
	  -d 'scope=mcp:tools:read mcp:tools:execute mcp:admin' | jq -r .access_token
