SHELL = bash
.DEFAULT_GOAL = build

$(shell mkdir -p bin)
export GOBIN = $(realpath bin)
export PATH := $(PATH):$(GOBIN)
export OS   := $(shell if [ "$(shell uname)" = "Darwin" ]; then echo "darwin"; else echo "linux"; fi)
export ARCH := $(shell if [ "$(shell uname -m)" = "x86_64" ]; then echo "amd64"; else echo "arm64"; fi)

SKIPERATOR_CONTEXT ?= kind-kind
IMAGE ?= skiperator

KUBERNETES_VERSION = 1.25

.PHONY: tools
tools:
	go install sigs.k8s.io/controller-tools/cmd/controller-gen
	go install github.com/kudobuilder/kuttl/cmd/kubectl-kuttl

bin/kubebuilder-tools:
	wget --no-verbose --output-document - "https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-${KUBERNETES_VERSION}.0-${OS}-${ARCH}.tar.gz" | \
    tar --gzip --extract --strip-components 2 --directory bin


.PHONY: generate
generate: tools
	go generate ./...

.PHONY: build
build: generate
	go build \
	-tags osusergo,netgo \
	-trimpath \
	-ldflags="-s -w" \
	-o ./bin/skiperator \
	./cmd/skiperator

# --control-plane-config is a workaround for https://github.com/kudobuilder/kuttl/issues/378
.PHONY: test
test: bin/kubebuilder-tools build
	TEST_ASSET_ETCD=bin/etcd \
	TEST_ASSET_KUBE_APISERVER=bin/kube-apiserver \
	kubectl kuttl test \
	--config tests/config.yaml \
	--start-control-plane

.PHONY: run-local
run-local: build
	kubectl --context ${SKIPERATOR_CONTEXT} apply -f deployment/
	./bin/skiperator

.PHONY: image
image:
	docker build --tag $(IMAGE) .

.PHONY: push
push: image
	docker push $(IMAGE)

.PHONY: deploy
deploy: generate push
	TF_VAR_image=$(IMAGE) \
	terraform -chdir=deployment apply -auto-approve
