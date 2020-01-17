[![Build Status](https://travis-ci.org/adevinta/vulcan-stream.svg?branch=master)](https://travis-ci.org/adevinta/vulcan-stream)

# Vulcan Stream

Vulcan Stream provides a one-way communication channel from Vulcan Persistence to Vulcan Agents.

Vulcan Persistence requires broadcast communication with Vulcan Agents in order to manage the Agent pool and control checks in execution. Because Agents might not be reachable from the internet, the Stream provides a websocket stream that Agents connect to in order to receive input from the Persistence.

### Communication Flow

![Alt text](/_doc/img/VulcanStreamCommunicationFlow.png)

### Requirements

Vulcan Stream works on top of two main services:
- Postgres [NOTIFY](https://www.postgresql.org/docs/9.6/static/sql-notify.html)/[LISTEN](https://www.postgresql.org/docs/9.6/static/sql-listen.html)
- [Go Websocket Events](https://github.com/danfaizer/gowse)

Vulcan Stream leverages Postgres NOTIFY/LISTEN feature to publish real-time messages to a websocket stream.

#### Why Postgres NOTIFY/LISTEN

We thought about using AWS SNS and/or SQS to get notifications from Vulcan Persistence and publishing messages to the websocket stream. However, as per the nature of AWS SNS/SQS, simplicity, and because we already had Postgres as a requirement in Vulcan infrastructure, we saw the NOTIFY/LISTEN feature as the best approach to real-time broadcast notifications with proper levels of resilience and scalability.

### Build & Run

Two binaries are provided:
- vulcan-stream
- vulcan-stream-test-client

Assuming you have Docker in your machine and there are no services listening on ports `5432` or `8080`.

Run vulcan-stream:

```
go get -x github.com/adevinta/vulcan-stream/cmd/vulcan-stream

docker run -d -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_DB=stream postgres

vulcan-stream ${GOPATH}/src/github.com/adevinta/vulcan-stream/_resources/config/local.toml
```

Run vulcan-stream websocket client integration test:

```
go get -x github.com/adevinta/vulcan-stream/cmd/vulcan-stream-test-client

vulcan-stream-test-client ${GOPATH}/src/github.com/adevinta/vulcan-stream.git/_resources/config/local.toml
```

Or, connect to the stream and notify some messages:

```
curl --include --no-buffer --header "Connection: Upgrade" --header "Upgrade: websocket" \
        --header "Host: localhost:8080" --header "Origin: http://localhost:8080" \
        --header "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" --header "Sec-WebSocket-Version: 13" \
        "http://localhost:8080/stream" &

psql -c "NOTIFY events, '{\"action\":\"test\", \"check_id\":\"00000000-0000-0000-0000-000000000000\"}';" \
-U postgres -h localhost stream
```

### Configure

You can see and modify Vulcan Stream configuration as required:

`_resources/config/local.toml`


# Docker execute

Those are the variables you have to setup:

|Variable|Description|Sample|
|---|---|---|
|PORT|Listen http port|8080|
|LOG_LEVEL||DEBUG|
|PG_(HOST\|DB\|USER\|PWD\|PORT)|Postgresql variables||
|PG_SSLMODE|One of these (disable,allow,prefer,require,verify-ca,verify-full)|disable|
|PG_CA_B64|A base64 encoded ca certificate||
|STREAM||stream|

```bash
docker build . -t vs

# Use the default config.toml customized with env variables.
docker run --env-file ./local.env vs

# Use custom config.toml
docker run -v `pwd`/custom.toml:/app/config.toml vs
```
