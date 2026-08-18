# Improvements

An audit of Skiperator for over-engineering, runtime cost, and separation of
concerns. Every item keeps the CRD surface, the generated resources, and the
observable behavior the same. Where an item carries a behavior risk, the item
says so.

The audit ran as five parallel passes: reconciliation core, resource
generators, controllers and startup wiring, runtime performance, and the
practice of long-lived upstream operators.

> This audit ran on the branch `gateway-api/shared-hostname`.

## How to read this

Each item has a tag, a location, and a replacement.

| Tag      | Meaning                                                                   |
|----------|---------------------------------------------------------------------------|
| `delete` | Dead code or unused flexibility. Replacement: nothing.                    |
| `stdlib` | Hand-rolled code that the standard library or a Kubernetes library ships. |
| `native` | A dependency that does what the platform already does.                    |
| `yagni`  | An abstraction with one implementation or one caller.                     |
| `shrink` | The same logic in fewer lines.                                            |
| `perf`   | A runtime cost, not a complexity cost.                                    |

**Easy wins** are small and self-contained. One pull request each.
**Bigger tasks** need a design decision before the code changes.

---

## Headline numbers

One Application reconcile costs about **48 uncached API-server round trips**.
44 of them are List calls from the resource processor.

The processor lists `unstructured.UnstructuredList`. Controller-runtime does
not cache unstructured objects by default. The default is set at
`sigs.k8s.io/controller-runtime@v0.24.1/pkg/cluster/cluster.go:208`
(`Unstructured: false`), and the client honors it at
`pkg/client/client.go:259`. Every one of those 44 Lists leaves the process.

At 1000 Applications with the 10-second requeue active, that is about **4800
API requests per second**. Item B1 removes most of it.

The repository already holds a measurement baseline from an earlier pass, in
`../requests.txt` and `../requests-9-apps-manual.txt` (commit `4da7d2132`). Nine
idle Applications over ten minutes produced 1025 GET calls and 51 PUT calls.

Code volume: about 19,900 lines of Go. The items below remove roughly 1200
lines and 3 direct dependencies, without changing behavior.

---

## Easy wins

### E1. `perf` — give the Namespace controller an event filter

[namespace.go:40](../internal/controllers/namespace.go:40) builds the controller
with no `WithEventFilter`. Every namespace label change, annotation change, or
status change starts a reconcile. Each reconcile costs 10 uncached List calls.

Action: add
`predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})`.

This is the largest single cut in reconcile count.

### E2. `perf` — drop `NamespaceList` from the namespace schema set

[schemas.go:117](../pkg/resourceschemas/schemas.go:117) puts `&corev1.NamespaceList{}`
in the schema set that the processor lists.

`client.ListOptions{Namespace: ...}` is dropped for a cluster-scoped kind, at
`controller-runtime@v0.24.1/pkg/client/unstructured_client.go:245`. The
processor therefore scans every Namespace in the cluster, twice per reconcile.

Skiperator never generates a Namespace.

Action: remove the entry.

### E3. `perf` — list only Certificates when looking for Certificates

[diffs.go:31](../pkg/resourceprocessor/diffs.go:31) calls `getCertificates`, and
[crud.go:109](../pkg/resourceprocessor/crud.go:109) forwards it to
`listResourcesByLabels`, which loops over all 22 schemas in the
`istio-gateways` namespace.

This exactly doubles the List cost of every Application reconcile, to fetch
one kind.

Action: pass a schema slice that holds `CertificateList` only.

The in-source comment at [diffs.go:29](../pkg/resourceprocessor/diffs.go:29)
already flags the code as wrong.

### E4. `perf` — scope the SKIPJob fan-out to one namespace

[skipjob.go:95](../internal/controllers/skipjob.go:95) watches every
NetworkPolicy in the cluster.
[skipjob.go:301](../internal/controllers/skipjob.go:301) then lists **every**
SKIPJob in the cluster and enqueues all of them.

A guard at [skipjob.go:296](../internal/controllers/skipjob.go:296) limits this to
NetworkPolicies owned by an Application, so it does not fire on every event.
When it does fire with `MaxConcurrentReconciles: 1`, one access-policy edit
serializes a full-fleet sweep.

Action: add `client.InNamespace(object.GetNamespace())` to the List.

### E5. `perf` — set `MaxConcurrentReconciles` on the other three controllers

Only the Application controller wires the value, at
[application.go:140](../internal/controllers/application.go:140). The Routing,
SKIPJob, and Namespace controllers call `.Complete(r)` with no
`WithOptions`, so they run one worker each.

The configured default is 1, at [config.go:87](../internal/config/config.go:87).
The recorded baseline in `../requests.txt` shows no change at 5 or 50 workers.
The traffic is per reconcile, not per worker.
Fix B1 first, then raise this.

Action: add `WithOptions(controller.Options{MaxConcurrentReconciles: n})` to
the three controllers, so the setting means the same thing everywhere.

### E6. `perf` — check the ServiceMonitor CRD once, not once per reconcile

[application.go:507](../internal/controllers/application.go:507) runs an
uncached GET against the apiextensions API on every Application reconcile. A
failure calls `panic` at
[application.go:156](../internal/controllers/application.go:156).

The Gateway API CRDs already use the correct pattern: one check at startup, at
[main.go:337](../cmd/skiperator/main.go:337).

Action: move the ServiceMonitor check next to `requireGatewayAPI`.

### E7. `delete` — remove the Kubernetes 1.27 version gate

`../pkg/k8sfeatures/version.go` holds a package-level mutable variable
`currentVersion` and calls `os.Exit(1)` inside a library function.

Its only consumer is `EnhancedPDBAvailable()`, at
[pod_disruption_budget.go:70](../pkg/resourcegenerator/pdb/pod_disruption_budget.go:70).
That gates `spec.unhealthyPodEvictionPolicy` on Kubernetes 1.27 or later. The
field is generally available since Kubernetes 1.31, and the repository builds
against `k8s.io/api v0.36`.

Action: erase the file and the `NewVersionInfo` call at
[main.go:326](../cmd/skiperator/main.go:326). Always set the field.

Saves about 50 lines, one global variable, and one `os.Exit` in a library.

### E8. `perf` — stop copying objects the processor already owns

[crud.go:100](../pkg/resourceprocessor/crud.go:100) deep-copies every live object
into a slice that was never preallocated. `client.List` already returns freshly
decoded objects, so the copy is pure garbage.

The two maps at [diffs.go:35](../pkg/resourceprocessor/diffs.go:35) and
[diffs.go:40](../pkg/resourceprocessor/diffs.go:40) are also unpreallocated.

Action: remove the `DeepCopyObject` call. Preallocate with `len(schema.Items)`.

### E9. `perf` — hoist the ingress-gateway regular expression

[application.go:569](../internal/controllers/application.go:569) calls
`regexp.MatchString("^.*-ingress-.*$", gateway.Name)` for every Istio Gateway
event. The same file already shows the correct pattern at
[application.go:102](../internal/controllers/application.go:102), with
`regexp.MustCompile` at package level.

Action: use `strings.Contains(name, "-ingress-")`. The pattern is not a real
regular expression.

### E10. `perf` — let the usage metrics reuse the manager cache

