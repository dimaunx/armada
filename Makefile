PROJECTNAME := armada
VERSION := $(shell git describe --tags | tr -d "v")
BUILD := $(shell git rev-parse HEAD)
USER := $(shell id -u)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
export GO111MODULE := on
export GOPROXY = https://proxy.golang.org

ifndef VERSION
override VERSION = dev
endif

# Go related variables.
GOCMD := go
GOBIN := $(shell go env GOPATH)/bin
GOBASE := $(shell pwd)
OUTPUTDIR := bin
GOLANGCILINT := $(GOBIN)/golangci-lint
PACKR := $(GOBIN)/packr2
GOIMPORTS := $(GOBIN)/goimports

# # Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-X github.com/dimaunx/armada/cmd/armada.Version=$(VERSION) -X github.com/dimaunx/armada/cmd/armada.Build=$(BUILD)"

$(GOLANGCILINT):
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v1.17.0

$(PACKR):
	curl -sL https://github.com/gobuffalo/packr/releases/download/v2.7.1/packr_2.7.1_$(OS)_amd64.tar.gz | tar xzvf - packr2
	mv $(GOBASE)/packr2 $(GOBIN)/packr2
	chmod a+x $(GOBIN)/packr2

$(GOIMPORTS):
	GO111MODULE=off $(GOCMD) get -u golang.org/x/tools/cmd/goimports

$(GINKGO):
	GO111MODULE=off $(GOCMD) get -u github.com/onsi/ginkgo

test: $(GINKGO)
	ginkgo -v -cover ./pkg/...
.PHONY: test

e2e: $(GINKGO)
	ginkgo -v ./test/e2e/...
.PHONY: e2e

validate: $(GOLANGCILINT) $(GOIMPORTS)
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do goimports -w -d "$$file"; done
	golangci-lint run ./...
.PHONY: validate

build: $(PACKR) validate
	$(GOCMD) mod tidy
	packr2 -v --ignore-imports
	CGO_ENABLED=0 $(GOCMD) build $(LDFLAGS) -o $(GOBASE)/$(OUTPUTDIR)/$(PROJECTNAME)
.PHONY: build

docker-run:
	${MAKE} docker ARGS="${ARGS}" || ${MAKE} fix-perm
.PHONY: docker-run

docker:
	docker run -it --rm --name $(PROJECTNAME)-$(VERSION)-runner -v /var/run/docker.sock:/var/run/docker.sock -v $(GOBASE):/$(PROJECTNAME) -w /$(PROJECTNAME) quay.io/submariner/dapper-base:latest ${ARGS}
	sudo chown -R $(USER):$(USER) $(GOBASE)
.PHONY: docker

clean: fix-perm
	rm -rf packrd debug packr2 $(OUTPUTDIR) $(GOBASE)/cmd/armada/armada-packr.go $(GOBASE)/pkg/*/*.cover* $(GOBASE)/pkg/*/output
	-docker ps -qf status=exited | xargs docker rm -f
	-docker ps -qaf name=$(PROJECTNAME)- | xargs docker rm -f
	-docker images -qf dangling=true | xargs docker rmi -f
	-docker volume ls -qf dangling=true | xargs docker volume rm -f
	-docker rmi $(PROJECTNAME):$(VERSION)
.PHONY: clean

fix-perm:
	sudo chown -R $(USER):$(USER) $(GOBASE)
.PHONY: fix-perm
