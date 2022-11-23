SHELL = bash
.DEFAULT_GOAL = build

$(shell mkdir -p bin)
export GOBIN = $(realpath bin)
export PATH := $(PATH):$(GOBIN)
export ARCH := $(shell if [ "$(shell uname -m)" = "x86_64" ]; then echo "amd64"; else echo "arm64"; fi)

IMAGE ?= skiperator

ETCD_VER = v3.5.6
ETCD_PATH = "etcd-${ETCD_VER}-linux-${ARCH}"
ETCD_PATH_DEPTH = $(shell awk -F / '{ print NF }' <<< "$(ETCD_PATH)")

KUBERNETES_VERSION = v1.25.4

.PHONY: tools
tools: bin/kubectl bin/etcd bin/kube-apiserver
	go install sigs.k8s.io/controller-tools/cmd/controller-gen
	go install github.com/kudobuilder/kuttl/cmd/kubectl-kuttl

bin/kubectl:
	wget --no-verbose --directory-prefix bin "https://dl.k8s.io/release/${KUBERNETES_VERSION}/bin/linux/${ARCH}/kubectl"
	chmod +x bin/kubectl

bin/etcd:
	wget --no-verbose --output-document - "https://github.com/etcd-io/etcd/releases/download/${ETCD_VER}/etcd-${ETCD_VER}-linux-${ARCH}.tar.gz" | \
	tar --gzip --extract --strip-components ${ETCD_PATH_DEPTH} --directory bin ${ETCD_PATH}/etcd

bin/kube-apiserver:
	wget --no-verbose --directory-prefix bin "https://dl.k8s.io/release/${KUBERNETES_VERSION}/bin/linux/${ARCH}/kube-apiserver"
	chmod +x bin/kube-apiserver

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
test: build
	TEST_ASSET_ETCD=bin/etcd \
	TEST_ASSET_KUBE_APISERVER=bin/kube-apiserver \
	kubectl kuttl test \
	--config tests/config.yaml \
	--control-plane-config tests/apiserver.conf

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
