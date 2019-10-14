FROM golang:alpine as builder
MAINTAINER Jack Murdock <jack_murdock@comcast.com>

WORKDIR /go/src/github.com/xmidt-org/heimdall
ARG VERSION=undefined
ARG GITCOMMIT=undefined
ARG BUILDTIME=undefined

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

COPY . .
RUN GO111MODULE=on go mod vendor
RUN GO111MODULE=on go build -ldflags "-X 'main.BuildTime=${BUILDTIME}' -X main.GitCommit=${GITCOMMIT} -X main.Version=${VERSION}" -o heimdall_linux_amd64

FROM alpine

RUN apk --no-cache add ca-certificates
RUN mkdir -p /etc/heimdall
VOLUME /etc/heimdall

EXPOSE 8080

COPY --from=builder /go/src/github.com/xmidt-org/heimdall/heimdall_linux_amd64 /
COPY heimdall.yaml /
ENTRYPOINT ["/heimdall_linux_amd64"]
