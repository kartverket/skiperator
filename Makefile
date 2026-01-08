SHELL = bash
.DEFAULT_GOAL = build

$(shell mkdir -p bin)
export OS   := $(shell if [ "$(shell uname)" = "Darwin" ]; then echo "darwin"; else echo "linux"; fi)
export ARCH := $(shell if [ "$(shell uname -m)" = "x86_64" ]; then echo "amd64"; else echo "arm64"; fi)

# Extracts the version number for a given dependency found in go.mod.
# Makes the test setup be in sync with what the operator itself uses.
extract-version = $(shell cat go.mod | grep $(1) | awk '{$$1=$$1};1' | cut -d' ' -f2 | sed 's/^v//')

#### TOOLS ####
TOOLS_DIR                          := $(PWD)/.tools
KIND                               := $(TOOLS_DIR)/kind
KIND_VERSION                       := v0.31.0
CERT_MANAGER_VERSION               := $(call extract-version,github.com/cert-manager/cert-manager)
ISTIO_VERSION                      := $(call extract-version,istio.io/client-go)
PROMETHEUS_VERSION                 := $(call extract-version,github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring)

#### VARS ####
SKIPERATOR_CONTEXT         ?= kind-$(KIND_CLUSTER_NAME)
KUBERNETES_VERSION          = 1.33.7
KIND_IMAGE                 ?= kindest/node:v$(KUBERNETES_VERSION)
KIND_CLUSTER_NAME          ?= skiperator
LOCAL_WEBHOOK_CERTS_DIR    := $(shell mktemp -d -t skiperator-webhook-certs.XXXXXXX)
WEBHOOK_HOST                = 0.0.0.0
.PHONY: generate
generate:
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
run-local: build install-skiperator
	@echo "Extracting webhook certificates for local development..."
	./hack/extract-webhook-certs.sh $(LOCAL_WEBHOOK_CERTS_DIR) $(SKIPERATOR_CONTEXT)
	@echo "Setting up webhook service to route to host..."
	@./hack/setup-local-webhook-endpoint.sh $(SKIPERATOR_CONTEXT)
	@echo ""
	@echo "Starting skiperator with webhook on 0.0.0.0:9443 (accessible from kind cluster)..."
	./bin/skiperator --webhook-cert-dir=$(LOCAL_WEBHOOK_CERTS_DIR) --webhook-host=$(WEBHOOK_HOST)

.PHONY: setup-local
setup-local: kind-cluster install-istio install-cert-manager install-prometheus-crds install-digdirator-crds install-skiperator
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

# Image tags for istio
ISTIO_IMAGES = docker.io/istio/proxyv2:$(ISTIO_VERSION) docker.io/istio/pilot:$(ISTIO_VERSION)

.PHONY: install-istio
install-istio:
	@echo "Creating istio-gateways namespace..."
	@kubectl create namespace istio-gateways --context $(SKIPERATOR_CONTEXT) || true

	# Manually pull and load images into the cluster for local testing

	@for image in $(ISTIO_IMAGES); do \
	  echo "Removing cached $$image if exists"; \
	  docker rmi -f "$$image" 2>/dev/null || true; \
	  echo "Pulling $$image for platform linux/$(ARCH)"; \
	  docker pull --platform linux/$(ARCH) "$$image"; \
	  echo "Loading $$image into kind cluster 'skiperator'"; \
	  docker save "$$image" | docker exec -i skiperator-control-plane ctr --namespace=k8s.io images import -; \
	done

	@echo "Downloading Istio..."
	@curl -L https://istio.io/downloadIstio | ISTIO_VERSION=$(ISTIO_VERSION) TARGET_ARCH=$(ARCH) sh -
	@echo "Installing Istio on Kubernetes cluster..."
	@./istio-$(ISTIO_VERSION)/bin/istioctl install -y --context $(SKIPERATOR_CONTEXT)
	@echo "Istio installation complete."

.PHONY: install-cert-manager
install-cert-manager:
	# Manually pull and load images into the cluster for local testing
	@echo "Pulling and loading cert-manager images"
	@curl -L -s https://github.com/cert-manager/cert-manager/releases/download/v$(CERT_MANAGER_VERSION)/cert-manager.yaml \
	| grep 'image:' \
	| sed -E 's/.*image:[[:space:]]*"?([^"]*)"?/\1/' \
	| while read -r image; do \
	    echo "Removing cached $$image if exists"; \
	    docker rmi -f "$$image" 2>/dev/null || true; \
	    echo "Pulling $$image for platform linux/$(ARCH)"; \
	    docker pull --platform linux/$(ARCH) "$$image"; \
	    echo "Loading $$image into kind cluster '$(KIND_CLUSTER_NAME)'"; \
	    docker save "$$image" | docker exec -i $(KIND_CLUSTER_NAME)-control-plane ctr --namespace=k8s.io images import -; \
	  done

	@echo "Installing cert-manager"
	@kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v$(CERT_MANAGER_VERSION)/cert-manager.yaml --context $(SKIPERATOR_CONTEXT)


