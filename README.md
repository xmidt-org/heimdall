# heimdall
(pronounced “hem-dahl”)

[![Build Status](https://github.com/xmidt-org/heimdall/workflows/CI/badge.svg)](https://github.com/xmidt-org/heimdall/actions)
[![codecov.io](http://codecov.io/github/xmidt-org/heimdall/coverage.svg?branch=main)](http://codecov.io/github/xmidt-org/heimdall?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/heimdall)](https://goreportcard.com/report/github.com/xmidt-org/heimdall)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/heimdall/blob/main/LICENSE)
[![GitHub release](https://img.shields.io/github/release/xmidt-org/heimdall.svg)](CHANGELOG.md)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=xmidt-org_heimdall&metric=alert_status)](https://sonarcloud.io/dashboard?id=xmidt-org_heimdall)


## Summary

Heimdall provides metrics to determine how accurate Codex is at determining if 
a device is connected to XMiDT.

For more information on Codex, check out [the Codex README](https://github.com/xmidt-org/codex-deploy).
For more information on XMiDT, check out [the XMiDT README](https://github.com/xmidt-org/xmidt).

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Details](#details)
- [Build](#build)
- [Deploy](#deploy)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/). 
By participating, you agree to this Code.

## Details


## Build

### Source

In order to build from the source, you need a working Go environment with 
version 1.11 or greater. Find more information on the [Go website](https://golang.org/doc/install).

You can directly use `go get` to put the Heimdall binary into your `GOPATH`:
```bash
GO111MODULE=on go get github.com/xmidt-org/heimdall
```

You can also clone the repository yourself and build using make:

```bash
mkdir -p $GOPATH/src/github.com/xmidt-org
cd $GOPATH/src/github.com/xmidt-org
git clone git@github.com:xmidt-org/heimdall.git
cd heimdall
make build
```

### Makefile

The Makefile has the following options you may find helpful:
* `make build`: builds the Heimdall binary
* `make docker`: builds a docker image for Heimdall, making sure to get all 
   dependencies
* `make local-docker`: builds a docker image for Heimdall with the assumption
   that the dependencies can be found already
* `make it`: runs `make docker`, then deploys Heimdall and a cockroachdb 
   database into docker.
* `make test`: runs unit tests with coverage for Heimdall
* `make clean`: deletes previously-built binaries and object files

### RPM

First have a local clone of the source and go into the root directory of the 
repository.  Then use rpkg to build the rpm:
```bash
rpkg srpm --spec <repo location>/<spec file location in repo>
rpkg -C <repo location>/.config/rpkg.conf sources --outdir <repo location>'
```

### Docker

The docker image can be built either with the Makefile or by running a docker 
command.  Either option requires first getting the source code.

See [Makefile](#Makefile) on specifics of how to build the image that way.

For running a command, either you can run `docker build` after getting all 
dependencies, or make the command fetch the dependencies.  If you don't want to 
get the dependencies, run the following command:
```bash
docker build -t heimdall:local -f deploy/Dockerfile .
```
If you want to get the dependencies then build, run the following commands:
```bash
GO111MODULE=on go mod vendor
docker build -t heimdall:local -f deploy/Dockerfile.local .
```

For either command, if you want the tag to be a version instead of `local`, 
then replace `local` in the `docker build` command.

### Kubernetes

WIP. TODO: add info

## Deploy

For deploying on Docker or in Kubernetes, refer to the [deploy README](https://github.com/xmidt-org/codex-deploy/tree/main/deploy/README.md).

For running locally, ensure you have the binary [built](#Source).  If it's in 
your `GOPATH`, run:
```
heimdall
```
If the binary is in your current folder, run:
```
./heimdall
```

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
