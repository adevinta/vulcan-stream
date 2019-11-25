#!/bin/sh

export PORT=${PORT:-8080}
export LOG_LEVEL=${LOG_LEVEL:-Debug}
export PG_PORT=${PG_PORT:-5432}

# Apply env variables
cat config.toml | envsubst > run.toml

/app/vulcan-stream run.toml
