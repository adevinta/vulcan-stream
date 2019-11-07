#!/bin/sh

if [ -f "$1" ]; then
  # Force logs to STDOUT
  cat $1 | sed 's/LogFile *=.*/LogFile = ""/g' > config.toml
else
  echo "ERROR: Expected config file"
  echo "Usage: $0 config.toml"
fi

/app/vulcan-stream config.toml
