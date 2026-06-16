# Gateway API migration flow

`gwapi` owns the routing migration state machine. Reconcilers call
`EvaluateRoutingState` before resource generation, then `UpdateRoutingStatus`
after generated resources have been processed.

```mermaid
stateDiagram-v2
    [*] --> LegacyOnly: routingProvider != standard

    [*] --> StandardOnly: standard ready + no legacy resources
    [*] --> CutoverReadyPruneLegacy: standard ready + legacy resources exist
    [*] --> GreenfieldPending: standard not ready + no legacy resources
    [*] --> MigratingWithFallback: standard not ready + legacy resources exist

    LegacyOnly: Generate legacy routing
    LegacyOnly: Ready=True

    GreenfieldPending: Generate Gateway API only
    GreenfieldPending: Ready=False
    GreenfieldPending: No migration event

    MigratingWithFallback: Generate Gateway API + legacy fallback
    MigratingWithFallback: Set MigrationStartedAt
    MigratingWithFallback: Emit GatewayAPIMigrationStarted

    MigratingWithFallback --> MigrationStalled: MigrationStartedAt older than 10m
    MigrationStalled: Keep legacy fallback
    MigrationStalled: Ready=False
    MigrationStalled: Emit GatewayAPIMigrationStalled once

    MigratingWithFallback --> CutoverReadyPruneLegacy: standard becomes ready
    MigrationStalled --> CutoverReadyPruneLegacy: standard becomes ready

    CutoverReadyPruneLegacy: Stop generating legacy routing
    CutoverReadyPruneLegacy: Prune legacy resources
    CutoverReadyPruneLegacy: Clear MigrationStartedAt
    CutoverReadyPruneLegacy: Emit GatewayAPIMigrationFinished

    CutoverReadyPruneLegacy --> StandardOnly: next reconcile + legacy resources gone

    StandardOnly: Generate Gateway API only
    StandardOnly: Ready=True
```

Legend:

- `standard ready` means ListenerSets, HTTPRoutes, Certificates, and TLS Secrets
  are accepted, programmed, and ready.
- `legacy resources exist` means previous Istio Gateway or VirtualService
  resources are still present.
- Greenfield standard routing never creates legacy fallback.
