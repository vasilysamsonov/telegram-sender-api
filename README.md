# Telegram Sender API

Telegram Sender API is a small HTTP service that accepts a Telegram bot token, a target chat ID, and a text message, then forwards the message to the Telegram Bot API.

The service is intentionally focused on one job:

- expose a simple REST endpoint for sending Telegram messages
- validate incoming requests
- return safe HTTP errors to clients
- emit structured logs
- attach a request ID to every HTTP response for log correlation
- shut down gracefully on `SIGINT` and `SIGTERM`

## Features

- Built with Go `1.26`
- HTTP server based on Fiber `v2`
- Environment-based configuration via `github.com/caarlos0/env/v11`
- Structured logging via `zerolog`
- Automatic request ID middleware
- Panic recovery middleware
- Safe error responses without leaking internal implementation details
- Unit tests for the HTTP layer, use case logic, and Telegram web API adapter behavior

## Use Case

This service is useful when:

- another backend needs a simple internal API for Telegram delivery
- you want to centralize Telegram send logic behind one HTTP endpoint
- different callers use different Telegram bot tokens
- you need one place to apply validation, logging, error mapping, and operational controls

## Requirements

- Go `1.26+`
- network access from the running service to `https://api.telegram.org`
- a valid Telegram bot token
- a valid Telegram chat ID that the bot can send messages to

## Project Layout

```text
cmd/app
config
internal/app
internal/controller/http
internal/entity
internal/repo
internal/repo/webapi/telegram
internal/usecase
internal/usecase/message
pkg/logger
```

High-level responsibilities:

- `cmd/app`: process entrypoint
- `config`: environment configuration
- `internal/app`: application bootstrap and lifecycle
- `internal/controller/http`: HTTP routes, middleware, request/response handling
- `internal/entity`: core data structures
- `internal/repo`: interfaces for external dependencies
- `internal/repo/webapi/telegram`: Telegram Bot API integration
- `internal/usecase/message`: message sending workflow and validation
- `pkg/logger`: structured logging wrapper

## Configuration

Configuration is loaded from environment variables. On startup, the service also tries to load values from a local `.env` file automatically.

Supported variables:

| Variable | Default | Description |
|---|---:|---|
| `APP_BIND_IP` | none | one IP or a comma-separated list of host IPs used to publish the service port |
| `APP_PORT` | `8086` | HTTP server port |
| `HTTP_TIMEOUT` | `10s` | outbound HTTP timeout for Telegram API calls |
| `SHUTDOWN_TIMEOUT` | `5s` | graceful shutdown timeout |
| `LOG_LEVEL` | `info` | logger level: `debug`, `info`, `warn`, `error` |

Example:

```env
APP_BIND_IP=127.0.0.1,10.0.0.5
APP_PORT=8086
HTTP_TIMEOUT=10s
SHUTDOWN_TIMEOUT=5s
LOG_LEVEL=info
```

Use `.env.example` as the starting point for your local `.env`.

## Running Locally

### 1. Install dependencies

```bash
go mod tidy
```

### 2. Configure the environment

You can either export variables manually:

```bash
export APP_BIND_IP=127.0.0.1,10.0.0.5
export APP_PORT=8086
export HTTP_TIMEOUT=10s
export SHUTDOWN_TIMEOUT=5s
export LOG_LEVEL=debug
```

Or create a local `.env` file based on `.env.example`.

### 3. Start the service

```bash
go run ./cmd/app
```

The server will listen on:

```text
http://localhost:8086
```

## Build

```bash
go build -o telegram-sender-api ./cmd/app
```

## Docker

### Build the image

```bash
docker build -t telegram-sender-api .
```

### Run the container directly

```bash
docker run --rm \
  --env-file .env \
  -p 8086:8086 \
  telegram-sender-api
```

If you use a different port in `.env`, adjust the published port accordingly.

## Docker Compose

The repository includes a base `docker-compose.yml` and a helper wrapper `scripts/compose.sh`.

