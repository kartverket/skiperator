# AGENTS.md

This file is for coding agents working in `github.com/kartverket/skiperator`.
Optimize for maintainers using a GitHub Copilot-style IDE agent with read/edit/search/web capabilities. Do not assume shell execution, cluster access, or permission to run commands.

## Purpose

- Keep changes targeted and repo-specific.
- Prefer reading the existing code and matching local patterns over inventing new abstractions.
- Suggest commands when useful, but treat cluster-heavy workflows as human-invoked unless the task explicitly says otherwise.

## Where To Look First

- `api/v1alpha1/`: CRD and API types for `Application`, `SKIPJob`, `Routing`, and related shared types.
- `api/v1beta1/`: `SKIPJob` promoted to v1beta1 — prefer this version for new SKIPJob features.
- `api/common/`: Shared interfaces (`SKIPObject`) and embedded types (`PodSettings`, `ContainerSettings`, etc.) reused across resource kinds.
- `internal/controllers/`: reconciliation entrypoints for each resource.
- `internal/controllers/common/`: shared `ReconcilerBase`, conditions helpers, and status utilities embedded by all controllers.
- `internal/webhook/`: defaulting and validation webhooks, one file per resource type.
- `internal/config/`: `Config` struct loaded from environment variables / flags at startup.
- `pkg/reconciliation/`: orchestration of reconciliation behavior.
- `pkg/resourcegenerator/`: generated Kubernetes resource construction; one subdirectory per generated resource kind (e.g. `deployment/`, `service/`, `istio/`, `networkpolicy/`, `hpa/`, `pdb/`, `serviceaccount/`, `ingress/`, `configmap/`, `batch/`).
- `pkg/resourceprocessor/`: diffing and apply logic for generated resources.
- `pkg/auth/`: GCP Workload Identity / service account auth helpers.
- `pkg/k8sfeatures/`: feature detection for cluster capabilities.
- `pkg/testutil/`: helpers for controller and integration tests.
- `pkg/util/`: general-purpose helpers (maps, slices, pointers, etc.).
- `tests/`: Chainsaw integration suites grouped by feature and resource.
- `config/`: generated CRDs, RBAC, and manifests used for local cluster setup.

## API Versions

- `v1alpha1`: `Application`, `Routing`, `SKIPJob` — original stable versions, still in active use.
- `v1beta1`: `SKIPJob` — promoted version. New SKIPJob features go here. When editing SKIPJob logic, check both versions for parity; do not add fields to `v1alpha1` that are not reflected in `v1beta1`.
- Generated deep-copy files (`zz_generated.deepcopy.go`) exist in both `api/v1alpha1/` and `api/v1beta1/` — do not hand-edit them.

## Controller Pattern

Each controller (`internal/controllers/*.go`) follows the same five-step pattern:

1. Fetch the resource from the API server.
2. Build a `reconciliation.Reconciliation` context struct (carries the object, config, client, scheme, logger, etc.).
3. Call resource generator functions (`pkg/resourcegenerator/<kind>/Generate(...)`) to produce the desired child resources. Generator functions must be pure — no API calls, no side effects.
4. Pass results to `pkg/resourceprocessor`, which diffs and applies/prunes child resources against the cluster.
5. Update status conditions via helpers in `internal/controllers/common/` (use typed conditions, not raw status strings).

`internal/controllers/common/reconciler.go` provides `ReconcilerBase`, which is embedded by all four controllers.

## Webhooks

`internal/webhook/` contains one file per resource type implementing `Default()` (defaulting) and/or `ValidateCreate/Update/Delete()` (validation). Webhooks are registered in `cmd/skiperator/main.go`. When adding a new field: update the type, update the webhook, and keep `Default()` idempotent.

## Configuration

`internal/config/` defines the `Config` struct populated from environment variables and command-line flags at operator startup. Pass config through `reconciliation.Reconciliation`; do not read environment variables anywhere else in the codebase.

## Canonical Workflows

Use these as suggestions, not assumptions:

- `go generate ./...`
  Run after API changes or kubebuilder marker changes. This updates generated files.
- `make build`
  Builds the operator binary.
- `make run-unit-tests`
  Cheapest verification path for Go changes.
- `make test-single dir=tests/...`
  Runs one Chainsaw suite when a local cluster is already available.
- `make test`
  Runs all Chainsaw suites when a local cluster is already available.
- `make setup-local`, `make run-local`, `make run-test`
  Human-invoked local cluster workflows. Do not assume these should be run by default.

## Boundaries And Defaults

- Prefer code reading, small edits, and narrow verification over broad refactors.
- Treat cluster setup, dependency installs, `kubectl apply`, and long Chainsaw runs as ask-first or human-run workflows.
- When suggesting `kubectl` or Chainsaw commands, use `kind-skiperator` explicitly or set `SKIPERATOR_CONTEXT=kind-skiperator`.
- Do not hand-edit generated files unless the source change requires regeneration.
- Do not replace existing docs with restated copies. Point to them when deeper context is needed.

## Generated Files

Changes to API types or kubebuilder markers usually require regenerated outputs:

- `api/v1alpha1/zz_generated.deepcopy.go`
- `api/v1beta1/zz_generated.deepcopy.go`
- `api/v1alpha1/*/zz_generated.deepcopy.go`
- `config/crd/*.yaml`
- `config/rbac/role.yaml`

`go generate ./...` is the source of truth for these updates, and CI checks that generated files are committed.

## Progressive Disclosure

- Use `README.md` for CR examples and high-level product context.
- Use `CONTRIBUTING.md` for local setup, testing, and development workflow details.
- Use `doc/access-policies.md` when working on access policy behavior.
- Use `api-docs.md` when you need the rendered API surface rather than the Go types directly.
