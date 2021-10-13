export GO111MODULE=on
unexport GOPATH

CONTAINER_COMMAND = $(shell if [ -x "$(shell which podman)" ];then echo "podman" ; else echo "docker";fi)
IMAGE := $(or ${IMAGE},quay.io/yrobla/openshift-ai-image-backup:latest)
GIT_REVISION := $(shell git rev-parse HEAD)
CONTAINER_BUILD_PARAMS = --label git_revision=${GIT_REVISION}

all: build build-image push-image
.PHONY: all

# Include the library makefile
include $(addprefix ./vendor/github.com/openshift/build-machinery-go/make/, \
	golang.mk \
	targets/openshift/deps-gomod.mk \
    targets/openshift/images.mk \
)

build:
	CGO_ENABLED=0 go build -o build/openshift-ai-image-backup pkg/main.go
.PHONY: build

build-image:
	$(CONTAINER_COMMAND) build $(CONTAINER_BUILD_PARAMS) -f Dockerfile . -t $(IMAGE)
.PHONY:

push-image:
	$(CONTAINER_COMMAND) push ${IMAGE}

clean:
	rm -rf build
.PHONY: build

GO_TEST_PACKAGES :=./pkg/... ./cmd/...
