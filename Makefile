SHELL = bash
.DEFAULT_GOAL = build

$(shell mkdir -p bin)
export GOBIN = $(realpath bin)
export PATH := $(PATH):$(GOBIN)
export OS   := $(shell if [ "$(shell uname)" = "Darwin" ]; then echo "darwin"; else echo "linux"; fi)
export ARCH := $(shell if [ "$(shell uname -m)" = "x86_64" ]; then echo "amd64"; else echo "arm64"; fi)

SKIPERATOR_CONTEXT ?= kind-kind
KUBERNETES_VERSION = 1.25

.PHONY: test-tools
test-tools:
	wget --no-verbose --output-document - "https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-${KUBERNETES_VERSION}.0-${OS}-${ARCH}.tar.gz" | \
    tar --gzip --extract --strip-components 2 --directory bin
	go install github.com/kudobuilder/kuttl/cmd/kubectl-kuttl@v0.15.0


.PHONY: generate
generate:
	go install sigs.k8s.io/controller-tools/cmd/controller-gen
	go generate ./...

.PHONY: build
build: generate
	go build \
	-tags osusergo,netgo \
	-trimpath \
	-ldflags="-s -w" \
	-o ./bin/skiperator \
	./cmd/skiperator

.PHONY: test
test: test-tools
	TEST_ASSET_ETCD=bin/etcd \
	TEST_ASSET_KUBE_APISERVER=bin/kube-apiserver \
	DEBUG_LEVEL=warn \
	kubectl kuttl test \
	--config tests/config.yaml \
	--start-control-plane \
	--suppress-log=events

.PHONY: build-test
build-test: build test

.PHONY: run-local
run-local: build
	kubectl --context ${SKIPERATOR_CONTEXT} apply -f config/ --recursive
	./bin/skiperator
