UNIT_TAG=unit
INTEGRATION_TAG=integration
E2E_TAG=e2e

DOCKER_COMPOSE_SCRIPTS = docker compose -f scripts/docker-compose.yml
DOCKER_COMPOSE_TEST_APP = docker compose -f docker-compose.base.yml -f docker-compose.test.yml
NETWORK_NAME=app-network

GO_TEST = go test -v

PLAYWRIGHT_VERSION=v0.5200.0

.PHONY: unit
unit:
	@echo "Running Unit Tests"
	$(GO_TEST) -tags=$(UNIT_TAG) ./...

.PHONY: integration
integration:
	@echo "Running Integration Tests"
	$(GO_TEST) -tags=$(INTEGRATION_TAG) ./...

.PHONY: e2e
e2e: create-network start-scripts start-app health-check
	@echo "Running E2E Tests"
	$(GO_TEST) -tags=$(E2E_TAG) ./internal/...

.PHONY: start-scripts
start-scripts:
	@echo "Starting Docker Compose from 'scripts' directory..."
	$(DOCKER_COMPOSE_SCRIPTS) up -d

.PHONY: start-app
start-app:
	@echo "Starting Docker Compose for the application..."
	$(DOCKER_COMPOSE_TEST_APP) up -d --build

.PHONY: stop
stop:
	@echo "Stopping and cleaning up Docker containers..."
	$(DOCKER_COMPOSE_SCRIPTS) down --volumes --remove-orphans
	$(DOCKER_COMPOSE_TEST_APP) down --volumes --remove-orphans

.PHONY: install-playwright
install-playwright:
	@echo "Installing Playwright version $(PLAYWRIGHT_VERSION)..."
	go run github.com/playwright-community/playwright-go/cmd/playwright@$(PLAYWRIGHT_VERSION) install --with-deps

.PHONY: create-network
create-network:
	@echo "Creating network $(NETWORK_NAME) if it does not exist..."
	docker network ls | grep -q $(NETWORK_NAME) || docker network create $(NETWORK_NAME)

.PHONY: health-check
health-check:
	@echo "Waiting for server to be ready on route /healthz..."
	@timeout=60; \
	while [ $$timeout -gt 0 ] && ! curl -sSf http://localhost:8080/healthz > /dev/null; do \
		echo "Still waiting for server to be ready..."; \
		sleep 2; \
		timeout=$$((timeout - 2)); \
	done; \
	if [ $$timeout -le 0 ]; then \
		echo "Server did not become ready in time"; \
		exit 1; \
	fi
