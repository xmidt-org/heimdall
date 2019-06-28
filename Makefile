DEFAULT: build

GO           ?= go
GOFMT        ?= $(GO)fmt
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
HEIMDALL    := $(FIRST_GOPATH)/bin/heimdall

PROGVER = $(shell grep 'applicationVersion.*= ' main.go | awk '{print $$3}' | sed -e 's/\"//g')

.PHONY: go-mod-vendor
go-mod-vendor:
	GO111MODULE=on $(GO) mod vendor

.PHONY: build
build: go-mod-vendor
	$(GO) build -o heimdall

rpm:
	mkdir -p ./OPATH/SOURCES
	tar -czvf ./OPATH/SOURCES/heimdall-$(PROGVER).tar.gz . --exclude ./.git --exclude ./OPATH --exclude ./conf --exclude ./deploy --exclude ./vendor
	cp conf/heimdall.service ./OPATH/SOURCES/
	cp conf/heimdall.yaml  ./OPATH/SOURCES/
	rpmbuild --define "_topdir $(CURDIR)/OPATH" \
    		--define "_version $(PROGVER)" \
    		--define "_release 1" \
    		-ba deploy/packaging/heimdall.spec

.PHONY: version
version:
	@echo $(PROGVER)


# If the first argument is "update-version"...
ifeq (update-version,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "update-version"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS):;@:)
endif

.PHONY: update-version
update-version:
	@echo "Update Version $(PROGVER) to $(RUN_ARGS)"
	sed -i "s/$(PROGVER)/$(RUN_ARGS)/g" main.go


.PHONY: install
install: go-mod-vendor
	echo go build -o $(HEIMDALL) $(PROGVER)

.PHONY: release-artifacts
release-artifacts: go-mod-vendor
	GOOS=darwin GOARCH=amd64 go build -o ./OPATH/heimdall-$(PROGVER).darwin-amd64
	GOOS=linux  GOARCH=amd64 go build -o ./OPATH/heimdall-$(PROGVER).linux-amd64

.PHONY: docker
docker:
	docker build -f ./deploy/Dockerfile -t heimdall:$(PROGVER) .

# build docker without running modules
.PHONY: local-docker
local-docker:
	docker build -f ./deploy/Dockerfile.local -t heimdall:local .

.PHONY: style
style:
	! gofmt -d $$(find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

.PHONY: test
test: go-mod-vendor
	go test -o $(HEIMDALL) -v -race  -coverprofile=cover.out $(go list ./... | grep -v "/vendor/")

.PHONY: test-cover
test-cover: test
	go tool cover -html=cover.out

.PHONY: codecov
codecov: test
	curl -s https://codecov.io/bash | bash

.PHONEY: it
it:
	./it.sh

.PHONY: clean
clean:
	rm -rf ./codex-heimdall ./OPATH ./coverage.txt ./vendor