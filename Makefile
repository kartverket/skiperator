.DEFAULT_GOAL = build

$(shell mkdir -p bin)
export GOBIN = $(realpath bin)
export PATH := $(PATH):$(GOBIN)

IMAGE ?= skiperator
KUBECONFIG ?= ~/.kube/config

.PHONY: tools
tools:
	go install sigs.k8s.io/controller-tools/cmd/controller-gen

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

.PHONY: image
image:
	docker build --tag $(IMAGE) .

.PHONY: push
push: image
	docker push $(IMAGE)

.PHONY: deploy
deploy: generate push
	KUBE_CONFIG_PATH=$(KUBECONFIG) \
	TF_VAR_image=$(IMAGE) \
	terraform -chdir=deployment apply -auto-approve