#!/usr/bin/env bash
set -euo pipefail

container=l4d2-manager
artifact=/tmp/l4d2-manager-linux-amd64
static_archive=/tmp/manager-static.tar.gz
backup_root=/opt/l4d2-manager-custom/backups
backup_dir="${backup_root}/$(date +%Y%m%d-%H%M%S)"
static_stage=/tmp/l4d2-manager-static-new

test -x "$artifact"
test -f "$static_archive"
mkdir -p "$backup_dir" "$static_stage"
tar -xzf "$static_archive" -C "$static_stage"

docker cp "${container}:/l4d2-manager" "${backup_dir}/l4d2-manager"
docker cp "${container}:/static" "${backup_dir}/static"
docker inspect "$container" > "${backup_dir}/container-inspect.json"

restore_manager() {
  docker stop "$container" >/dev/null 2>&1 || true
  docker cp "${backup_dir}/l4d2-manager" "${container}:/l4d2-manager"
  docker cp "${backup_dir}/static/." "${container}:/static"
  docker start "$container" >/dev/null
}

docker stop "$container" >/dev/null
docker cp "$artifact" "${container}:/l4d2-manager"
docker cp "${static_stage}/static/." "${container}:/static"
docker start "$container" >/dev/null

for _ in $(seq 1 20); do
  if [[ "$(docker inspect -f '{{.State.Running}}' "$container")" == "true" ]]; then
    if curl -sS -o /dev/null --max-time 2 http://127.0.0.1:27020/; then
      printf 'Manager deployed; backup: %s\n' "$backup_dir"
      exit 0
    fi
  fi
  sleep 1
done

echo 'Manager failed its startup check; restoring backup' >&2
restore_manager
exit 1
