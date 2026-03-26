#!/usr/bin/env bash

set -euo pipefail

IMAGE_NAME="${IMAGE_NAME:-telegram-sender-api}"
CURRENT_TAG="${IMAGE_TAG:?IMAGE_TAG is required}"

mapfile -t image_tags < <(docker image ls "$IMAGE_NAME" --format '{{.Tag}}' | awk 'NF && $1 != "<none>"')

mapfile -t ordered_tags < <(
  for tag in "${image_tags[@]}"; do
    if [[ "$tag" == "latest" ]]; then
      continue
    fi

    created="$(docker image inspect "${IMAGE_NAME}:${tag}" --format '{{.Created}}' 2>/dev/null || true)"
    if [[ -n "$created" ]]; then
      printf '%s %s\n' "$created" "$tag"
    fi
  done | sort -r | awk '{print $2}'
)

previous_tag=""
for tag in "${ordered_tags[@]}"; do
  if [[ "$tag" != "$CURRENT_TAG" ]]; then
    previous_tag="$tag"
    break
  fi
done

keep_tags=("latest" "$CURRENT_TAG")
if [[ -n "$previous_tag" ]]; then
  keep_tags+=("$previous_tag")
fi

for tag in "${image_tags[@]}"; do
  keep=false
  for allowed_tag in "${keep_tags[@]}"; do
    if [[ "$tag" == "$allowed_tag" ]]; then
      keep=true
      break
    fi
  done

  if [[ "$keep" == false ]]; then
    docker image rm --no-prune "${IMAGE_NAME}:${tag}" >/dev/null 2>&1 || true
  fi
done

docker image prune -f >/dev/null 2>&1 || true
