UNIT_TAG=unit
INTEGRATION_TAG=integration
E2E_TAG=e2e
ARCHITECTURE_TAG=arch

DOCKER_COMPOSE_SCRIPTS = docker compose -f scripts/docker-compose.yml
DOCKER_COMPOSE_TEST_APP = docker compose -f docker-compose.base.yml -f docker-compose.test.yml
NETWORK_NAME=app-network

GO_TEST = go test -v

PLAYWRIGHT_VERSION=v0.5200.0

MODULES = apps/subscription apps/weather apps/notification apps/gateway apps/email

.PHONY: unit
unit:
	@echo "Running Unit Tests"
	@exit_code=0; \
    for dir in $(MODULES); do \
    	echo "Running unit tests in $$dir..."; \
    	cd $$dir && $(GO_TEST) -tags=$(UNIT_TAG) ./... || exit_code=$$?; \
    	cd - > /dev/null; \
    done; \
    exit $$exit_code

.PHONY: integration
integration:
	@echo "Running Integration Tests"
	@exit_code=0; \
    for dir in $(MODULES); do \
    	echo "Running integration tests in $$dir..."; \
    	cd $$dir && $(GO_TEST) -tags=$(INTEGRATION_TAG) ./... || exit_code=$$?; \
    	cd - > /dev/null; \
    done; \
    exit $$exit_code

.PHONY: e2e
e2e: create-network start-scripts start-app
	@echo "Running E2E Tests"
	$(GO_TEST) -tags=$(E2E_TAG) ./tests/...

.PHONY: architecture
architecture:
	@echo "Running Architecture Tests"
	@exit_code=0; \
    for dir in $(MODULES); do \
    	echo "Running architecture tests in $$dir..."; \
    	cd $$dir && $(GO_TEST) -tags=$(ARCHITECTURE_TAG) ./... || exit_code=$$?; \
    	cd - > /dev/null; \
    done; \
    exit $$exit_code

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

CONFIG_PATH = ../../.golangci.yml

.PHONY: lint
lint:
	@exit_code=0; \
    for dir in $(MODULES); do \
    	echo "Linting $$dir..."; \
    	cd $$dir && golangci-lint run --fix --config=$(CONFIG_PATH) || exit_code=$$?; \
    	cd - > /dev/null; \
    done; \
    exit $$exit_code

PROTO_MODULE_PATH = $$(go list -f '{{ .Dir }}' -m github.com/B1lok/proto-contracts)/proto

define GEN_PROTO
	@echo "Generating protos for module $(1)..."
	cd $(1) && \
	protoc --proto_path=$(PROTO_MODULE_PATH) \
		--go_out=paths=source_relative:$(2) \
		--go-grpc_out=paths=source_relative:$(2) \
		$(3)
endef

.PHONY: proto-gen-all
proto-gen-all: proto-gen-subscription proto-gen-weather proto-gen-notification proto-gen-gateway

.PHONY: proto-gen-subscription
proto-gen-subscription:
	$(call GEN_PROTO,apps/subscription,internal/interface/grpc/subscription,subscription.proto)
	$(call GEN_PROTO,apps/subscription,internal/infrastructure/grpc/validator,validator.proto)

.PHONY: proto-gen-weather
proto-gen-weather:
	$(call GEN_PROTO,apps/weather,internal/interface/grpc/weather,weather.proto)
	$(call GEN_PROTO,apps/weather,internal/interface/grpc/validator,validator.proto)

.PHONY: proto-gen-notification
proto-gen-notification:
	$(call GEN_PROTO,apps/notification,internal/grpc/weather,weather.proto)

.PHONY: proto-gen-gateway
proto-gen-gateway:
	$(call GEN_PROTO,apps/gateway,internal/controllers/subscription,subscription.proto)
	$(call GEN_PROTO,apps/gateway,internal/controllers/weather,weather.proto)