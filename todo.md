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
- [x] `pkg/resourcegenerator/gatewayapi/routing.go` — the warn closure passed to
      `backendRule` can never fire: Routing passes `retries: nil` and
      `retryPolicy` returns early on nil. Delete.

## Duplication

- [x] `api/common/status_types.go` — 4 near-identical `SetXCondition` methods
      differing only in `Type`. One private `setCondition` + 4 one-liners.
- [x] `api/common/status_types.go` — `SetSummaryProgressingMessage` copies
      `SetSummaryProgressing`. Share one setter.
- [x] `pkg/resourcegenerator/resourceutils/metadata.go` — shared-routing label
      map is also written in `pkg/gwapi/membership.go`. One helper, two callers.
- [x] `pkg/metrics/usage/routing_provider.go` — `routingProviderFromObject` +
      `routingProviderOrLegacy` are one lookup with one default. One function.

## Speculative structure

- [x] `pkg/gwapi/migration.go` — the 6-value `routingState` enum collapsed to
      one `stalled` bool. Only 2 of the 6 values drove behaviour; the migration
      tests now build state through `determineRoutingState` instead of
      hand-writing result literals, which also removes literals for states the
      function cannot produce. The 6 names stay in the README for humans.
- [~] `pkg/gwapi/readiness.go` — `legacyResourceExists` wraps one `Get` in
      `retry.OnError`. Rejected: a requeue is not a free retry here, it first
      writes Ready=False plus a warning event for a blip. Kept, with the reason
      as a comment.
- [x] `pkg/gwapi/readiness.go` — the 5 probe helpers build success messages with
      `fmt.Sprintf` that `observeReadiness` throws away. Return
      `Readiness{Ready: true}`.
- [x] `pkg/gwapi/routable.go` — `legacyRoutable` exists only to be embedded in
      `routablePlanner`, and both concrete types already have its two methods.
      Fold into `Routable`.
- [x] `pkg/resourcegenerator/gatewayapi/gatewayapi.go` —
      `unsupportedRetryOptionFunc` is threaded 3 levels deep with `!= nil`
      guards for one real caller. Pass the logger instead.

## Small stuff

- [x] `pkg/resourcegenerator/gatewayapi/application.go` — `backendRule` is called
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
