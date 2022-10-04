SHELL = bash
.DEFAULT_GOAL = build

$(shell mkdir -p bin)
export GOBIN = $(realpath bin)
export PATH := $(PATH):$(GOBIN)

IMAGE ?= skiperator

.PHONY: tools
tools: bin/etcd bin/kube-apiserver
	go install sigs.k8s.io/controller-tools/cmd/controller-gen
	go install github.com/kudobuilder/kuttl/cmd/kubectl-kuttl

ETCD_URL = https://github.com/etcd-io/etcd/releases/download/v3.5.5/etcd-v3.5.5-linux-amd64.tar.gz
ETCD_PATH = etcd-v3.5.5-linux-amd64
ETCD_PATH_DEPTH = $(shell awk -F / '{ print NF }' <<< "$(ETCD_PATH)")
bin/etcd:
	wget --no-verbose --output-document - $(ETCD_URL) | \
	tar --gzip --extract --strip-components $(ETCD_PATH_DEPTH) --directory bin $(ETCD_PATH)/etcd

KUBE_APISERVER_URL = dl.k8s.io/v1.25.2/bin/linux/amd64/kube-apiserver
bin/kube-apiserver:
	wget --no-verbose --directory-prefix bin dl.k8s.io/v1.25.2/bin/linux/amd64/kube-apiserver
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