The wrapper reads `APP_BIND_IP` and generates the correct Docker Compose port mappings. This is required because standard Compose variable interpolation cannot expand a comma-separated IP list into multiple `ports` entries.
It also fixes the Compose project name to `telegram-sender-api` and refuses to operate on an existing `telegram-sender-api` container if that container is not owned by the same Compose project and service.

Start the service:

```bash
bash scripts/compose.sh up --build
```

Run in background:

```bash
bash scripts/compose.sh up --build -d
```

Build the image without starting containers:

```bash
bash scripts/compose.sh build
```

Rebuild the image without using cache:

```bash
bash scripts/compose.sh build --no-cache
```

Recreate containers with a fresh build:

```bash
bash scripts/compose.sh up --build --force-recreate -d
```

Stop the service:

```bash
bash scripts/compose.sh stop
```

Stop and remove containers, networks, and compose-managed resources:

```bash
bash scripts/compose.sh down
```

Show running services:

```bash
bash scripts/compose.sh ps
```

Follow logs:

```bash
bash scripts/compose.sh logs -f
```

Follow logs for the application only:

```bash
bash scripts/compose.sh logs -f telegram-sender-api
```

Restart the application container:

```bash
bash scripts/compose.sh restart telegram-sender-api
```

Remove old dangling images after rebuilds:

```bash
docker image prune -f
```

Compose behavior:

- builds the application from the local repository
- loads configuration from `.env`
- publishes every IP from `APP_BIND_IP` to the same container port
- restarts the container with `unless-stopped`

`APP_BIND_IP` is required. It can contain:

- one IP, for example `127.0.0.1`
- multiple IPs separated by commas, for example `127.0.0.1,10.0.0.5`

If it is not set, startup fails fast instead of silently binding to an unintended interface. `APP_PORT` still falls back to `8086`.

## Deployment With GitHub Actions

The repository includes a GitHub Actions workflow for tag-based deployment to a server running a self-hosted runner:

- workflow file: `.github/workflows/deploy.yml`
- deploy script: `scripts/deploy.sh`
- image cleanup script: `scripts/cleanup_images.sh`

### Deployment Model

The deployment flow is:

1. push a Git tag
2. GitHub Actions starts on the self-hosted runner
3. the runner checks out the tagged revision on the target server
4. the runner builds a Docker image tagged with the Git tag
5. the application is restarted through `scripts/compose.sh`
6. the workflow verifies `http://<first-bind-ip>:${APP_PORT}/healthz` on the target server
7. old image versions are deleted, keeping only:
   - `latest`
   - the current tag
   - the previous tag

This means everything older than the previous deployed version is removed from local Docker images on the server.

### What This Deployment Strategy Assumes

This setup assumes:

- the GitHub self-hosted runner is installed on the target deployment server
- Docker and Docker Compose are available on that same server
- the runner user is allowed to run Docker commands
- the repository is checked out by the runner during each workflow execution
- the server keeps a persistent environment file outside the repository workspace

This is a simple server-side deployment model:

- no image registry is required
- the build happens directly on the server
- the running container is replaced on each tagged release

### Requirements For The Server

The target server must have:

- Docker installed
- Docker Compose installed
- a GitHub self-hosted runner installed and connected to the repository
- an environment file available on the server at:

```text
/opt/telegram-sender-api/.env
```

The deploy script uses that file by default.

### Recommended Server Layout

A practical layout on the target server:

```text
/opt/telegram-sender-api/.env
/home/<runner-user>/actions-runner
```

Notes:

- `.env` is stored outside the repository checkout
- the repository workspace is managed by the GitHub Actions runner
- the deployment script reads runtime configuration from `/opt/telegram-sender-api/.env`

### Example Server `.env`

Example server configuration:

```env
APP_PORT=8092
HTTP_TIMEOUT=10s
SHUTDOWN_TIMEOUT=5s
LOG_LEVEL=info
```