.PHONY: install-prometheus-crds
install-prometheus-crds:
	@echo "Installing prometheus crds"
	@kubectl apply -f https://github.com/prometheus-operator/prometheus-operator/releases/download/v$(PROMETHEUS_VERSION)/stripped-down-crds.yaml --context $(SKIPERATOR_CONTEXT)

.PHONY: install-digdirator-crds
install-digdirator-crds:
	@echo "Installing digdirator crds"
	@kubectl apply -f https://raw.githubusercontent.com/nais/liberator/main/config/crd/bases/nais.io_idportenclients.yaml --context $(SKIPERATOR_CONTEXT)
	@kubectl apply -f https://raw.githubusercontent.com/nais/liberator/main/config/crd/bases/nais.io_maskinportenclients.yaml --context $(SKIPERATOR_CONTEXT)

.PHONY: install-skiperator
install-skiperator: generate
	@kubectl create namespace skiperator-system --context $(SKIPERATOR_CONTEXT) || true
	@kustomize build config/cert-manager | kubectl apply -f - --context $(SKIPERATOR_CONTEXT) || true
	@kustomize build config/crd | kubectl apply -f - --context $(SKIPERATOR_CONTEXT) || true
	@kustomize build config/webhook | kubectl apply -f - --context $(SKIPERATOR_CONTEXT) || true
	@kubectl apply -f config/rbac --context $(SKIPERATOR_CONTEXT) || true
	@kubectl apply -f config/static --context $(SKIPERATOR_CONTEXT) || true
	@kubectl apply -f config/skiperator-config.yaml --context $(SKIPERATOR_CONTEXT) || true
	@kubectl apply -f config/docker-config.yaml --context $(SKIPERATOR_CONTEXT) || true
	@kubectl apply -f config/github-config.yaml --context $(SKIPERATOR_CONTEXT) || true
	@kubectl apply -f tests/cluster-config/ --recursive --context $(SKIPERATOR_CONTEXT) || true


#### TESTS ####
.PHONY: test-single
test-single: install-skiperator
	@go tool chainsaw test --kube-context $(SKIPERATOR_CONTEXT) --config tests/config.yaml --test-dir $(dir) && \
    echo "Test succeeded" || (echo "Test failed" && exit 1)

.PHONY: test
test: install-skiperator
	@go tool chainsaw test --kube-context $(SKIPERATOR_CONTEXT) --config tests/config.yaml --test-dir tests/ && \
    echo "Test succeeded" || (echo "Test failed" && exit 1)

.PHONY: run-unit-tests
run-unit-tests:
	@failed_tests=$$(go test ./... 2>&1 | grep "^FAIL" | awk '{print $$2}'); \
		if [ -n "$$failed_tests" ]; then \
			echo -e "\033[31mFailed Unit Tests: [$$failed_tests]\033[0m" && exit 1; \
		else \
			echo -e "\033[32mAll unit tests passed\033[0m"; \
		fi

.PHONY: run-test
run-test: build install-skiperator
	@echo "Starting skiperator in background..."
	@LOG_FILE=$$(mktemp -t skiperator-test.XXXXXXX); \
	./bin/skiperator > "$$LOG_FILE" 2>&1 & \
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
	(echo "Stopping skiperator (PID $$PID)..." && kill $$PID && echo "running unit tests..." && $(MAKE) run-unit-tests)  || (echo "Test or skiperator failed. Stopping skiperator (PID $$PID)" && kill $$PID && exit 1)

# Checks the delta of requests made to the kube api from the controller.
.PHONY: benchmark-chainsaw-tests
benchmark-chainsaw-tests: build install-skiperator
	@echo "Starting skiperator in background..."
	@LOG_FILE=$$(mktemp -t skiperator-test.XXXXXXX); \
	METRICS_BEFORE=$$(mktemp -t metrics-before.XXXXXXX); \
	METRICS_AFTER=$$(mktemp -t metrics-after.XXXXXXX); \
	./bin/skiperator > "$$LOG_FILE" 2>&1 & \
	PID=$$!; \
	echo "Waiting for skiperator to start and sync..."; \
	sleep 10s; \
	echo "Fetching metrics before..."; \
	curl -s http://127.0.0.1:8181/metrics | grep rest_client_requests_total{ > "$$METRICS_BEFORE"; \
	echo "Run application tests"; \
	make test; \
	echo "Fetching metrics after..."; \
	curl -s http://127.0.0.1:8181/metrics | grep rest_client_requests_total{ > "$$METRICS_AFTER"; \
    kill $$PID; \
	cat $$METRICS_BEFORE; \
	echo "---"; \
	cat $$METRICS_AFTER; \
	echo "Kubernetes API usage (delta):"; \
	gawk ' \
		FNR==NR && $$0 ~ /rest_client_requests_total/ { \
			split($$0, a, " "); \
			split(a[1], b, "method=\""); \
			split(b[2], c, "\""); \
			method = c[1]; \
			val = a[2]; \
			before[method] += val; \
			next; \
		} \
		$$0 ~ /rest_client_requests_total/ { \
			split($$0, a, " "); \
			split(a[1], b, "method=\""); \
			split(b[2], c, "\""); \
			method = c[1]; \
			val = a[2]; \
			delta = val - before[method]; \
			if (delta > 0) { \
				method_delta[method] += delta; \
				total += delta; \
			} \
			before[method] = val; \
		} \
		END { \
			n = asorti(method_delta, sorted); \
			for (i=1; i<=n; i++) { \
				m = sorted[i]; \
				printf("  %s: %d\n", m, method_delta[m]); \
			} \
			printf("  Total: %d\n", total); \
		} \
	' "$$METRICS_BEFORE" "$$METRICS_AFTER"; \
	echo "Done. Logs saved to $$LOG_FILE"