[usage.go:45](../pkg/metrics/usage/usage.go:45) says it creates a client "in order
to utilize the built-in caching mechanisms", then passes
`client.Options{Cache: &client.CacheOptions{Unstructured: true}}` with no
`Cache.Reader`.

`controller-runtime@v0.24.1/pkg/client/client.go:204` returns an uncached
client when `Cache.Reader` is nil. Nothing is cached. The sweeper then runs
four unfiltered cluster-wide List calls every 30 seconds, at
[sweep.go:56](../pkg/metrics/usage/sweep.go:56) and
[sweep.go:75](../pkg/metrics/usage/sweep.go:75).

Action: pass `mgr.GetClient()` into `NewUsageMetrics`. Use typed lists.

### E11. `delete` — remove dead exported code

| What                                                                                                                   | Where                                                           | Lines |
|------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------|-------|
| `ErrIsMissingOrNil`, zero callers                                                                                      | [helperfunctions.go:101](../pkg/util/helperfunctions.go:101)       | 12    |
| `HasUpperCaseLetter`, zero callers                                                                                     | [helperfunctions.go:140](../pkg/util/helperfunctions.go:140)       | 9     |
| `GetService`, zero callers                                                                                             | [helperfunctions.go:93](../pkg/util/helperfunctions.go:93)         | 7     |
| `PointToInt64`, zero callers                                                                                           | [helperfunctions.go:125](../pkg/util/helperfunctions.go:125)       | 3     |
| `Processor` interface, zero implementations                                                                            | [processor.go:12](../pkg/resourceprocessor/processor.go:12)        | 3     |
| `ResourceProcessor.scheme`, written, never read. **Skip this row. B1 gives the field a use** | [processor.go:20](../pkg/resourceprocessor/processor.go:20)        | 2     |
| `SubResourceError.Retryable`, never set outside its test                                                               | [errors.go:21](../pkg/reconciliation/errors.go:21)                 | 2     |
| `SetSummaryPending`, zero callers                                                                                      | [status_types.go:93](../api/common/status_types.go:93)             | 8     |
| Four `SKIPJobCustomDefaulter` fields, written, never read                                                              | [skipjob_webhook.go:21](../internal/webhook/skipjob_webhook.go:21) | 11    |
| Empty package directory                                                                                                | `../pkg/resourcegenerator/networkpolicy/cloudsql`               | 0     |

`ResourceProcessor.scheme` also removes a constructor parameter that five call
sites pass twice.

Note: the `../pkg/util` entries are exported Go symbols. They break no operator
behavior and no CRD. They do change the Go package surface, so make sure that
no other repository imports `github.com/kartverket/skiperator/pkg/util` before
you delete them.

### E12. `delete` — remove the unreachable type guards in generators

Six generators check a type that the registry already selected for them. One
example: [certificate/routing.go:19](../pkg/resourcegenerator/certificate/routing.go:19)
returns an error for a non-Routing type, but
[certificate/certificate.go:12](../pkg/resourcegenerator/certificate/certificate.go:12)
registers it for `RoutingType` only.

The same holds at `certificate/application.go:22`,
`istio/gateway/application.go:23`, `istio/gateway/routing.go:23`,
`prometheus/pod_monitor.go:23`, and `prometheus/service_monitor.go:21`.

Item B6 removes the registry, which removes these guards as a side effect.

### E13. `shrink` — `compareObject` decides nothing

[diffs.go:82](../pkg/resourceprocessor/diffs.go:82) compares group, version, kind,
namespace, and name. The map key at
[diffs.go:37](../pkg/resourceprocessor/diffs.go:37) is built from exactly those
values, so a key match already proves the comparison true.

Every matched object therefore reaches `shouldUpdate` or `shouldPatch`, and
the real work happens later inside `update` and `patch`.

Action: delete the function and its call.

### E14. `stdlib` — return one joined error, not a count

`ResourceProcessor.Process` returns `[]error` at
[processor.go:28](../pkg/resourceprocessor/processor.go:28). All three controllers
collapse it to `fmt.Errorf("found %d errors", len(errs))`, at
[application.go:353](../internal/controllers/application.go:353),
[routing.go:223](../internal/controllers/routing.go:223), and
[skipjob.go:224](../internal/controllers/skipjob.go:224). The error text loses
every cause.

Action: return `errors.Join(errs...)`.

### E15. `stdlib` — replace the hand-rolled helpers

| Hand-rolled                                              | Replacement                                             | Where                                                                      |
|----------------------------------------------------------|---------------------------------------------------------|----------------------------------------------------------------------------|
| `util.PointTo`, 26 call sites                            | `new(v)`, the language builtin                          | [helperfunctions.go:121](../pkg/util/helperfunctions.go:121)                  |
| `hasExternalIngress`, `hasInternalIngress`               | `slices.ContainsFunc`                                   | [common.go:264](../pkg/resourcegenerator/networkpolicy/dynamic/common.go:264) |
| `routeHasHostname` manual lowercase loop                 | `strings.EqualFold`, as used three lines below          | [conflicts.go:174](../pkg/gwapi/conflicts.go:174)                             |
| `withFallback[T ~string]`, one caller                    | `cmp.Or`                                                | [idporten.go:162](../pkg/resourcegenerator/idporten/idporten.go:162)          |
| `MatchesPredicate[T]`, unchecked type assertion          | `predicate.NewTypedPredicateFuncs[T]`                   | [predicates.go:1](../pkg/util/predicates.go:1)                                |
| `../pkg/util/array`, 3 generics, 2 callers, one is O(n²) | inline loops, `slices.Sorted` and `slices.Compact`      | [array/main.go:1](../pkg/util/array/main.go:1)                                |
| `GetPodVolumes` insertion-ordered map                    | the `seen` set already in `pod.AppendUniqueVolumes:180` | [volumes.go:88](../pkg/resourcegenerator/volume/volumes.go:88)                |

`../pkg/util/array` also duplicates a pattern that lives two packages over.

### E16. `native` — drop `github.com/pkg/errors`

One import remains, at [usage.go:8](../pkg/metrics/usage/usage.go:8). The package
is archived. `fmt.Errorf` with the `%w` verb replaces it.

### E17. `delete` — remove the stale requeue comment

[util.go:27](../internal/controllers/common/util.go:27) says
`// TODO: exponential backoff`. Controller-runtime already applies exponential
backoff, from 5 ms to 1000 s, to a reconcile that returns an error.

The comment describes work the framework does. See B4 for the rate limiter
that is worth setting.

### E18. `shrink` — clean up the Makefile

| Target               | Problem                                                                                        | Action          |
|----------------------|------------------------------------------------------------------------------------------------|-----------------|
| `run-unit-tests`     | Reimplements the exit status of `go test` with `grep` and `awk`, and hides build failures      | `go test ./...` |
| `check-kind`         | Guards a command that the next line runs anyway                                                | Delete          |
| `benchmark-long-run` | 73 lines, hardcodes nine manifests twice, and a stray `@` at line 306 makes it fail as written | Delete          |

[Makefile:213](../Makefile:213), [Makefile:93](../Makefile:93),
[Makefile:296](../Makefile:296).

`patch-skipjob-crd` and `ensure-kubectl` are genuine. Keep them.

### E19. `perf` — make the workload predicates cheap, and drop a dependency