### Self-Hosted Runner

The workflow currently targets:

```yaml
runs-on:
  - self-hosted
  - linux
```

If your runner uses additional labels, update `.github/workflows/deploy.yml` accordingly.

### Installing The Self-Hosted Runner

On the target server:

1. open the repository in GitHub
2. go to `Settings -> Actions -> Runners`
3. click `New self-hosted runner`
4. choose Linux
5. use the exact download URL and version shown by GitHub for your repository

GitHub Actions runners are released progressively. Do not hardcode the version blindly. Always prefer the commands shown in the repository runner setup screen, because the latest global runner release may not be available to your repository yet.

Example Linux installation flow:

If you want the runner under `/opt/github`, verify permissions first:

```bash
ls -ld /opt /opt/github
```

If `/opt/github` does not exist or is not writable by the runner user, create it with `sudo` and assign ownership:

```bash
sudo mkdir -p /opt/github
sudo chown <runner-user>:<runner-user> /opt/github
```

Then install the runner there:

```bash
mkdir -p /opt/github/telegram-sender-api-runner
cd /opt/github/telegram-sender-api-runner
curl -o actions-runner-linux-x64-2.333.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.333.0/actions-runner-linux-x64-2.333.0.tar.gz
echo "7ce6b3fd8f879797fcc252c2918a23e14a233413dc6e6ab8e0ba8768b5d54475  actions-runner-linux-x64-2.333.0.tar.gz" | shasum -a 256 -c
tar xzf ./actions-runner-linux-x64-2.333.0.tar.gz
./config.sh --url https://github.com/<owner>/<repo> --token <runner-token>
./run.sh
```

Replace:

- `<owner>/<repo>` with your GitHub repository path
- `<runner-token>` with the registration token generated in the GitHub UI
- `<runner-user>` with the Linux user that will own and run the runner service

For a persistent production setup, install the runner as a service using the commands GitHub provides after configuration:

```bash
sudo ./svc.sh install
sudo ./svc.sh start
```

After installation, verify that the runner appears as `Idle` in `Settings -> Actions -> Runners`.

### Docker Permissions For The Runner

The runner user must be able to execute:

- `docker version`
- `docker compose version`
- `docker compose build`
- `docker compose up`
- `docker image ls`
- `docker rmi`

Typical Linux setup:

```bash
sudo usermod -aG docker <runner-user>
```

After changing group membership, restart the session or the runner service.

### How To Deploy

Create a tag and push it:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Or push an already existing local tag:

```bash
git push origin --tags
```

The workflow is configured for version tags that start with `v`.

Examples:

- `v0.1.0`
- `v1.2.3`
- `v2026.03.25`

Tags that do not start with `v` do not trigger deployment.

### What The Workflow Uses

The deploy workflow sets:

- `IMAGE_NAME=telegram-sender-api`
- `IMAGE_TAG=<git tag name>`
- `DEPLOY_ENV_FILE=/opt/telegram-sender-api/.env`
- `APP_BIND_IP=<GitHub Actions variable APP_BIND_IP>`

`APP_BIND_IP` is expected to be stored in GitHub repository or environment variables, not in the runner workspace.
It may contain one IP or a comma-separated list of IPs.

The compose file then uses those values as:

```yaml
image: "${IMAGE_NAME:-telegram-sender-api}:${IMAGE_TAG:-local}"
```

The deploy script exports `DEPLOY_ENV_FILE` before calling `scripts/compose.sh`, so the same server env file is used consistently by:

- the compose `env_file` directive
- the image build and recreate commands
- the post-deploy health check

The workflow also validates that `APP_BIND_IP` is present before deployment starts.

### Deployment Commands Used On The Runner

The deploy step performs these operations:

```bash
bash scripts/compose.sh build telegram-sender-api
docker tag telegram-sender-api:<tag> telegram-sender-api:latest
bash scripts/compose.sh up -d --force-recreate telegram-sender-api
```

