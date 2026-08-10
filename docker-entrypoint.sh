#!/bin/sh
# Container entrypoint for the `karel` service.
#
# Production (LITESTREAM_ENABLED=true, the default — Coolify): restore the SQLite
# database from Cloudflare R2 if the local file is absent (a fresh volume or a
# rebuilt image), then run the app under `litestream replicate -exec` so every
# write is streamed to R2. Litestream exits when the app exits, so Coolify sees a
# normal process lifecycle. The initial migration seeds the singletons, so a
# restored build never double-seeds.
#
# Local (LITESTREAM_ENABLED=false, docker-compose): run the app directly with no
# R2 dependency — the offline end-to-end harness.
set -eu

: "${KAREL_DB_PATH:?KAREL_DB_PATH must be set}"
mkdir -p "$(dirname "${KAREL_DB_PATH}")"

if [ "${LITESTREAM_ENABLED:-true}" = "true" ]; then
  echo "entrypoint: restoring ${KAREL_DB_PATH} from R2 if it does not exist"
  litestream restore -if-db-not-exists -if-replica-exists -config /etc/litestream.yml "${KAREL_DB_PATH}"
  echo "entrypoint: starting karel under 'litestream replicate -exec'"
  exec litestream replicate -config /etc/litestream.yml -exec "karel"
else
  echo "entrypoint: LITESTREAM_ENABLED=false — starting karel without replication"
  exec karel
fi