[predicates.go:26](../internal/controllers/common/predicates.go:26) and
[predicates.go:48](../internal/controllers/common/predicates.go:48) run, per
Deployment or StatefulSet event: two `DeepCopyObject` calls, then two
`hashstructure.Hash` reflection walks over a whole PodSpec.

The predicate itself is **correct and load-bearing**. It masks
`Spec.Replicas` so that HPA scaling does not start a reconcile. Keep that.
The comment at [predicates.go:39](../internal/controllers/common/predicates.go:39)
records the accepted trade-off, that a manual replica edit no longer
reconciles. Keep that too.

Only the implementation is heavy. A shallow struct copy plus one semantic
comparison does the same job:

```go
oldDep, ok1 := e.ObjectOld.(*appsv1.Deployment)
newDep, ok2 := e.ObjectNew.(*appsv1.Deployment)
if !ok1 || !ok2 {
    return true
}
// HPA must not trigger reconciles: compare with replicas masked.
masked := newDep.Spec              // struct copy, does not mutate newDep
masked.Replicas = oldDep.Spec.Replicas
return !equality.Semantic.DeepEqual(oldDep.Spec, masked) ||
    !maps.Equal(oldDep.Labels, newDep.Labels)
```

`GetHashForStructs` at [helperfunctions.go:37](../pkg/util/helperfunctions.go:37)
has no other caller, so this also removes the dependency
`github.com/mitchellh/hashstructure/v2`.

The current code also calls `panic(err)` on a hash error, inside a predicate.

### E20. `delete` — move the measurement notes out of the repository root

`../requests.txt` and `../requests-9-apps-manual.txt` hold API-call counts from an
earlier optimization pass. The numbers are useful. Their location is not.

Action: move both into `doc/` as a baseline record.

---

## Bigger tasks

### B1. `perf` — stop listing unstructured objects in the resource processor

**This is the largest win in the audit.**

[crud.go:96](../pkg/resourceprocessor/crud.go:96) lists 22
`unstructured.UnstructuredList` values per reconcile. None of them are cached.

Action: use typed lists. The typed informers already exist, because the
controllers already `Owns()` these kinds. The Lists become cache reads at no
memory cost.

Do not reach for `client.CacheOptions{Unstructured: true}` instead.
Controller-runtime keeps **separate informer maps** for structured and
unstructured objects, at
`controller-runtime@v0.24.1/pkg/cache/internal/informers.go:132-134`. Every
type that a controller already watches typed then gets cached a second time,
as unstructured. Memory roughly doubles, and it buys nothing that typed lists
do not already give.

The typed lists are **already written**. `GetApplicationSchemas` at
[schemas.go:64](../pkg/resourceschemas/schemas.go:64) builds a
`[]client.ObjectList` of 22 typed lists. Then `addGVKToList` at
[schemas.go:50](../pkg/resourceschemas/schemas.go:50) throws every one of them
away. It keeps only the GVK, inside an empty `unstructured.UnstructuredList`.

The file header states the reason, at
[schemas.go:4](../pkg/resourceschemas/schemas.go:4). A typed List does not
populate `TypeMeta` on its items, so the GVK stays empty. The processor keys
its diff map on that GVK, at
[diffs.go:37](../pkg/resourceprocessor/diffs.go:37).

That reason is solvable in two lines. The processor knows which list it is
iterating, so it knows the GVK.

```go
for _, list := range r.schemas {           // []client.ObjectList
    if err := r.client.List(ctx, list, listOpts); err != nil { return err }
    gvk, _ := apiutil.GVKForObject(list, r.scheme)
    _ = meta.EachListItem(list, func(o runtime.Object) error {
        obj := o.(client.Object)
        obj.GetObjectKind().SetGroupVersionKind(gvk)   // what unstructured gave for free
        *objList = append(*objList, obj)
        return nil
    })
}
```

`meta.EachListItem` ships in `k8s.io/apimachinery/pkg/api/meta`.

This also gives `ResourceProcessor.scheme` a real use, so drop that row from
E11.

Expected effect: uncached round trips per Application reconcile fall from
about 48 to about 4.

### B2. `stdlib` — replace the hand-rolled apply logic with `CreateOrUpdate`

`../pkg/resourceprocessor/crud.go` implements create, update, and patch by hand.
The three functions call each other in a cycle, at
[crud.go:19](../pkg/resourceprocessor/crud.go:19),
[crud.go:33](../pkg/resourceprocessor/crud.go:33), and
[crud.go:54](../pkg/resourceprocessor/crud.go:54). Together they re-implement
create-or-update.

The no-op check re-implements equality with reflection and a hash, at
[crud.go:114-143](../pkg/resourceprocessor/crud.go:114). Two defects follow from
it:

1. **Annotations are never compared.** The code says so, at
   [crud.go:113](../pkg/resourceprocessor/crud.go:113).
2. **Objects with no `Spec` field are always written.** `getSpecHash` returns
   an empty string for them, at
   [crud.go:134](../pkg/resourceprocessor/crud.go:134). The caller then reads an
   empty hash as "changed", at
   [crud.go:122](../pkg/resourceprocessor/crud.go:122). ConfigMap, Secret, and
   ServiceAccount all take a GET and a PUT on every reconcile.

Action: replace all three functions with `controllerutil.CreateOrUpdate` and
`equality.Semantic.DeepEqual`. Both ship in libraries the repository already
depends on. That removes the reflection, the hash, the recursion, and both
defects, and it keeps the current ownership and rollout behavior.

This step is not the destination. It is the part that carries no migration
risk, and it must land first. Server-side apply is item B17.

