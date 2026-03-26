#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

DEPLOY_ENV_FILE="${DEPLOY_ENV_FILE:-.env}"
COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-telegram-sender-api}"
CONTAINER_NAME="telegram-sender-api"
SERVICE_NAME="telegram-sender-api"

if [[ ! -f "$DEPLOY_ENV_FILE" ]]; then
  echo "compose env file not found: $DEPLOY_ENV_FILE" >&2
  exit 1
fi

set -a
source "$DEPLOY_ENV_FILE"
set +a

: "${APP_BIND_IP:?APP_BIND_IP is required}"
APP_PORT="${APP_PORT:-8086}"

container_id="$(docker ps -aq -f "name=^${CONTAINER_NAME}$" 2>/dev/null || true)"
if [[ -n "$container_id" ]]; then
  existing_project="$(docker inspect -f '{{ index .Config.Labels "com.docker.compose.project" }}' "$container_id" 2>/dev/null || true)"
  existing_service="$(docker inspect -f '{{ index .Config.Labels "com.docker.compose.service" }}' "$container_id" 2>/dev/null || true)"

  if [[ "$existing_project" != "$COMPOSE_PROJECT_NAME" || "$existing_service" != "$SERVICE_NAME" ]]; then
    echo "container ${CONTAINER_NAME} exists but is not managed by ${COMPOSE_PROJECT_NAME}/${SERVICE_NAME}" >&2
    exit 1
  fi
fi

TMP_COMPOSE_FILE="$(mktemp /tmp/telegram-sender-api-compose.XXXXXX.yml)"
cleanup() {
  rm -f "$TMP_COMPOSE_FILE"
}
trap cleanup EXIT

{
  printf 'services:\n'
  printf '  telegram-sender-api:\n'
  printf '    ports:\n'

  IFS=',' read -r -a bind_ips <<< "$APP_BIND_IP"
  for raw_ip in "${bind_ips[@]}"; do
    bind_ip="$(printf '%s' "$raw_ip" | xargs)"
    if [[ -z "$bind_ip" ]]; then
      continue
    fi
    printf '      - "%s:%s:%s"\n' "$bind_ip" "$APP_PORT" "$APP_PORT"
  done
} > "$TMP_COMPOSE_FILE"

if ! grep -q ' - "' "$TMP_COMPOSE_FILE"; then
  echo "APP_BIND_IP must contain at least one IP address" >&2
  exit 1
fi

docker compose \
  -p "$COMPOSE_PROJECT_NAME" \
  --env-file "$DEPLOY_ENV_FILE" \
  -f docker-compose.yml \
  -f "$TMP_COMPOSE_FILE" \
  "$@"
