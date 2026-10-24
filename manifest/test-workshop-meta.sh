#!/usr/bin/env bash
set -euo pipefail

manager=http://127.0.0.1:27020
share=http://127.0.0.1:27830
test_map=meta_integration_test.vpk
renamed_map=meta_integration_renamed.vpk
addons=/opt/l4d2/server/left4dead2/addons
password=$(docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' l4d2-manager \
  | sed -n 's/^L4D2_MANAGER_PASSWORD=//p' | head -n 1)
auth="Authorization: Bearer ${password}"

cleanup() {
  curl -fsS -H "$auth" -X POST -F "map=${test_map}" "${manager}/remove" >/dev/null 2>&1 || true
  curl -fsS -H "$auth" -X POST -F "map=${renamed_map}" "${manager}/remove" >/dev/null 2>&1 || true
}
trap cleanup EXIT
cleanup

old_summary=$(curl -fsS -H "$auth" -H 'Content-Type: application/json' -X POST \
  --data '{"maps":["openroad.vpk"]}' "${manager}/maps/summary")
python3 -c 'import json,sys; item=json.load(sys.stdin)["items"]["openroad.vpk"]; assert item.get("workshop", {}).get("workshop_id")=="123734884"' <<<"$old_summary"
old_catalog=$(curl -fsS "${share}/api/v1/catalog")
python3 -c 'import json,sys; item=next(x for x in json.load(sys.stdin)["files"] if x["name"]=="openroad.vpk"); assert item.get("workshop_id")=="123734884" and item.get("preview_url")' <<<"$old_catalog"

parsed=$(curl -fsS -H "$auth" -H 'Content-Type: application/json' -X POST \
  --data '{"url":"123734884"}' "${manager}/download/link/parse")
download_url=$(python3 -c 'import json,sys; print(json.load(sys.stdin)["items"][0]["file_url"])' <<<"$parsed")

# Deliberately omit workshop_meta: the server must recover it from the parse cache.
curl -fsS -H "$auth" -X POST \
  --form-string "url=${download_url}" \
  -F "filename=${test_map}" \
  "${manager}/download/add" >/dev/null

for _ in $(seq 1 60); do
  status=$(curl -fsS -H "$auth" -X POST "${manager}/download/list" \
    | TEST_MAP="$test_map" python3 -c 'import json,os,sys; items=[x for x in json.load(sys.stdin) if x.get("filename")==os.environ["TEST_MAP"]]; print(items[-1].get("status",-1) if items else -1)')
  if [[ "$status" == "2" ]]; then
    break
  fi
  if [[ "$status" == "3" ]]; then
    echo 'download task failed' >&2
    exit 1
  fi
  sleep 1
done

test -f "${addons}/${test_map}"
test -f "${addons}/${test_map%.vpk}.meta"
python3 -c 'import json,sys; d=json.load(open(sys.argv[1], encoding="utf-8")); assert d.get("workshop_id")=="123734884" and d.get("preview_url")' \
  "${addons}/${test_map%.vpk}.meta"

summary=$(curl -fsS -H "$auth" -H 'Content-Type: application/json' -X POST \
  --data "{\"maps\":[\"${test_map}\"]}" "${manager}/maps/summary")
TEST_MAP="$test_map" python3 -c 'import json,os,sys; w=json.load(sys.stdin)["items"][os.environ["TEST_MAP"]]["workshop"]; assert w.get("workshop_id")=="123734884" and w.get("preview_url")' \
  <<<"$summary"

detail=$(curl -fsS -H "$auth" -X POST -F "map=${test_map}" "${manager}/maps/detail")
python3 -c 'import json,sys; w=json.load(sys.stdin)["workshop"]; assert w.get("workshop_id")=="123734884" and w.get("preview_url")' <<<"$detail"

sleep 6
catalog=$(curl -fsS "${share}/api/v1/catalog")
TEST_MAP="$test_map" python3 -c 'import json,os,sys; items=[x for x in json.load(sys.stdin)["files"] if x.get("name")==os.environ["TEST_MAP"]]; assert items and items[0].get("workshop_id")=="123734884" and items[0].get("preview_url")' \
  <<<"$catalog"

curl -fsS -H "$auth" -X POST -F "oldName=${test_map}" -F "newName=${renamed_map}" "${manager}/rename" >/dev/null
test ! -e "${addons}/${test_map}"
test ! -e "${addons}/${test_map%.vpk}.meta"
test -f "${addons}/${renamed_map}"
test -f "${addons}/${renamed_map%.vpk}.meta"

cleanup
test ! -e "${addons}/${renamed_map}"
test ! -e "${addons}/${renamed_map%.vpk}.meta"
trap - EXIT

echo 'Workshop metadata integration test passed'
