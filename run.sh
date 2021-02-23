#!/bin/sh

export PORT=${PORT:-8080}
export LOG_LEVEL=${LOG_LEVEL:-Debug}
export DOGSTATSD_ENABLED=${DOGSTATSD_ENABLED:-false}
export REDIS_PORT=${REDIS_PORT:-6379}
export REDIS_DB=${REDIS_DB:-0}
export REDIS_TTL=${REDIS_TTL:-0}

# Apply env variables
cat config.toml | envsubst > run.toml

if [ ! -z "$PG_CA_B64" ]; then
  echo $PG_CA_B64 | base64 -d > /etc/ssl/certs/pg.crt  # for go app
fi

sleep 999999
/app/vulcan-stream run.toml
