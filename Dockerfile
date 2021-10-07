FROM registry.ci.openshift.org/openshift/release:golang-1.16 AS builder
ENV GOFLAGS=-mod=mod
WORKDIR /go/src/github.com/redhat-ztp/openshift-ai-image-backup

# Bring in the go dependencies before anything else so we can take
# advantage of caching these layers in future builds.
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .
RUN make build

FROM quay.io/centos/centos:stream8

COPY --from=builder /go/src/github.com/redhat-ztp/openshift-ai-image-backup/build/openshift-ai-image-backup /usr/bin/openshift-ai-image-backup

ENTRYPOINT ["/usr/bin/openshift-ai-image-backup"]