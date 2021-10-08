CONTAINER_COMMAND = $(shell if [ -x "$(shell which podman)" ];then echo "podman" ; else echo "docker";fi)
IMAGE := $(or ${IMAGE},quay.io/yrobla/openshift-ai-image-backup:latest)
GIT_REVISION := $(shell git rev-parse HEAD)
CONTAINER_BUILD_PARAMS = --label git_revision=${GIT_REVISION}

all: build build-image push-image

build:
	CGO_ENABLED=0 go build -o build/openshift-ai-image-backup src/main.go

build-image:
	$(CONTAINER_COMMAND) build $(CONTAINER_BUILD_PARAMS) -f Dockerfile . -t $(IMAGE)

push-image:
	$(CONTAINER_COMMAND) push ${IMAGE}

clean:
	rm -rf build
