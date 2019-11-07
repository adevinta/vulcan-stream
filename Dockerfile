FROM golang:1.13-alpine3.10 as builder

WORKDIR /app

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o vulcan-stream -a -tags netgo -ldflags '-w' cmd/vulcan-stream/main.go
# RUN go build -o vulcan-stream-test-client -a -tags netgo -ldflags '-w' cmd/vulcan-stream-test-client/main.go

# final stage
FROM alpine:3.10
WORKDIR /app
COPY --from=builder /app/vulcan-stream /app/
COPY run.sh .
EXPOSE 8080
CMD ["./run.sh"]
