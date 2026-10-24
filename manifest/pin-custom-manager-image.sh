#!/usr/bin/env bash
set -euo pipefail

container=l4d2-manager
previous_container=l4d2-manager-before-workshop-meta
image=local/l4d2-manager-next:workshop-meta-20260820.3
env_file=/opt/l4d2-manager-custom/manager.env

if docker container inspect "$previous_container" >/dev/null 2>&1; then
  echo "Rollback container ${previous_container} already exists" >&2
  exit 1
fi

install -d -m 0700 /opt/l4d2-manager-custom
docker inspect "$container" | python3 -c '
import json, sys
data = json.load(sys.stdin)[0]["Config"]["Env"]
for item in data:
    if "\n" in item or "\r" in item:
        raise SystemExit("container environment contains unsupported newlines")
    print(item)
' > "$env_file"
chmod 0600 "$env_file"

docker commit \
  --change 'LABEL l4d2.custom.workshop-meta="20260820.3"' \
  "$container" "$image" >/dev/null

restore_previous() {
  docker rm -f "$container" >/dev/null 2>&1 || true
  docker rename "$previous_container" "$container"
  docker start "$container" >/dev/null
}

docker stop "$container" >/dev/null
docker rename "$container" "$previous_container"
docker run -d \
  --name "$container" \
  --network host \
  --restart unless-stopped \
  --env-file "$env_file" \
  --volumes-from "$previous_container" \
  "$image" >/dev/null

for _ in $(seq 1 20); do
  if [[ "$(docker inspect -f '{{.State.Running}}' "$container")" == "true" ]] && \
    curl -sS -o /dev/null --max-time 2 http://127.0.0.1:27020/; then
    printf 'Pinned Manager to %s; rollback container: %s\n' "$image" "$previous_container"
    exit 0
  fi
  sleep 1
done

echo 'Pinned container failed its startup check; restoring previous container' >&2
restore_previous
exit 1
