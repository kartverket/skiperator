# AGENTS.md

This file is for coding agents working in `github.com/kartverket/skiperator`.
Optimize for maintainers using a GitHub Copilot-style IDE agent with read/edit/search/web capabilities. Do not assume shell execution, cluster access, or permission to run commands.

## Purpose

- Keep changes targeted and repo-specific.
- Prefer reading the existing code and matching local patterns over inventing new abstractions.
- Suggest commands when useful, but treat cluster-heavy workflows as human-invoked unless the task explicitly says otherwise.

## Where To Look First

- `api/v1alpha1/`: CRD and API types for `Application`, `SKIPJob`, `Routing`, and related shared types.
- `internal/controllers/`: reconciliation entrypoints for each resource.
- `pkg/reconciliation/`: orchestration of reconciliation behavior.
- `pkg/resourcegenerator/`: generated Kubernetes resource construction.
- `pkg/resourceprocessor/`: diffing and apply logic for generated resources.
- `tests/`: Chainsaw integration suites grouped by feature and resource.
- `config/`: generated CRDs, RBAC, and manifests used for local cluster setup.

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
- `api/v1alpha1/*/zz_generated.deepcopy.go`
- `config/crd/*.yaml`
- `config/rbac/role.yaml`

`go generate ./...` is the source of truth for these updates, and CI checks that generated files are committed.

## Progressive Disclosure

- Use `README.md` for CR examples and high-level product context.
- Use `CONTRIBUTING.md` for local setup, testing, and development workflow details.
- Use `doc/access-policies.md` when working on access policy behavior.
- Use `api-docs.md` when you need the rendered API surface rather than the Go types directly.
