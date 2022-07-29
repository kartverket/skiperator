SHELL = bash
.DEFAULT_GOAL = build

$(shell mkdir -p bin)
export GOBIN = $(realpath bin)
export PATH := $(PATH):$(GOBIN)

IMAGE ?= skiperator

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
	TF_VAR_image=$(IMAGE) \
	terraform -chdir=deployment apply -auto-approve

vault_context = $(shell kubectl config view --output jsonpath='{.current-context}')
vault_cluster = $(shell kubectl config view --output jsonpath='{.contexts[?(@.name == "$(vault_context)")].context.cluster}')
vault_cluster_address = $(shell kubectl config view --output jsonpath='{.clusters[?(@.name == "$(vault_cluster)")].cluster.server}')
vault_cluster_certificate = $(shell kubectl config view --raw --output jsonpath='{.clusters[?(@.name == "$(vault_cluster)")].cluster.certificate-authority-data}')
.PHONY: vault
vault:
	docker start vault &> /dev/null || \
	docker run \
	--name=vault \
	--cap-add=IPC_LOCK \
	--network host \
	--env VAULT_DEV_ROOT_TOKEN_ID=vault \
	--detach \
	vault > /dev/null

	@until docker exec --env VAULT_ADDR='http://127.0.0.1:8200' vault vault status; \
	do sleep 1; \
	done &> /dev/null

	kubectl create serviceaccount vault \
	--dry-run=client --output yaml | kubectl apply --filename -
	kubectl create clusterrolebinding vault-auth-delegator \
	--clusterrole=system:auth-delegator \
	--serviceaccount=default:vault \
	--dry-run=client --output yaml | kubectl apply --filename -
	kubectl create secret generic vault-token \
	--dry-run=client --output yaml | kubectl apply --filename -

	docker exec --env VAULT_ADDR='http://127.0.0.1:8200' --env VAULT_TOKEN='vault' vault \
	vault auth enable kubernetes || true
	docker exec --env VAULT_ADDR='http://127.0.0.1:8200' --env VAULT_TOKEN='vault' vault \
	vault write auth/kubernetes/config \
	token_reviewer_jwt="$$(kubectl create token vault --bound-object-kind Secret --bound-object-name vault-token)" \
	kubernetes_host='$(vault_cluster_address)' \
	kubernetes_ca_cert="$$(base64 --decode <<< '$(vault_cluster_certificate)')"