#!/usr/bin/env bash

set -euo pipefail

IMAGE_NAME="${IMAGE_NAME:-telegram-sender-api}"
CURRENT_TAG="${IMAGE_TAG:?IMAGE_TAG is required}"

previous_tag="$(
  docker image ls "$IMAGE_NAME" --format '{{.Tag}}' \
    | awk 'NF && $1 != "<none>" && $1 != "latest"' \
    | while read -r tag; do
        created="$(docker image inspect "${IMAGE_NAME}:${tag}" --format '{{.Created}}' 2>/dev/null || true)"
        if [[ -n "$created" ]]; then
          printf '%s %s\n' "$created" "$tag"
        fi
      done \
    | sort -r \
    | awk -v current="$CURRENT_TAG" '$2 != current { print $2; exit }'
)"

if [[ -z "$previous_tag" ]]; then
  echo "previous image tag not found for ${IMAGE_NAME}, current tag: ${CURRENT_TAG}" >&2
  exit 1
fi

printf '%s\n' "$previous_tag"
