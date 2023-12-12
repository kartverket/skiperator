SHELL = bash
.DEFAULT_GOAL = build

$(shell mkdir -p bin)
export GOBIN = $(realpath bin)
export PATH := $(PATH):$(GOBIN)
export OS   := $(shell if [ "$(shell uname)" = "Darwin" ]; then echo "darwin"; else echo "linux"; fi)
export ARCH := $(shell if [ "$(shell uname -m)" = "x86_64" ]; then echo "amd64"; else echo "arm64"; fi)

# Extracts the version number for a given dependency found in go.mod.
# Makes the test setup be in sync with what the operator itself uses.
extract-version = $(shell cat go.mod | grep $(1) | awk '{$$1=$$1};1' | cut -d' ' -f2 | sed 's/^v//')

#### TOOLS ####
TOOLS_DIR                          := $(PWD)/.tools
KIND                               := $(TOOLS_DIR)/kind
KIND_VERSION                       := v0.20.0
CHAINSAW_VERSION                   := $(call extract-version,github.com/kyverno/chainsaw)
CONTROLLER_GEN_VERSION             := $(call extract-version,sigs.k8s.io/controller-tools)
CERT_MANAGER_VERSION               := $(call extract-version,github.com/cert-manager/cert-manager)
ISTIO_VERSION                      := $(call extract-version,istio.io/api)
PROMETHEUS_VERSION                 := $(call extract-version,github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring)

#### VARS ####
SKIPERATOR_CONTEXT         ?= kind-$(KIND_CLUSTER_NAME)
KUBERNETES_VERSION          = 1.28.0
KIND_IMAGE                 ?= kindest/node:v$(KUBERNETES_VERSION)
KIND_CLUSTER_NAME          ?= skiperator

.PHONY: generate
generate:
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v${CONTROLLER_GEN_VERSION}
	go generate ./...

.PHONY: build
build: generate
	go build \
	-tags osusergo,netgo \
	-trimpath \
	-ldflags="-s -w" \
	-o ./bin/skiperator \
	./cmd/skiperator

.PHONY: run-local
run-local: build
	kubectl --context ${SKIPERATOR_CONTEXT} apply -f config/ --recursive
	./bin/skiperator

.PHONY: setup-local
setup-local: kind-cluster install-istio install-cert-manager install-prometheus-crds install-skiperator
	@echo "Cluster $(SKIPERATOR_CONTEXT) is setup"


#### KIND ####

.PHONY: kind-cluster check-kind
check-kind:
	@which kind >/dev/null || (echo "kind not installed, please install it to proceed"; exit 1)

.PHONY: kind-cluster
kind-cluster: check-kind
	@echo Create kind cluster... >&2
	@kind create cluster --image $(KIND_IMAGE) --name ${KIND_CLUSTER_NAME}


#### SKIPERATOR DEPENDENCIES ####

.PHONY: install-istio
install-istio:
	@echo "Creating istio-gateways namespace..."
	@kubectl create namespace istio-gateways --context $(SKIPERATOR_CONTEXT) || true
	@echo "Downloading Istio..."
	@curl -L https://istio.io/downloadIstio | ISTIO_VERSION=$(ISTIO_VERSION) TARGET_ARCH=$(ARCH) sh -
	@echo "Installing Istio on Kubernetes cluster..."
	@./istio-$(ISTIO_VERSION)/bin/istioctl install -y --context $(SKIPERATOR_CONTEXT)
	@echo "Istio installation complete."

.PHONY: install-cert-manager
install-cert-manager:
	@echo "Installing cert-manager"
	@kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v$(CERT_MANAGER_VERSION)/cert-manager.yaml --context $(SKIPERATOR_CONTEXT)

.PHONY: install-prometheus-crds
install-prometheus-crds:
	@echo "Installing prometheus crds"
	@kubectl apply -f https://github.com/prometheus-operator/prometheus-operator/releases/download/v$(PROMETHEUS_VERSION)/stripped-down-crds.yaml --context $(SKIPERATOR_CONTEXT)

.PHONY: install-skiperator
install-skiperator: generate
	@kubectl create namespace skiperator-system --context $(SKIPERATOR_CONTEXT) || true
	@kubectl apply -f config/ --recursive --context $(SKIPERATOR_CONTEXT)
	@kubectl apply -f samples/ --recursive --context $(SKIPERATOR_CONTEXT) || true

.PHONY: install-test-tools
install-test-tools:
	go install github.com/kyverno/chainsaw@v${CHAINSAW_VERSION}

#### TESTS ####
.PHONY: test-single
test-single: install-test-tools
	@./bin/chainsaw test --kube-context $(SKIPERATOR_CONTEXT) --config tests/config.yaml --test-dir $(dir) && \
    echo "Test succeeded" || (echo "Test failed" && exit 1)

.PHONY: test
test: install-test-tools
	@./bin/chainsaw test --kube-context $(SKIPERATOR_CONTEXT) --config tests/config.yaml --test-dir tests/ && \
    echo "Test succeeded" || (echo "Test failed" && exit 1)

.PHONY: run-test
run-test: build
	@echo "Starting skiperator in background..."
	@LOG_FILE=$$(mktemp -t skiperator-test.XXXXXXX); \
	./bin/skiperator -e error > "$$LOG_FILE" 2>&1 & \
	PID=$$!; \
	echo "skiperator PID: $$PID"; \
	echo "Log redirected to file: $$LOG_FILE"; \
	( \
		if [ -z "$(TEST_DIR)" ]; then \
			$(MAKE) test; \
		else \
			$(MAKE) test-single dir=$(TEST_DIR); \
		fi; \
	) && \
	(echo "Stopping skiperator (PID $$PID)..." && kill $$PID)  || (echo "Test or skiperator failed. Stopping skiperator (PID $$PID)" && kill $$PID && exit 1)
