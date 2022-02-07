SHELL :=/bin/bash
export GO111MODULE=on
unexport GOPATH

CONTAINER_COMMAND = $(shell if [ -x "$(shell which podman)" ];then echo "podman" ; else echo "docker";fi)
IMAGE := $(or ${IMAGE},quay.io/redhat_ztp/openshift-ai-image-backup:latest)
GIT_REVISION := $(shell git rev-parse HEAD)
CONTAINER_BUILD_PARAMS = --label git_revision=${GIT_REVISION}

all: build build-image
.PHONY: all

# Include the library makefile
include $(addprefix ./vendor/github.com/openshift/build-machinery-go/make/, \
	golang.mk \
	targets/openshift/deps-gomod.mk \
	targets/openshift/bindata.mk \
	targets/openshift/images.mk \
)

# This will call a macro called "add-bindata" which will generate bindata specific targets based on the parameters:
# $0 - macro name
# $1 - target suffix
# $2 - input dirs
# $3 - prefix
# $4 - pkg
# $5 - output
# It will generate targets {update,verify}-bindata-$(1) logically grouping them in unsuffixed versions of these targets
# and also hooked into {update,verify}-generated for broader integration.
$(call add-bindata,recovery,./bindata/recovery/...,bindata,recovery_assets,internal/recovery_assets/bindata.go)

build:
	hack/build-go.sh
.PHONY: build

build-image: build
	$(CONTAINER_COMMAND) build $(CONTAINER_BUILD_PARAMS) -f Dockerfile . -t $(IMAGE)
.PHONY:

push-image: build-image
	$(CONTAINER_COMMAND) push ${IMAGE}

.PHONY: build

check: | verify golangci-lint shell-check
.PHONY: check

golangci-lint:
		golangci-lint run --verbose --print-resources-usage --modules-download-mode=vendor --timeout=5m0s
.PHONY: golangci-lint

ifeq ($(shell which shellcheck 2>/dev/null),)
shell-check:
	@echo "Skipping shellcheck: Not installed"
else
shell-check:
	find . -name '*.sh' -not -path './vendor/*' -not -path './git/*' -print0 | xargs -0 --no-run-if-empty shellcheck
endif
.PHONY: shell-check

GO_TEST_PACKAGES :=./pkg/... ./cmd/...
