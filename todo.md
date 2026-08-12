# Cleanup todo — gwapi branch

Findings from an over-engineering review of the Gateway API routing PR. Scope is
complexity only: no behaviour changes, no bug fixes. Each item names the file,
what to cut, and what replaces it.

## Done

- [x] `api/v1alpha1/application_types.go` — `hashHostname` was a copy of
      `util.GenerateHashFromName`. Deleted, call util.

## Dead code

- [x] `internal/controllers/common/reconciler.go` — `SetSyncedState` has no
      callers left; `skipjob.go` inlined it. Delete.
- [x] `internal/controllers/common/reconciler.go` — private `updateStatus` is a
      pass-through to `UpdateStatus`. Delete, call `UpdateStatus` at the 3 sites.
- [x] `internal/controllers/common/util.go` — `SetReadyCondition`,
      `SetReadyInvalidConfig`, `SetReadyReconciled` return `[]metav1.Condition`
      that no caller reads. Drop the return type and the `FindStatusCondition`
      tail.
- [x] `pkg/k8sfeatures/crd.go` — `crd == nil` after `err == nil` is
      unreachable. Delete.
- [ ] `pkg/resourcegenerator/gatewayapi/routing.go` — the warn closure passed to
      `backendRule` can never fire: Routing passes `retries: nil` and
      `retryPolicy` returns early on nil. Delete.

## Duplication

- [ ] `api/common/status_types.go` — 4 near-identical `SetXCondition` methods
      differing only in `Type`. One private `setCondition` + 4 one-liners.
- [ ] `api/common/status_types.go` — `SetSummaryProgressingMessage` copies
      `SetSummaryProgressing`. Share one setter.
- [ ] `pkg/resourcegenerator/resourceutils/metadata.go` — shared-routing label
      map is also written in `pkg/gwapi/membership.go`. One helper, two callers.
- [ ] `pkg/metrics/usage/routing_provider.go` — `routingProviderFromObject` +
      `routingProviderOrLegacy` are one lookup with one default. One function.

## Speculative structure

- [ ] `pkg/gwapi/migration.go` — the 6-value `routingState` enum has 4 values
      (`LegacyOnly`, `GreenfieldPending`, `CutoverReadyPruneLegacy`,
      `StandardOnly`) that nothing reads, in code or tests. Behaviour needs one
      bit: stalled or not. The 6 names stay in the README for humans.
- [ ] `pkg/gwapi/readiness.go` — `legacyResourceExists` wraps one `Get` in
      `retry.OnError`. The function returns the error and both controllers
      requeue on it, so the requeue is the retry. Delete the wrapper.
- [ ] `pkg/gwapi/readiness.go` — the 5 probe helpers build success messages with
      `fmt.Sprintf` that `observeReadiness` throws away. Return
      `Readiness{Ready: true}`.
- [ ] `pkg/gwapi/routable.go` — `legacyRoutable` exists only to be embedded in
      `routablePlanner`, and both concrete types already have its two methods.
      Fold into `Routable`.
- [ ] `pkg/resourcegenerator/gatewayapi/gatewayapi.go` —
      `unsupportedRetryOptionFunc` is threaded 3 levels deep with `!= nil`
      guards for one real caller. Pass the logger instead.

## Small stuff

- [ ] `pkg/resourcegenerator/gatewayapi/application.go` — `backendRule` is called
      with a placeholder rule name, then `BackendRefs[0].Name` is overwritten.
      Pass the service name in.
- [ ] `internal/controllers/routing.go` — `SetSubresourceDefaults(resources, …)`
      sits inside the loop over `resources`, so it does a full pass per
      resource. Hoist above the loop. Pre-existing; this PR touches the loop.
- [ ] `pkg/gwapi/README.md` — 20 lines of prose restate the mermaid state list
      line by line. Keep the legend, cut the walkthrough.

## Out of scope, worth doing later

- [ ] `pkg/metrics/usage/` — `forEachRoutableResource` runs once per gauge, so
      each tick does 2 namespace lists and 4 CR lists. One sweep feeding both
      gauges is the same code and half the API load.
