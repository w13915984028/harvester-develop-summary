#!/usr/bin/env bash

set -euo pipefail

prune_images_by_age() {
    local days_threshold="$1"
    local seconds_threshold=$(( days_threshold * 86400 ))
    local current_time
    current_time=$(date +%s)

    echo "==> Scanning for Docker images older than ${days_threshold} days..."

    # Collect Image IDs currently used by ANY container (running or stopped)
    local used_image_ids
    used_image_ids=$(docker ps -a --format "{{.Image}}" | sort -u)

    local deleted_count=0
    local skipped_count=0

    # Read image details (Format: ID|CreatedAt|Repository:Tag)
    while IFS='|' read -r image_id created_at image_ref; do
        [[ -z "$image_id" ]] && continue

        # Clean "2026-08-04 08:55:17 +0200 CEST" -> "2026-08-04 08:55:17 +0200"
        local clean_date
        clean_date=$(echo "$created_at" | awk '{print $1, $2, $3}')

        local image_time
        image_time=$(date -d "$clean_date" +%s)

        local age_seconds=$(( current_time - image_time ))
        local age_days=$(( age_seconds / 86400 ))

        if [ "$age_seconds" -gt "$seconds_threshold" ]; then
            # Check if image is referenced by a container
            if echo "$used_image_ids" | grep -q "$image_id"; then
                echo "[SKIP] $image_ref ($image_id) is $age_days days old, but currently used by a container."
                skipped_count=$(( skipped_count + 1 ))
            else
                echo "[DELETE] $image_ref ($image_id) is $age_days days old. Removing..."
                if docker rmi "$image_id" >/dev/null 2>&1; then
                    deleted_count=$(( deleted_count + 1 ))
                else
                    # Fallback to repository:tag if image ID removal fails due to multi-tag references
                    if [ "$image_ref" != "<none>:<none>" ] && docker rmi "$image_ref" >/dev/null 2>&1; then
                        deleted_count=$(( deleted_count + 1 ))
                    else
                        echo "       └─ Failed to remove $image_ref (may have dependent child layers)."
                    fi
                fi
            fi
        fi
    done < <(docker images --format "{{.ID}}|{{.CreatedAt}}|{{.Repository}}:{{.Tag}}")

    echo ""
    echo "==> Summary:"
    echo "    Deleted: $deleted_count image(s)"
    echo "    Skipped (in-use): $skipped_count image(s)"

    # Prune all associated BuildKit cache layers
    echo ""
    echo "==> Pruning all builder cache..."
    docker builder prune -a -f
}

main() {
    # Default to 30 days if no argument is passed
    local days_threshold="${1:-30}"

    # Validate that input is an integer
    if ! [[ "$days_threshold" =~ ^[0-9]+$ ]]; then
        echo "Error: Days threshold must be a positive integer." >&2
        exit 1
    fi

    prune_images_by_age "$days_threshold"
}

# ------------------------------------------------------------------------------
# Entry Point
# ------------------------------------------------------------------------------

main "$@"

