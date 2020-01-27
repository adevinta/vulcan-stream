#!/bin/sh

export PORT=${PORT:-8080}
export LOG_LEVEL=${LOG_LEVEL:-Debug}
export PG_PORT=${PG_PORT:-5432}

# Apply env variables
cat config.toml | envsubst > run.toml

if [ ! -z "$PG_CA_B64" ]; then
  echo $PG_CA_B64 | base64 -d > /etc/ssl/certs/pg.crt  # for go app
fi

/app/vulcan-stream run.toml