Then the workflow runs a separate cleanup step:

```bash
bash scripts/cleanup_images.sh
```

Between deploy and cleanup, the workflow also verifies service health with:

```bash
source "$DEPLOY_ENV_FILE"
HEALTHCHECK_HOST="${APP_BIND_IP%%,*}"
HEALTHCHECK_HOST="${HEALTHCHECK_HOST// /}"
if [[ "$HEALTHCHECK_HOST" == "0.0.0.0" ]]; then
  HEALTHCHECK_HOST="127.0.0.1"
fi
curl --fail --silent --show-error "http://${HEALTHCHECK_HOST}:${APP_PORT}/healthz"
```

### What The Cleanup Script Removes

After a successful deploy, the cleanup script keeps only:

- `telegram-sender-api:latest`
- `telegram-sender-api:<current-tag>`
- `telegram-sender-api:<previous-tag>` if one exists

All older local tagged images for `telegram-sender-api` are removed.

The previous tag is selected by image creation time, excluding:

- `latest`
- the current tag

This means:

- you keep the currently deployed version
- you keep one previous version for quick manual rollback
- disk usage does not grow indefinitely

### If You Need A Different Env File Path

The deploy script supports overriding the env file path with:

```bash
DEPLOY_ENV_FILE=/path/to/.env
```

If needed, add that variable to the workflow before the deploy step.

### Verifying A Deployment

After pushing a tag:

1. open the GitHub Actions run and confirm the workflow succeeded
2. on the server, confirm the container is running:

```bash
bash scripts/compose.sh ps
```

3. inspect logs:

```bash
bash scripts/compose.sh logs -f telegram-sender-api
```

4. verify health:

```bash
curl http://localhost:<APP_PORT>/healthz
```

5. verify the expected image tags exist:

```bash
docker image ls telegram-sender-api
```

### Manual Rollback

The automated workflow does not perform rollback on failure. Rollback remains a manual operation.

Because the cleanup script keeps the previous tag, a simple rollback path is still available.

Example manual rollback:

```bash
docker tag telegram-sender-api:<previous-tag> telegram-sender-api:latest
IMAGE_NAME=telegram-sender-api IMAGE_TAG=<previous-tag> DEPLOY_ENV_FILE=/opt/telegram-sender-api/.env bash scripts/compose.sh up -d --force-recreate telegram-sender-api
```

Replace `<previous-tag>` with the actual tag you want to restore.

### Common Deployment Failure Cases

Typical causes of failed deployment:

- self-hosted runner is offline
- runner user has no Docker access
- `/opt/telegram-sender-api/.env` does not exist
- `APP_PORT` is missing or invalid in the server env file
- the selected port is already occupied on the server
- Docker daemon is unavailable
- outbound network access is blocked during runtime

### Operational Notes For This Setup

This deployment approach is intentionally simple and server-centric.

Good fit for:

- one server
- one self-hosted runner
- one service
- low operational overhead

Things it does not do:

- blue/green deployment
- canary rollout
- automatic rollback
- registry-based artifact promotion
- multi-server orchestration

If the service grows beyond a single-host deployment, you will likely want a registry-based build pipeline and a more explicit rollout strategy.

## Testing

Run all tests:

```bash
go test ./...
```

## HTTP API

### Health Check

`GET /healthz`

Response:

```http
HTTP/1.1 200 OK
```

This endpoint is intended for liveness/readiness checks.

### Send Message

`POST /v1/messages/send`

Request body:

```json
{
  "bot_token": "123456:ABCDEF",
  "chat_id": 123456789,
  "text": "Hello from Telegram Sender API"
}
```

Request fields:

| Field | Type | Required | Description |
|---|---|---|---|
| `bot_token` | `string` | yes | Telegram bot token used for the outgoing request |
| `chat_id` | `integer` | yes | target Telegram chat ID |
| `text` | `string` | yes | text message to send |

Successful response:

```json
{
  "status": "ok"
}
```

### Example `curl`