**Do not keep the hash either.** prometheus-operator keeps one, so the pattern
is legitimate at scale. But it hashes **named, declared inputs**:
`createSSetInputHash(p, config, ruleConfigMapNames, tlsAssets, existingStatefulSet.Spec)`,
stored under a `prometheus-operator-input-hash` annotation, with tests that
assert hash stability across semantically equal inputs
([pkg/operator/kubernetes.go](https://github.com/prometheus-operator/prometheus-operator/blob/main/pkg/operator/kubernetes.go)).

Skiperator instead reflects over whichever field happens to be named `Spec`.
No upstream project does that. `equality.Semantic.DeepEqual` needs no hash at
all.

This removes about 150 lines. `update` and `patch` currently re-fetch
objects that `getDiff` already holds, at
[diffs.go:26](../pkg/resourceprocessor/diffs.go:26).

### B3. `perf` — restrict the informer cache

`ctrl.Options` at [main.go:169](../cmd/skiperator/main.go:169) sets no `Cache`
field. No namespace, label, or field restriction exists anywhere in the
repository.

Cached cluster-wide today: **every Secret** (from
[application.go:129](../internal/controllers/application.go:129) and
[namespace.go:44](../internal/controllers/namespace.go:44)) and **every
ConfigMap** (from [application.go:110](../internal/controllers/application.go:110)),
plus about 25 other kinds.

Across 500 namespaces this holds every service-account token, every Helm
release secret, and every TLS secret in memory. It is usually the largest item
on the heap.

The consumers are already narrow. The Secret map function wants only
`type` containing `digdirator.nais.io`, at
[application.go:533](../internal/controllers/application.go:533). The Namespace
controller wants only `github-auth`.

Action:

1. Add `Cache: cache.Options{ByObject: {...}}` with a label selector for
   Secret and ConfigMap.
2. Add a `TransformFunc` that strips `managedFields` and
   `kubectl.kubernetes.io/last-applied-configuration` from the large types.

### B4. `perf` — make Gateway API readiness event-driven

[application.go:358](../internal/controllers/application.go:358) and
[routing.go:234](../internal/controllers/routing.go:234) return
`RequeueAfter: 10 * time.Second` while standard routing is not ready.

The polling loop is load-bearing, not decorative. The event filter at
[application.go:131](../internal/controllers/application.go:131) uses
`predicate.Or(GenerationChangedPredicate{}, LabelChangedPredicate{})`. That
filter drops status-only updates on ListenerSet, HTTPRoute, and Certificate.
Those updates are the ones that report readiness.

During a fleet migration of 1000 Applications, this is 6 reconciles per minute
per object, and each one pays the full uncached List cost.

Action:

1. Apply `predicate.ResourceVersionChangedPredicate{}` per source on the
   Gateway API `Owns()` registrations, instead of one filter for the whole
   controller.
2. Reduce the timer to a long safety net.
3. Set an explicit `RateLimiter` in `controller.Options`. The default bucket
   is 10 queries per second, shared across the fleet, which is thin for 1000
   objects during a registry outage.

### B5. `perf` — write status once, and stop the timestamp churn

Two problems compound here.

**Two status writes per reconcile.** `SetProgressingState` writes status
before any work happens, at
[application.go:265](../internal/controllers/application.go:265), through a
`RetryOnConflict` loop that adds its own GET, at
[reconciler.go:171](../internal/controllers/common/reconciler.go:171). The
terminal path writes again at
[application.go:378](../internal/controllers/application.go:378). The same shape
appears in `routing.go:132` and `skipjob.go:174`.

**Every status value changes every reconcile.** `Status.TimeStamp` is set from
`metav1.Now().String()` in all five summary setters and in
`AddSubResourceStatus`, at
[status_types.go:96](../api/common/status_types.go:96) and
[status_types.go:177](../api/common/status_types.go:177).

The second problem produced the workaround for the first.
`filterOutStatusTimestamps`, at
[util.go:188](../internal/controllers/common/util.go:188), exists to hide the
churn that the operator itself creates, and it needs a whole diffing library
to do it.

Action:

1. Keep the previous `TimeStamp` when the status value and message do not
   change.
2. Buffer conditions and write status once, at the end.
3. Then `equality.Semantic.DeepEqual` replaces `GetObjectDiff`, and the
   dependency `github.com/r3labs/diff/v3` goes away.

`GetObjectDiff` is also worth reading on its own, at
[util.go:171](../internal/controllers/common/util.go:171). Its guard compares
`reflect.Kind` of two values of the same type parameter, so the guard can never
fire. Four of its five callers only test `len(diff) > 0`.

### B6. `delete` — remove the generator registry

`../pkg/resourcegenerator/resourceutils/generator/multigenerator.go` holds a
`map[ObjectType]func(Reconciliation) error`. Fourteen `init()` functions
register into it. Seven packages carry a one-line dispatcher file whose whole
body is a map lookup.

The lookup re-derives a fact that the compiler already knows. The controllers
build type-specific slices at
[application.go:292](../internal/controllers/application.go:292),
[routing.go:185](../internal/controllers/routing.go:185), and
[skipjob.go:185](../internal/controllers/skipjob.go:185). An Application-only
slice already names `certificate.Generate`.

The codebase now has three dispatch styles for one job. `defaultdeny` and
`imagepullsecret` use a struct with a method instead.

Action: export `GenerateForApplication`, `GenerateForRouting`, and
`GenerateForSKIPJob`. Call them directly.

Removes about 150 lines, plus the six unreachable guards in E12, plus every
`init()` side effect.

### B7. `yagni` — move validation and defaulting out of the Application controller

`../internal/controllers/application.go` is 781 lines. **295 of them, 38%, are
not controller work.**

| Lines   | What                                                                                               | Belongs in         |
|---------|----------------------------------------------------------------------------------------------------|--------------------|
| 365-427 | Status formatting, and a status-name value plumbed only for one comparison                         | `../api/common`    |
| 480-505 | `setApplicationDefaults`, plus the spec write-back dance at 183-210 that exists only to persist it | Mutating webhook   |
| 574-623 | `validateIngresses`, `validateExtraContainers`                                                     | Validating webhook |
| 625-743 | Digdirator auth-config assembly. No controller-runtime concept appears in it                       | `../pkg/auth`      |
| 745-781 | `validateStatefulUnchanged`, `validateApplicationStatefulFields`                                   | Validating webhook |

`../internal/webhook` already exists and is wired through `EnableWebhooks`.
`SKIPJobCustomDefaulter.Default` at
[skipjob_webhook.go:48](../internal/webhook/skipjob_webhook.go:48) is the pattern
to copy.

The `hostMatchExpression` regular expression is also expressible as a CRD CEL
rule, the way `spec.routingProvider` already is at
[application_types.go:75](../api/v1alpha1/application_types.go:75).

Note: a webhook rejects a bad manifest at admission, whereas the controller
today accepts it and reports an error condition. That is a user-visible change
in *when* the error appears. Decide it deliberately.

### B8. `shrink` — share one pod template between Deployment and StatefulSet

`deployment.go:47-174` and `statefulset.go:47-173` hold **about 90 identical
lines**. A sorted line comparison gives that number. The shared part covers:

- pod options, container build, and `CreatePodSpec`
- GCP volumes and Cloud SQL
- ID-porten and Maskinporten secrets
- pod template labels, generated-spec annotations, and `SetApplicationLabels`
- extra containers

The StatefulSet version adds volume-claim templates. Everything else matches.

Action: extract `buildPodTemplate(r, application) (corev1.PodTemplateSpec, error)`
into `../pkg/resourcegenerator/pod`.

### B9. `shrink` — collapse the repeated controller scaffolding

Each of the three controllers holds its own copy of the same five blocks.

| Duplicated block                                                                           | Copies | Lines |
|--------------------------------------------------------------------------------------------|--------|-------|
| Generator loop with `SubResourceError` unwrapping                                          | 3      | 28    |
| Processor construction and error drain                                                     | 3      | 20    |
| Typed getter with `IsNotFound` handling. The SKIPJob copy says "routing" in its error text | 3      | 22    |
| Certificate map function, differing only in one label value                                | 2      | 22    |
| `teamNameForNamespace`, byte-identical                                                     | 2      | 12    |
| Finalizer removal. Both finalizer constants hold the same string                           | 2      | 12    |
| `cleanUpWatchedResources`                                                                  | 2      | 10    |

`updateApplicationStatus` at
[application.go:381](../internal/controllers/application.go:381) is a retype of
`ReconcilerBase.UpdateStatus` at
[reconciler.go:166](../internal/controllers/common/reconciler.go:166).
`Application` already satisfies `common.SKIPObject`.

Four near-identical SKIPJob condition constructors at
[skipjob.go:316](../internal/controllers/skipjob.go:316) reduce to one closure and
a four-row table.

### B10. `shrink` — one generic guard for the generators

Twenty `if r.GetType() != reconciliation.XType` guards and twenty-four
`failed to cast resource to application` blocks are copy-pasted across the
generators. They are identical apart from the resource noun.

Action: add `skipObject[T](r) (T, error)`. The repository is Go 1.26.

About 110 lines.

### B11. `perf` — fix the image digest resolution

[deployment.go:217](../pkg/resourcegenerator/deployment/deployment.go:217) and
[statefulset.go:205](../pkg/resourcegenerator/statefulset/statefulset.go:205) call
`util.ResolveImageTags` on every reconcile.

Inside, [digest.go:14](../pkg/util/digest.go:14) marshals the Deployment to JSON,
converts it to a YAML node, and marshals it back. That is three full
serializations of a whole Deployment.

Then `k8s-digester@v0.1.16/pkg/keychain/keychain.go:54` builds a **brand-new
`kubernetes.Clientset`**, which means a new HTTP transport and a new TLS
handshake. On top of that come uncached ServiceAccount and image-pull-secret
reads, and one external registry round trip.

Action:

1. Build the clientset and the keychain once at startup.
2. Cache resolved digests, keyed by image reference.

Related dependency question: `k8s-digester` pulls in 105 cloud-SDK packages
(AWS, Azure, and Google Cloud) out of the binary's 1173 total, through
`go-containerregistry/pkg/authn/k8schain`. The built binary is about 39 MB. The
repository already depends on `go-containerregistry` directly.

Measure the binary-size and CVE-surface saving before deciding. Do not remove
the dependency for its own sake.

### B12. `yagni` — remove the two-implementation planner interface

[planner.go:16](../pkg/gwapi/planner.go:16) defines `routablePlanner`. It has
exactly two implementations, each a wrapper struct that embeds a concrete type
and forwards two methods.

Action: two `switch obj := routable.(type)` blocks, inline in
`ValidateConflicts` and `observeStandardRouting`. About 45 lines.

### B13. `yagni` — delete the unreachable retry translation

`pkg/resourcegenerator/gatewayapi/gatewayapi.go:269-401` translates Istio
retries to Gateway API. The author documents it as unreachable at
[gatewayapi.go:280](../pkg/resourcegenerator/gatewayapi/gatewayapi.go:280),
because a CEL rule at
[application_types.go:75](../api/v1alpha1/application_types.go:75) rejects the
combination.

About 135 lines of code and 84 lines of test.

This is a product decision, not a cleanup. Gateway API retry is
experimental-channel only today. Delete the code. Re-add it when Gateway API
retry graduates to the standard channel.

### B14. `perf` — index the Gateway API conflict lookups

[conflicts.go:24](../pkg/gwapi/conflicts.go:24),
[conflicts.go:68](../pkg/gwapi/conflicts.go:68), and
[conflicts.go:93](../pkg/gwapi/conflicts.go:93) run unfiltered cluster-wide Lists
of ListenerSets and HTTPRoutes per reconcile. A Routing hits two of them in one
pass.

These are typed, so the cache serves them. Each one still copies every object
in the cluster, and the triple loop at
[conflicts.go:27](../pkg/gwapi/conflicts.go:27) is O(hosts × listenerSets ×
listeners).

Action: `SharedRoutingLabels` already writes a
`skiperator.kartverket.no/hostname` label. Add a field index and use
`client.MatchingLabels`.

### B15. Test the pure generators in Go, not on a cluster

`../tests` holds **422 Chainsaw YAML files**. They need a live cluster.

Meanwhile 24 generator packages hold **zero Go unit tests**, including
`statefulset`, `job`, `service`, `pdb`, `hpa`, `prometheus`, `idporten`,
`maskinporten`, and every Istio authorization-policy generator.

`../AGENTS.md` line 43 states the rule that makes this cheap to fix: "Generator
functions must be pure — no API calls, no side effects."

Two exceptions break that rule today. Both are listed in the architecture notes
below.

Naiserator already runs the exact test shape this needs: `goldenfile.Run` over
**140 YAML files** in
[pkg/resourcecreator/testdata](https://github.com/nais/naiserator/tree/master/pkg/resourcecreator/testdata).
Each file holds an input CR, any pre-existing cluster objects, and the expected
output. No cluster. File names are behaviour-scoped, such as
`application_zero_replicas.yaml` and `deployment_command.yaml`.

That test suite is possible only because their generators are pure functions.
It is the best idea in that repository, and it is the one directly transferable
thing.

Action:

1. Fix the two purity leaks (B11 and the `RoutingReconciliation` cast).
2. Add golden-file tests per generator, in naiserator's shape.
3. Keep Chainsaw for the flows that genuinely need a cluster.

Do this **before** touching the apply layer. It is the safety net that makes
B1, B2, and B16 cheap to attempt.

### B16. `perf` — stop discovering children, and the 22 Lists disappear

**Read this before doing B1.** B1 makes the 22 Lists cheap. B16 removes the
need for them. If B16 is accepted, B1 becomes a smaller change.

The research pass found **no CNCF-graduated operator that label-lists many
kinds on the reconcile path**. The only match is naiserator, which shares
Skiperator's ancestry, so it is not independent evidence.

**Why the Lists exist.** Owner references already handle parent deletion.
[refs.go:9](../pkg/resourcegenerator/resourceutils/refs.go:9) sets a controller
owner reference on every same-namespace child, and Kubernetes garbage
collection removes them when the Application goes away.

The Lists exist for the other case: a child that is **no longer generated while
the parent still lives**. A user removes an ingress, so its Certificate and its
ListenerSet must go. Garbage collection cannot see that. The name is derived
from the removed hostname, so it cannot be recomputed from the new spec either.

Upstream solves this with an inventory of what was applied last time. Flux
stores it in `status.inventory`
([inventory_types.go](https://github.com/fluxcd/kustomize-controller/blob/main/api/v1/inventory_types.go)).
cluster-api uses a sidecar object,
[ClusterResourceSetBinding](https://github.com/kubernetes-sigs/cluster-api/blob/main/api/addons/v1beta2/clusterresourcesetbinding_types.go).
Pruning becomes a diff against that record. One object read, zero cross-kind
Lists.

**Skiperator already writes this inventory.** `status.subresources` is a
`map[string]Status`, written for every applied object at
[processor.go:61](../pkg/resourceprocessor/processor.go:61), keyed as
`Kind[name]` at
[status_types.go:173](../api/common/status_types.go:173). It records exactly the
set that the next reconcile needs to diff against. Nothing reads it.

One gap: the key holds kind and name, not namespace. The cross-namespace
certificates in `istio-gateways` can collide with a same-named object in the
namespace of the Application.

Action:

1. Add a new status field holding a namespaced inventory. Do not widen the
   `subresources` key, because that changes what users already see in
   `status.subresources`. An added field is not an API break.
2. Prune by diffing the inventory against the objects generated this pass.
3. Keep owner references. They still cover parent deletion, and they cover the
   case where the operator was down when the parent was deleted.
4. Keep one label-list as a slow reconciliation path, on a long timer, to catch
   an inventory that drifted. Do not run it per reconcile.

This removes 22 Lists per Application reconcile, which is the same target as
B1, but it removes the work rather than making it cached.

Effort is higher than B1 and the risk is real: an inventory that is wrong
orphans resources or deletes live ones. Land B1 first for the immediate win.
Treat this item as the direction of travel, not as a next step.

### B17. `native` — move the write path to server-side apply

This is where B2 goes next. B2 removes the hand-rolled equality check.
B17 removes the read that precedes the write.

`controllerutil.CreateOrUpdate` still costs one GET per object per reconcile,
because the client decides equality. Server-side apply moves that decision to
the API server:

```go
client.Patch(ctx, obj, client.Apply, client.ForceOwnership, client.FieldOwner("skiperator"))
```

The API server then chooses create against update, and it no-ops an unchanged
object with no read at all.

Controller-runtime v0.24.1, the pinned version, also ships a typed
`client.Apply(ctx, applyConfiguration, ...)` on the `Writer` interface, at
`controller-runtime@v0.24.1/pkg/client/interfaces.go:65`. It arrived in v0.22.0
through [PR 3253](https://github.com/kubernetes-sigs/controller-runtime/pull/3253).
It needs generated `ApplyConfiguration` types, which `kubebuilder create api --ssa`
scaffolds. The `client.Patch` form above needs no code generation, so it comes
first.

**The argument against, stated fairly.** The kubebuilder book says: "For
controllers that are the sole owner of a resource and manage the entire
object, traditional `Update()` is simpler and sufficient"
([book.kubebuilder.io/reference/server-side-apply](https://book.kubebuilder.io/reference/server-side-apply)).
Skiperator *is* the sole owner of every object it generates, so apply buys no
co-ownership benefit. What it buys is the removed GET, and field *deletion*
without keeping state in an annotation. cert-manager names that second point
as its main reason, in
[design/20220118.server-side-apply.md](https://github.com/cert-manager/cert-manager/blob/master/design/20220118.server-side-apply.md).

CAUTION: server-side apply takes over `managedFields` on first rollout. Fields
that Skiperator stops setting become removable. That is the intended behavior
and it is still a behavior change. Do the first rollout on a sandbox cluster.

Preconditions. Do not start B17 until all four hold:

1. B15 is in place. The client-go fake client cannot do an Apply patch, so the
   generator tests must not depend on it.
2. B2 has landed and is stable. The two write-amplification defects are gone
   before ownership changes.
3. Every CRD list field carries correct `listType` and `listMapKey` markers.
   cluster-api added a linter for exactly this, in
   [PR 13340](https://github.com/kubernetes-sigs/cluster-api/pull/13340).
4. A feature gate exists, so the rollout is reversible per cluster.
   cert-manager shipped apply behind `ServerSideApply` in v1.8. Copy that.

The upstream cost table further down lists what each of these prevents. Read
it before starting. cluster-api found apply expensive enough to cache its own
no-op applies. Measure the GET saving against the apply cost on a real fleet,
with the ratio in the measurement plan.

---

## Architecture notes

### The flow today

A controller fetches the CR, runs validation gates, and builds a
`reconciliation.Reconciliation`. That value is a mutable bag: context, logger,
mesh mode, `rest.Config`, the object, and a growing `resources []client.Object`
slice.

The bag is passed to a fixed list of `Generate(r) error` functions. Each one
builds objects and calls `r.AddResource(obj)`. The controller stamps GVK,
labels, annotations, and owner references onto the whole slice. It then builds
a fresh `ResourceProcessor` and calls `Process`. `Process` lists the live
objects, buckets them into create, update, patch, and delete, then applies
each bucket.

The layering is sound. Four places leak.

### Where the layering leaks

**1. Generators reach the API server.**
`deployment.go:217` and `statefulset.go:205` call `util.ResolveImageTags`,
which builds its own Kubernetes client from `rest.Config`. `GetRestConfig()`
and `GetCtx()` exist on the `Reconciliation` interface only to serve that one
call. The interface carries its own note about it, at
[reconciliation.go:26](../pkg/reconciliation/reconciliation.go:26).

Result: generation is neither pure nor testable offline, contrary to
`../AGENTS.md` line 43.

**2. A generator casts back to a concrete reconciliation type.**
[routing.go:37](../pkg/resourcegenerator/networkpolicy/dynamic/routing.go:37)
reads `if routingReconciliation, ok := r.(*reconciliation.RoutingReconciliation); ok`.
One generator punctures the interface, which is the whole reason the interface
exists.

**3. The write layer knows domain kinds.**
`requirePatch` and `preparePatch`, at
[resource.go:21](../pkg/resourceprocessor/resource.go:21), type-switch on
Deployment, StatefulSet, and Job, and encode HPA replica semantics and
`kubectl rollout restart` semantics. That is generator knowledge living in the
processor. `getCertificates` hardcodes `mesh.GatewayNamespace`, flagged in
source as ugly at [diffs.go:29](../pkg/resourceprocessor/diffs.go:29).

**4. The ignore label is implemented twice.**
Once for events, in `common.ShouldReconcile` at
[util.go:32](../internal/controllers/common/util.go:32). Once for diffing, in
`shouldIgnoreObject` at [diffs.go:111](../pkg/resourceprocessor/diffs.go:111).
Both call the same metrics functions.

### The status layer feeds nothing on the current API version

`SkiperatorStatus` carries `Summary`, `SubResources`, `Conditions`,
`MigrationStartedAt`, and `AccessPolicies`. Around them sit 6 `StatusNames`,
9 condition types, 5 summary setters, and 5 condition setters. A
`SortConditions` function exists only because `meta.SetStatusCondition`
appends in first-seen order.

What a user actually sees, through printer columns:
`.status.conditions[Ready].status`, `.reason`, `.status.accessPolicies`,
`.status.applicationKind`, and `.spec.routingProvider`.

`SubResources` is written by the processor and never printed. `Summary` is
printed only on the deprecated v1alpha1 SKIPJob. v1beta1 dropped it.

This does not need a fix today, because the fields are part of the published
API. It does mean that B5 can stop maintaining the timestamps in that layer
with no user-visible loss.

### What is correctly built

Do not "fix" these.

- **The conversion webhook.** `v1beta1.SKIPJob` is the hub, `v1alpha1` converts
  to it. That matches upstream convention.
- **Ownership and garbage collection.** `SetControllerReference` plus explicit
  orphan deletion for the cross-namespace resources the garbage collector
  cannot reach.
- **Legacy and standard routing side by side.** Neither path is dead.
  `spec.routingProvider` is a user-set CRD field with
  `+kubebuilder:default=Legacy`, and `pkg/gwapi` runs a real migration state
  machine. No cut is available until `Legacy` leaves the enum.
- **`patch-skipjob-crd` in the Makefile.** It merges a multi-version CRD that
  `controller-gen` cannot emit.
- **The comments in `pkg/metrics/usage/sweep.go`.** They explain *why*, which
  is what a comment is for.
- **Config fields.** All 14 `SkiperatorConfig` fields are read.

---

## Comparison with mature operators

### What the canonical guidance says

**Reconcile the whole world, every time.** The controller-runtime FAQ answers
the question "can I write logic per event type" with "**A**: You should not.
Reconcile functions should be idempotent, and should always reconcile state by
reading all the state it needs, then writing updates"
([FAQ.md](https://github.com/kubernetes-sigs/controller-runtime/blob/main/FAQ.md)).

Skiperator obeys this. Keep it. It is also the reason B1 matters: a reconcile
that reads the whole world must read it from cache, not from the API server.

**One controller per kind.** Both the kubebuilder book
([good-practices](https://book.kubebuilder.io/reference/good-practices)) and
the operator-sdk best practices
([common-recommendation](https://sdk.operatorframework.io/docs/best-practices/common-recommendation/))
state this. Skiperator obeys it: four controllers, four kinds.

**`phase`-style status fields are deprecated.** The Kubernetes API conventions
say so directly: "The pattern of using `phase` is deprecated. Newer API types
should use conditions instead… adding new enum values breaks backward
compatibility"
([api-conventions.md](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties)).

Skiperator's `Summary.Status` and its six `StatusNames` values are exactly that
deprecated shape. The project already moved the right way, by adding a `Ready`
condition (commit `574f2dd60`). The old layer stays for compatibility. Item B5
stops maintaining timestamps in it, which is the part that costs writes.

The same document says that conditions "are observations and not, themselves,
state machines". It requires `Reason`, in CamelCase. It also requires condition
types to be adjectives or past-tense verbs, not present-tense. The Skiperator
types `InternalRulesValid`, `StandardRoutingReady`, and `SharedRoutingResources`
all fit.

**Server-side apply is conditional, not automatic.** The kubebuilder book
recommends `Update()` for a sole owner. B2 takes that route first. B17 moves
to apply once its four preconditions hold.

### What a large rewrite learned, and it applies here

The Operator Lifecycle Manager rewrote itself from v0 to v1. Its retrospective
is unusually blunt and one section is titled **"Do not fight Kubernetes"**
([olmv1_design_decisions.md](https://operator-framework.github.io/operator-controller/project/olmv1_design_decisions/)).

Two of its conclusions map onto findings above.

1. **Indirection that re-derives a known fact gets deleted.** OLM v0 ran a SAT
   solver, `github.com/go-air/gini`, for dependency resolution
   ([solver package](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/pkg/controller/registry/resolver/solver/doc.go)).
   OLM v1 dropped it. The successor project `deppy` is archived. Skiperator's
   generator registry (B6) is the same shape at much smaller scale: a runtime
   lookup for a compile-time fact.

2. **v1 applies plain rendered manifests through server-side apply**, with
   single ownership per object
   ([clusterobjectsets.md](https://github.com/operator-framework/operator-controller/blob/main/docs/draft/concepts/clusterobjectsets.md)).
   It moved *away* from an imperative install strategy, toward "render, then
   apply". Skiperator's generate-then-process split is already that shape. The
   hand-rolled apply in `crud.go` is the part that is not.

OLM v0 is now formally in maintenance mode. Its README lists the rules:
"1. We will accept no new features."
([README](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/README.md)).
Complexity that nobody wants to maintain is how a project reaches that state.

### nais/naiserator, the sibling project

Skiperator and naiserator share ancestry. Skiperator depends on
`nais/liberator` and cites naiserator in
[helperfunctions.go:155](../pkg/util/helperfunctions.go:155).

The audit found that **naiserator has the same two designs that this document
flags as problems**. That is useful. It means those designs are inherited, not
reasoned.

| Naiserator does                                                                                                                                                                                                                           | Same as Skiperator?                | Verdict                                                                    |
|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------|----------------------------------------------------------------------------|
| Generators are plain functions `(source, ast, cfg) error`, called in one explicit ordered list in [pkg/generators/application.go](https://github.com/nais/naiserator/blob/master/pkg/generators/application.go). No registry. No `init()` | **No.** Skiperator uses a registry | Copy naiserator. This is B6                                                |
| Golden-file tests over pure generators: **140 YAML files** in [pkg/resourcecreator/testdata](https://github.com/nais/naiserator/tree/master/pkg/resourcecreator/testdata), driven by `goldenfile.Run`                                     | No                                 | Copy naiserator. This is B15                                               |
| Hand-rolled Get then Create-or-Update in [updater/updater.go](https://github.com/nais/naiserator/blob/master/updater/updater.go). No field manager, no `controllerutil`, no apply                                                         | Yes                                | **Do not copy.** Same shape as `crud.go`                                   |
| Finds stale children by listing a fixed kind list with an `app=` label, then diffing with `reflect.TypeOf`, guarded by a listable check and a `recover()` per List                                                                        | **Yes**, almost exactly            | **Do not copy.** See B16                                                   |
| Status: seven scalar fields plus `Conditions *[]metav1.Condition`, a pointer to a slice                                                                                                                                                   | Partly                             | Do not copy. The pointer-to-slice breaks the standard `listType=map` shape |
| A per-application background goroutine polls the Deployment until replicas match, in [pkg/synchronizer/monitoring.go](https://github.com/nais/naiserator/blob/master/pkg/synchronizer/monitoring.go)                                      | No                                 | Do not copy. `Watches` plus requeue is the framework answer                |

Take naiserator's **generator shape and its golden-file tests**. Take nothing
else from it. For apply, pruning, cache, status, and validation placement,
follow cert-manager, cluster-api, and Flux.

### How the CNCF-scale projects build and apply

**Building desired objects: plain functions, everywhere.**
prometheus-operator uses package-private `makeStatefulSet(...)`,
`makeStatefulSetService(...)` per component. cert-manager uses a small
`apply.go` per controller. Neither has a registry or an `init()` dispatch. This
is the third independent confirmation of B6.

**Applying: split, and instructive.**

cert-manager **migrated to server-side apply deliberately**, with a written
design document,
[design/20220118.server-side-apply.md](https://github.com/cert-manager/cert-manager/blob/master/design/20220118.server-side-apply.md),
shipped behind the `ServerSideApply` feature gate in v1.8. Stated reasons:
resource-version conflicts between controllers, and the inability to express
field *deletion* without keeping state in annotations. Each controller carries
its own `fieldManager string`.

prometheus-operator **did not**. It still runs a hand-rolled
`CreateOrUpdateService` with `retry.RetryOnConflict` in
[pkg/k8s/network.go](https://github.com/prometheus-operator/prometheus-operator/blob/main/pkg/k8s/network.go).
Apply appears only under `test/framework/`.

**The most useful detail for Skiperator is how prometheus-operator hashes.**
It keeps a spec hash, so the pattern is legitimate at scale. But
`createSSetInputHash(p, config, ruleConfigMapNames, tlsAssets, existingStatefulSet.Spec)`
hashes **named, declared inputs that the operator controls**, stored under the
`prometheus-operator-input-hash` annotation, with dedicated tests asserting
hash stability across semantically equal inputs.

Skiperator instead reflects over whatever field happens to be called `Spec`
([crud.go:133](../pkg/resourceprocessor/crud.go:133)) and silently returns an
empty hash when there is none. **No upstream project hashes by reflection over
the whole object.** If the hash stays, narrow it to declared inputs.

### What server-side apply actually costs

B2 starts with `CreateOrUpdate` and B17 moves to apply later. The upstream
record is what sets that order, and it is what B17's preconditions come from.

| Project      | Problem hit                                                                                                                                        | Link                                                                                                                                         |
|--------------|----------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| cert-manager | Objects written by the old client-side path keep an `...-update` manager entry the new `...-apply` manager does not own, so fields linger silently | [issue 6969](https://github.com/cert-manager/cert-manager/issues/6969)                                                                       |
| cert-manager | The client-go fake client cannot do an Apply patch, so tests must move to envtest                                                                  | design doc, "testing" section                                                                                                                |
| cert-manager | Still fixing apply regressions years later, in v1.21.1                                                                                             | [issue 9084](https://github.com/cert-manager/cert-manager/issues/9084)                                                                       |
| cluster-api  | CRD `listType` and `listMapKey` markers must be right or apply conflicts on lists. They added a linter to enforce it                               | [PR 13340](https://github.com/kubernetes-sigs/cluster-api/pull/13340), [PR 12470](https://github.com/kubernetes-sigs/cluster-api/pull/12470) |
| cluster-api  | Apply is expensive enough that they **cache no-op applies**, and reverted some paths to a regular patch for CPU reasons                            | [internal/util/ssa/cache.go](https://github.com/kubernetes-sigs/cluster-api/blob/main/internal/util/ssa/cache.go)                            |
| cluster-api  | Apply broke when old API versions were removed                                                                                                     | [issue 10051](https://github.com/kubernetes-sigs/cluster-api/issues/10051)                                                                   |
| Flux         | A managed-fields regression forced a minimum Kubernetes version per release line                                                                   | [flux2 issue 1889](https://github.com/fluxcd/flux2/issues/1889)                                                                              |

cert-manager gated it. B17 requires the same gate.

### Cache configuration, with real code

cluster-api's manager options, in
[core/setup/setup.go](https://github.com/kubernetes-sigs/cluster-api/blob/main/core/setup/setup.go),
are the model for B3. Two parts matter.

```go
// cache: only Secrets carrying the cluster-name label, with fields dropped
ByObject: map[client.Object]cache.ByObject{
    &corev1.Secret{}: {
        Label: clusterSecretCacheSelector,
        Transform: func(in any) (any, error) {
            if s, ok := in.(*corev1.Secret); ok {
                s.SetManagedFields(nil)
                if !strings.HasSuffix(s.Name, fmt.Sprintf("-%s", secret.Kubeconfig)) {
                    s.Data = nil
                }
            }
            return in, nil
        },
    },
},
```

```go
// client: cache unstructured reads, but never cache the heavy types
Cache: &client.CacheOptions{
    DisableFor:   []client.Object{&corev1.ConfigMap{}, &corev1.Secret{}},
    Unstructured: true,
},
```

Note the second block. `Unstructured: true` duplicates every informer and
doubles memory, so cluster-api pairs it with `DisableFor` on exactly the kinds
that make that expensive. Skiperator does not need either half. Typed lists,
in B1, create no second informer set at all.

cluster-api also documents each exclusion with a one-line reason. Copy that
habit.

### Nobody lists N kinds per reconcile

This is the finding that reframes B1. The research pass found **no
CNCF-graduated operator that label-lists many kinds on the reconcile path.**
The nearest match is naiserator, which shares Skiperator's ancestry.

Upstream uses one of two patterns instead, and B16 below applies them.

1. **Owner-reference garbage collection only.** prometheus-operator carries the
   same comment in four operator files: "Dependent resources are cleaned up by
   K8s via OwnerReferences". It works because children are named
   deterministically from the parent, so the set of possible children is known
   without discovery.
2. **An inventory of what was applied.** Flux stores
   [ResourceInventory](https://github.com/fluxcd/kustomize-controller/blob/main/api/v1/inventory_types.go)
   in `status.inventory`. cluster-api uses a sidecar object,
   [ClusterResourceSetBinding](https://github.com/kubernetes-sigs/cluster-api/blob/main/api/addons/v1beta2/clusterresourcesetbinding_types.go).
   Pruning becomes: diff the previous inventory against this pass. One object
   read, zero cross-kind Lists. Flux needs it because it applies arbitrary user
   YAML whose object set cannot be known in advance.

For "which children does this parent own" without listing, the upstream tool is
a field index on owner references, through `mgr.GetFieldIndexer().IndexField`,
queried with `client.MatchingFields`. That is a cached in-memory lookup.

---

## Measurement plan

Prove each claim before and after. The metrics endpoint already exists, at
[main.go:174](../cmd/skiperator/main.go:174).

| Item | Metric                                                                                                 | What proves it                                                                                                                               |
|------|--------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| B1   | `rest_client_requests_total` divided by `controller_runtime_reconcile_total{controller="application"}` | Calls per reconcile. About 48 today, about 4 after. Cache reads never appear here, so this ratio is the cleanest single number in the audit. |
| B4   | `controller_runtime_reconcile_total{result="requeue_after"}` against `result="success"`                | If the requeue bucket dominates, the timer drives the fleet, not events.                                                                     |
| E5   | `workqueue_depth` and `workqueue_queue_duration_seconds` against `controller_runtime_active_workers`   | Workers pinned at 1 with a growing depth means concurrency is the limit.                                                                     |
| B11  | `controller_runtime_reconcile_time_seconds` p99 against p50                                            | A long tail with flat client latency points at the registry call.                                                                            |
| B3   | `go tool pprof -inuse_space` on `:8281` (set `enableProfiling: true`)                                  | Look for `cache.(*threadSafeStore).Add` held by the Secret and ConfigMap reflectors.                                                         |

Cheapest regression guard: wrap the client in a
`sigs.k8s.io/controller-runtime/pkg/client/interceptor` that counts List, Get,
and Update calls. Run one Application reconcile under envtest and assert the
count. That turns the headline number into a test.

Benchmarks worth writing, with `-benchmem`:

- `BenchmarkIsObjectIdentical` over a real Deployment. Measures the reflection,
  the JSON marshal, and the SHA-256, twice per object.
- `BenchmarkDeploymentPredicate` with a real `event.UpdateEvent`. Measures two
  `DeepCopyObject` calls and two `hashstructure.Hash` reflection walks per
  Deployment event, at [predicates.go:32](../internal/controllers/common/predicates.go:32).
- `BenchmarkGetObjectDiff` on a populated `ApplicationSpec`, against
  `equality.Semantic.DeepEqual`.

---

## Suggested order

1. **B15** — golden-file tests over the generators. Do this **first**. It is
   the safety net that makes everything below cheap to attempt, and it is the
   one thing naiserator already proves works.
2. **E1, E2, E3, E4, E6** — five small perf fixes, no design decision needed.
3. **B1** — typed lists. The largest immediate win. Measure with the
   ratio in the table above.
4. **B3** — cache selectors and a transform. The largest memory win. About 20
   lines in `main.go`, copied from cluster-api's shape.
5. **B5** — one status write, and stop the timestamp churn. Removes a
   dependency as a side effect.
6. **B2** — `CreateOrUpdate` plus `equality.Semantic.DeepEqual`. Fixes the two
   write-amplification defects.
7. **B4** — event-driven readiness. Then the 10-second timer can go.
8. **E7 to E20** — the deletions. Cheap, and they shrink the diff for everything
   after.
9. **B6 to B14** — the structural work. B6 is the highest value of these, and
   three independent upstream projects agree with it.
10. **B16** — the inventory. Direction of travel, not a next step. It makes B1
    mostly unnecessary, but it carries the risk of orphaning or wrongly
    deleting resources, so it needs B15 in place first.
11. **B17** — server-side apply, behind a feature gate. The destination for the
    write path. Its four preconditions include B15 and B2, so it cannot come
    earlier.