.PHONY: benchmark-long-run
benchmark-long-run: build install-skiperator
		@echo "Applying anonymous metrics RBAC..."; \
    	kubectl apply -f tests/cluster-config/allow-anonymous-metrics.yaml; \
    	echo "Starting port-forward to API server..."; \
    	APISERVER_POD=$$(kubectl -n kube-system get pods -l component=kube-apiserver -o jsonpath='{.items[0].metadata.name}'); \
    	kubectl -n kube-system port-forward pod/$$APISERVER_POD 8443:6443 >/dev/null 2>&1 & \
    	PID=$$!; \
    	sleep 3; \
		echo "Starting skiperator in background..."; \
		@LOG_FILE=$$(mktemp -t skiperator-test.XXXXXXX); \
		./bin/skiperator > "$$LOG_FILE" 2>&1 & \
		SPID=$$!; \
    	echo "Waiting for skiperator to start and sync..."; \
		sleep 20; \
    	echo "Summing apiserver_request_total metrics by verb (before)..."; \
    	METRICS_BEFORE=$$(curl -sk https://localhost:8443/metrics | grep '^apiserver_request_total{' | \
    	awk ' \
    		{ \
    			if (match($$0, /verb="[^"]+"/)) { \
    				verb=substr($$0, RSTART+6, RLENGTH-7); \
    				val=$$NF; \
    				sum[verb]+=val; \
    			} \
    		} \
    		END { \
    			for (v in sum) printf "%s %d\n", v, sum[v]; \
    		}'); \
    	echo "Applying resources..."; \
		kubectl apply -f tests/application/access-policy/external-ip-policy.yaml; \
		kubectl apply -f tests/application/ingress/application.yaml; \
		kubectl apply -f tests/application/minimal/application.yaml; \
		kubectl apply -f tests/application/custom-certificate/application.yaml; \
		kubectl apply -f tests/application/copy/application.yaml; \
		kubectl apply -f tests/application/gcp/application.yaml; \
		kubectl apply -f tests/application/replicas/application.yaml; \
		kubectl apply -f tests/application/service/application.yaml; \
		kubectl apply -f tests/application/telemetry/application.yaml; \
		sleep 600s; \
    	echo "Summing apiserver_request_total metrics by verb (after)..."; \
    	METRICS_AFTER=$$(curl -sk https://localhost:8443/metrics | grep '^apiserver_request_total{' | \
    	awk ' \
    		{ \
    			if (match($$0, /verb="[^"]+"/)) { \
    				verb=substr($$0, RSTART+6, RLENGTH-7); \
    				val=$$NF; \
    				sum[verb]+=val; \
    			} \
    		} \
    		END { \
    			for (v in sum) printf "%s %d\n", v, sum[v]; \
    		}'); \
    	echo "Delta by verb:"; \
		echo "$$METRICS_AFTER" | while read va aval; do \
			bval=$$(echo "$$METRICS_BEFORE" | awk '$$1=="'$$va'" {print $$2}'); \
			[ -z "$$bval" ] && bval=0; \
			delta=$$((aval - bval)); \
			if [ "$$delta" != "0" ]; then \
				printf "%s: %d\n" "$$va" "$$delta"; \
			fi; \
		done; \
    	echo "Cleaning up port-forward, skiperator, resources..."; \
		kubectl delete -f tests/application/access-policy/external-ip-policy.yaml; \
		kubectl delete -f tests/application/ingress/application.yaml; \
		kubectl delete -f tests/application/minimal/application.yaml; \
		kubectl delete -f tests/application/custom-certificate/application.yaml; \
		kubectl delete -f tests/application/copy/application.yaml; \
		kubectl delete -f tests/application/gcp/application.yaml; \
		kubectl delete -f tests/application/replicas/application.yaml; \
		kubectl delete -f tests/application/service/application.yaml; \
		kubectl delete -f tests/application/telemetry/application.yaml; \
    	kill $$PID; \
        kill $$SPID