```bash
curl --request POST \
  --url http://localhost:8086/v1/messages/send \
  --header 'Content-Type: application/json' \
  --data '{
    "bot_token": "123456:ABCDEF",
    "chat_id": 123456789,
    "text": "Hello from Telegram Sender API"
  }'
```

## Error Handling

The API returns safe, client-facing errors. Internal stack details, transport details, and upstream implementation details are not returned in the response body.

Common response format:

```json
{
  "status": "error",
  "error": "..."
}
```

### Status Codes

| Status | When |
|---:|---|
| `200` | message accepted and successfully sent via Telegram |
| `400` | request JSON is invalid or required request data is missing |
| `413` | request body is larger than allowed |
| `500` | unexpected internal failure |
| `502` | Telegram API request failed or Telegram returned an unsuccessful result |

### Typical Error Bodies

Invalid JSON:

```json
{
  "status": "error",
  "error": "invalid json body"
}
```

Multiple JSON objects in one request:

```json
{
  "status": "error",
  "error": "request body must contain a single JSON object"
}
```

Validation error:

```json
{
  "status": "error",
  "error": "invalid request data"
}
```

Telegram API failure:

```json
{
  "status": "error",
  "error": "failed to send message"
}
```

Unexpected internal error:

```json
{
  "status": "error",
  "error": "internal server error"
}
```

## Request ID

Every request gets a request ID.

The service includes an `X-Request-ID` response header, which can be used to correlate:

- client-side errors
- server-side logs
- panic recovery logs
- upstream call failures

Example:

```http
X-Request-ID: 94d97a14-7a63-4f4a-a210-8e53d1ecdb0a
```

## Logging

The service uses structured logging with `zerolog`.

Logs include:

- timestamp
- log level
- caller information
- request ID in HTTP middleware and controller error logs

What is logged:

- incoming HTTP requests with method, path, status, and duration
- application shutdown signal reception
- controller-level send failures
- panic recovery events

What is not exposed to clients:

- internal error chains
- upstream transport details
- implementation-level stack context

These details remain in logs instead.

## Shutdown Behavior

The process listens for:

- `SIGINT`
- `SIGTERM`

On shutdown:

1. the server stops accepting new requests
2. Fiber attempts graceful shutdown
3. shutdown is bounded by `SHUTDOWN_TIMEOUT`

## Operational Notes

This is a focused MVP. It is production-oriented in the sense that it already includes:

- config validation from env
- graceful shutdown
- request ID support
- structured logs
- panic recovery
- safe HTTP error messages
- adapter tests and route tests

Things that are intentionally not part of the current service:

- authentication or authorization on the HTTP endpoint
- rate limiting
- retries with backoff for Telegram
- message persistence
- message queueing
- delivery history
- metrics export
- OpenAPI / Swagger docs

If you expose this service to untrusted clients, you should add:

- authentication
- rate limiting
- request size controls at the edge
- network-level restrictions
- secret handling policies for incoming bot tokens

## Telegram Notes

This service forwards requests to the Telegram Bot API `sendMessage` endpoint.

Important assumptions:

- the bot token is provided by the caller on every request
- the bot has permission to send to the target chat
- the `chat_id` is valid for that bot
- the text is sent as-is

The service does not currently:

- escape Markdown/HTML
- choose parse mode
- support captions, keyboards, media, or attachments
- split oversized Telegram messages

## Troubleshooting

### `400 invalid json body`

Check:

- `Content-Type: application/json`
- valid JSON syntax
- correct field names

### `400 invalid request data`

Check:

- `bot_token` is present and not empty
- `chat_id` is non-zero
- `text` is present and not empty

### `502 failed to send message`

Typical causes:

- invalid bot token
- bot has no access to the target chat
- target chat does not exist
- Telegram API is unavailable
- network timeout on outbound request

### Service exits during startup

Check:

- env values are valid
- chosen `APP_PORT` is free
- process can bind to the selected port
