SHELL := bash
.SHELLFLAGS := -e -o pipefail -c
.ONESHELL:

.DEFAULT_GOAL := help

APP_NAME ?= telegram-sender-api
SERVICE_NAME ?= $(APP_NAME)
COMPOSE_PROJECT_NAME ?= $(APP_NAME)
IMAGE_NAME ?= $(APP_NAME)
IMAGE_TAG ?= local
DEPLOY_ENV_FILE ?= .env
APP_PORT ?= 8086
APP_BIND_IP ?=
HEALTHCHECK_RETRIES ?= 30
HEALTHCHECK_DELAY ?= 1
GO_MAIN ?= ./cmd/app
BINARY_PATH ?= ./bin/$(APP_NAME)
GO_CACHE_DIR ?= $(CURDIR)/.cache/go-build

export APP_NAME
export SERVICE_NAME
export COMPOSE_PROJECT_NAME
export IMAGE_NAME
export IMAGE_TAG
export DEPLOY_ENV_FILE
export APP_PORT
export APP_BIND_IP
export GOCACHE := $(GO_CACHE_DIR)

define require_command
command -v $(1) >/dev/null 2>&1 || { echo "required command not found: $(1)" >&2; exit 1; }
endef

define require_file
[[ -f "$(1)" ]] || { echo "required file not found: $(1)" >&2; exit 1; }
endef

.PHONY: \
	help \
	print-vars \
	preflight \
	verify-deploy-env \
	verify-docker \
	fmt \
	verify-format \
	test \
	check \
	build \
	run \
	docker-build \
	docker-tag-latest \
	compose-build \
	compose-up \
	compose-down \
	compose-stop \
	compose-ps \
	compose-logs \
	compose-restart \
	healthcheck \
	smoke \
	previous-tag \
	cleanup-images \
	deploy \
	deploy-release \
	rollback \
	rollback-release

help: ## Show available targets
	awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_.-]+:.*## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

print-vars: ## Print resolved operational variables
	printf '%s=%s\n' "APP_NAME" "$(APP_NAME)"
	printf '%s=%s\n' "SERVICE_NAME" "$(SERVICE_NAME)"
	printf '%s=%s\n' "COMPOSE_PROJECT_NAME" "$(COMPOSE_PROJECT_NAME)"
	printf '%s=%s\n' "IMAGE_NAME" "$(IMAGE_NAME)"
	printf '%s=%s\n' "IMAGE_TAG" "$(IMAGE_TAG)"
	printf '%s=%s\n' "DEPLOY_ENV_FILE" "$(DEPLOY_ENV_FILE)"
	printf '%s=%s\n' "APP_PORT" "$(APP_PORT)"
	printf '%s=%s\n' "APP_BIND_IP" "$(APP_BIND_IP)"
	printf '%s=%s\n' "HEALTHCHECK_RETRIES" "$(HEALTHCHECK_RETRIES)"
	printf '%s=%s\n' "HEALTHCHECK_DELAY" "$(HEALTHCHECK_DELAY)"
	printf '%s=%s\n' "GO_MAIN" "$(GO_MAIN)"
	printf '%s=%s\n' "BINARY_PATH" "$(BINARY_PATH)"
	printf '%s=%s\n' "GOCACHE" "$(GOCACHE)"

preflight: ## Verify required local tooling
	$(call require_command,go)
	$(call require_command,docker)

verify-deploy-env: ## Verify deployment env file and required variables
	$(call require_file,$(DEPLOY_ENV_FILE))
	set -a
	source "$(DEPLOY_ENV_FILE)"
	set +a
	: "$${APP_BIND_IP:=$(APP_BIND_IP)}"
	: "$${APP_BIND_IP:?APP_BIND_IP is required}"

verify-docker: ## Print Docker and Docker Compose versions
	$(call require_command,docker)
	docker version
	docker compose version

fmt: ## Format Go code
	$(call require_command,gofmt)
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

verify-format: ## Verify that Go code is formatted
	$(call require_command,gofmt)
	unformatted="$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))"
	[[ -z "$$unformatted" ]] || { echo "gofmt check failed for:" >&2; printf '%s\n' "$$unformatted" >&2; exit 1; }

