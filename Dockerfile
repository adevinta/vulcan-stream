FROM golang:1.9
RUN mkdir -p $GOPATH/src/github.com/adevinta/vulcan-stream
WORKDIR $GOPATH/src/github.com/adevinta/vulcan-stream
ADD . $GOPATH/src/github.com/adevinta/vulcan-stream
WORKDIR $GOPATH/src/github.com/adevinta/vulcan-stream/cmd/vulcan-stream
