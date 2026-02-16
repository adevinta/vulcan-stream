# Copyright 2019 Adevinta

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

WORKDIR /app

ARG TARGETOS TARGETARCH

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN GO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -tags netgo -ldflags '-w' ./cmd/vulcan-stream

# final stage
FROM alpine:3.23
RUN apk add --no-cache --update gettext

WORKDIR /app
COPY --from=builder /app/vulcan-stream /app/
COPY config.toml .
COPY run.sh .
CMD ["./run.sh"]
