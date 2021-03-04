[![Build Status](https://travis-ci.org/adevinta/vulcan-stream.svg?branch=master)](https://travis-ci.org/adevinta/vulcan-stream)

# Vulcan Stream

Vulcan Stream provides a channel of communication between Vulcan Scan Engine and the Vulcan Agents.

Vulcan Scan Engine requires broadcast communication with Vulcan Agents in order to manage the Agent pool and control checks in execution. Because Agents might not be reachable from the internet, the Stream provides a websocket stream that Agents connect to in order to receive input from the Scan Engine.

### Requirements
Vulcan Stream works on top of two main services:
- [Go Websocket Events](https://github.com/danfaizer/gowse)
- [Redis](https://redis.io/)

### Constraints
Current implementation of vulcan-stream must be deployed as a single instance.
The reason for this is we took a design decision to maintain a local in memory cache to speed up checks endpoint requests so we could maximize Vulcan agents performance, which have to query this endpoint before executing each check.

### API
Vulcan Stream exposes two endpoints to abort and retrieve the list of aborted checks.

Abort checks:
```
curl -X POST https://stream.vulcan.com/abort -H "Content-Type: application/json" -d '{"checks": ["<check_id1>", "<check_id2>", ... ]}'
->
<-
200 OK
```

Get checks:
```
curl -X GET https://stream.vulcan.com/checks
->
<-
200 OK 
["<check_id1>", "<check_id2>", ...]
...
```

### Build & Run

Two binaries are provided:
- vulcan-stream
- vulcan-stream-test-client

Assuming you have Docker in your machine and there are no services listening on ports `6379` or `8080`.

Run vulcan-stream:

```
go get -x github.com/adevinta/vulcan-stream/cmd/vulcan-stream

docker run -d -p 6379:6379 redis

vulcan-stream ${GOPATH}/src/github.com/adevinta/vulcan-stream/_resources/config/local.toml
```

Run vulcan-stream websocket client integration test:

```
go get -x github.com/adevinta/vulcan-stream/cmd/vulcan-stream-test-client

vulcan-stream-test-client ${GOPATH}/src/github.com/adevinta/vulcan-stream.git/_resources/config/local.toml
```

Or, connect to the stream and push some messages:

```
curl --include --no-buffer --header "Connection: Upgrade" --header "Upgrade: websocket" \
        --header "Host: localhost:8080" --header "Origin: http://localhost:8080" \
        --header "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" --header "Sec-WebSocket-Version: 13" \
        "http://localhost:8080/stream" &

curl -X POST http://localhost:8080/abort -H 'Content-Type: application/json' -d '{"checks": ["00000000-0000-0000-0000-000000000000"]}'
```

### Configure

You can see and modify Vulcan Stream configuration as required:

`_resources/config/local.toml`


# Docker execute

These are the variables you have to setup:

|Variable|Description|Sample|
|---|---|---|
|PORT|Listen http port|8080|
|LOG_LEVEL||DEBUG|
|REDIS_(HOST\|PORT\|USR\|PWD\|PORT\DB)|Redis variables||
|REDIS_TTL|TTL to apply for aborted check entries|7 days|

```bash
docker build . -t vs

# Use the default config.toml customized with env variables.
docker run --env-file ./local.env vs

# Use custom config.toml
docker run -v `pwd`/custom.toml:/app/config.toml vs
```
