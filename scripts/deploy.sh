#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

IMAGE_NAME="${IMAGE_NAME:-telegram-sender-api}"
IMAGE_TAG="${IMAGE_TAG:?IMAGE_TAG is required}"
DEPLOY_ENV_FILE="${DEPLOY_ENV_FILE:-/opt/telegram-sender-api/.env}"

if [[ ! -f "$DEPLOY_ENV_FILE" ]]; then
  echo "deploy env file not found: $DEPLOY_ENV_FILE" >&2
  exit 1
fi

export IMAGE_NAME
export IMAGE_TAG
export DEPLOY_ENV_FILE

bash scripts/compose.sh build telegram-sender-api
docker tag "${IMAGE_NAME}:${IMAGE_TAG}" "${IMAGE_NAME}:latest"
bash scripts/compose.sh up -d --force-recreate telegram-sender-api