test: ## Run unit tests
	$(call require_command,go)
	mkdir -p "$(GOCACHE)"
	go test ./...

check: verify-format test build ## Run the local quality gate

build: ## Build the application binary
	$(call require_command,go)
	mkdir -p "$(GOCACHE)"
	mkdir -p "$(dir $(BINARY_PATH))"
	go build -o "$(BINARY_PATH)" "$(GO_MAIN)"

run: ## Run the application locally
	$(call require_command,go)
	mkdir -p "$(GOCACHE)"
	go run "$(GO_MAIN)"

docker-build: verify-docker ## Build the Docker image for the current IMAGE_TAG
	docker build --tag "$(IMAGE_NAME):$(IMAGE_TAG)" .

docker-tag-latest: ## Tag the current IMAGE_TAG as latest
	docker image inspect "$(IMAGE_NAME):$(IMAGE_TAG)" >/dev/null
	docker tag "$(IMAGE_NAME):$(IMAGE_TAG)" "$(IMAGE_NAME):latest"

compose-build: verify-deploy-env verify-docker ## Build the compose service image
	bash scripts/compose.sh build "$(SERVICE_NAME)"

compose-up: verify-deploy-env verify-docker ## Start or recreate the compose service
	bash scripts/compose.sh up -d --force-recreate "$(SERVICE_NAME)"

compose-down: verify-deploy-env verify-docker ## Stop and remove compose resources
	bash scripts/compose.sh down

compose-stop: verify-deploy-env verify-docker ## Stop the compose service
	bash scripts/compose.sh stop

compose-ps: verify-deploy-env verify-docker ## Show compose service status
	bash scripts/compose.sh ps

compose-logs: verify-deploy-env verify-docker ## Follow compose logs
	bash scripts/compose.sh logs -f "$(SERVICE_NAME)"

compose-restart: verify-deploy-env verify-docker ## Restart the compose service
	bash scripts/compose.sh restart "$(SERVICE_NAME)"

healthcheck: verify-deploy-env ## Check the deployed HTTP health endpoint
	$(call require_command,curl)
	set -a
	source "$(DEPLOY_ENV_FILE)"
	set +a
	app_bind_ip="$${APP_BIND_IP:=$(APP_BIND_IP)}"
	app_port="$${APP_PORT:-$(APP_PORT)}"
	healthcheck_host="$${app_bind_ip%%,*}"
	healthcheck_host="$${healthcheck_host// /}"
	if [[ "$$healthcheck_host" == "0.0.0.0" ]]; then
		healthcheck_host="127.0.0.1"
	fi
	healthcheck_url="http://$${healthcheck_host}:$${app_port}/healthz"
	for _ in $$(seq 1 "$(HEALTHCHECK_RETRIES)"); do
		if curl --fail --silent --show-error "$$healthcheck_url" >/dev/null; then
			echo "health check passed: $$healthcheck_url"
			exit 0
		fi
		sleep "$(HEALTHCHECK_DELAY)"
	done
	echo "health check failed: $$healthcheck_url" >&2
	exit 1

smoke: healthcheck ## Run a post-deploy smoke check

previous-tag: verify-docker ## Print the previous Docker image tag for rollback
	bash scripts/previous_image_tag.sh

cleanup-images: verify-docker ## Remove old Docker images, keep current, previous, and latest
	bash scripts/cleanup_images.sh

deploy: verify-deploy-env docker-build docker-tag-latest compose-up ## Build and deploy the application

deploy-release: deploy healthcheck cleanup-images ## Execute the full production deployment flow

rollback: verify-deploy-env verify-docker ## Roll back the running service to the previous Docker image tag
	current_tag="$(IMAGE_TAG)"
	previous_tag="$$(bash scripts/previous_image_tag.sh)"
	echo "rolling back $(SERVICE_NAME) from $$current_tag to $$previous_tag"
	IMAGE_TAG="$$previous_tag" bash scripts/compose.sh up -d --force-recreate "$(SERVICE_NAME)"

rollback-release: rollback healthcheck ## Roll back and verify service health
