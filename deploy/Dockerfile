FROM golang:alpine as builder
MAINTAINER Jack Murdock <jack_murdock@comcast.com>

WORKDIR /go/src/github.com/comcast/codex-status

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

COPY . .
RUN GO111MODULE=on go mod vendor
RUN GO111MODULE=on go build -o status_linux_amd64

FROM alpine

RUN apk --no-cache add ca-certificates
RUN mkdir -p /etc/status
VOLUME /etc/status

EXPOSE 8080
EXPOSE 7203

COPY --from=builder /go/src/github.com/comcast/codex-status/status_linux_amd64 /
COPY status.yaml /
ENTRYPOINT ["/status_linux_amd64"]
