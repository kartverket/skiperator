# Skiperator

Skiperator is an operator intended to make the setup of applications simple from
the users' point of view. When using the operator an application developer can
set up all associated resources for an optimal deployment using a simple custom
resource called `Application`.

## Prerequisites

- The Dockerfile must build an image where the user ID is set to `150` as this UID
  is hard coded in Skiperator

## Application reference

Below you will find a list of all accepted input parameters to the `Application`
custom resource.

To see explanations and requirements for all inputs, see the documentation under [the API documentation](https://skip.kartverket.no/docs/applikasjon-utrulling/skiperator/api-docs#application).

```yaml
apiVersion: skiperator.kartverket.no/v1alpha1
kind: Application
metadata:
  name: teamname-frontend
  namespace: yournamespace
spec:
  # Required, everything beyond image and port is optional
  image: "kartverket/example"
  port: 8080
  
  priority: medium
  
  additionalPorts:
    - name: metrics-port
      port: 8181
      protocol: TCP
    - name: another-port
      port: 8282
      protocol: TCP
      
  command:
    - node
    - ./server.js
     
  ingresses:
    - testapp.dev.skip.statkart.no
    
  replicas: 2
  # or
  replicas:
    min: 2
    max: 5
    targetCpuUtilization: 80
    targetMemoryUtilization: 80
    
  gcp:
    auth:
      serviceAccount: some-serviceaccount@some-project-id.iam.gserviceaccount.com
      
  env:
    - name: ENV
      value: PRODUCTION
    - name: USERNAME
      valueFrom:
        configMapKeyRef:
          name: some-configmap
          key: username
    - name: PASSWORD
      valueFrom:
        secretKeyRef:
          name: some-secret
          key: password
          
  envFrom:
    - configMap: some-configmap
    - secret: some-secret
  filesFrom:
    - emptyDir: temp-dir
      mountPath: /tmp
    - configMap: some-configmap
      mountPath: /var/run/configmap
    - secret: some-secret
      mountPath: /var/run/secret
    - persistentVolumeClaim: some-pvc
      mountPath: /var/run/volume
  
  strategy:
    type: RollingUpdate
  
  liveness:
    path: "/"
    port: 8080
    failureThreshold: 3
    timeout: 1
    initialDelay: 0
  readiness:
    # Readiness has the same options as liveness
    path: ..
  startup:
    # Startup has the same options as liveness
    path: ..
    
  labels:
    someLabel: some-label
    
  resourceLabels:
    Deployment:
      labelKey: A value for the label
    Service:
      labelKeyOne: A value for the one label
      labelKeyTwo: A value for the two label
      
  prometheus:
    port: 8181
    path: "/metrics"
  authorizationSettings:
    allowAll: false
    allowList:
      - "/actuator/health"
      - "/actuator/info"

  resources:
    limits:
      cpu: 1000m # Avoid using this
      memory: 1G
    requests:
      cpu: 25m
      memory: 250M
  
  enablePDB: true
  # Note: If replicas is set to 0, a PodDisruptionBudget will not be created, regardless of the value of enablePDB.
  
  accessPolicy:
    inbound:
      # The rules list specifies a list of applications. When no namespace is
      # specified it refers to an app in the current namespace. For apps in
      # other namespaces, namespace is required. Alternately you can define
      # namespacesByLabel as a value-map of namespace labels. If both
      # namespace and namespacesByLabel are defined for an application,
      # namespacesByLabel is ignored
      rules:
        - application: other-app
        - application: third-app
          namespace: other-namespace
        - application: fourth-app
          namespacesByLabel:
            somelabel: somevalue
            anotherlabel: anothervalue
    # outbound specifies egress rules. Which apps on the cluster and the
    # internet are the Application allowed to send requests to? Alternately
    # you can define namespacesByLabel as a value-map of namespace labels.
    # If both namespace and namespacesByLabel are defined for an application,
    # namespacesByLabel is ignored
    outbound:
      rules:
        - application: some-app
          namespacesByLabel:
            somelabel: somevalue
        - application: other-app
      external:
        - host: nrk.no
        - host: smtp.mailgrid.com
          ip: "123.123.123.123"
          ports:
            - name: smtp
              protocol: TCP
              port: 587
  # podSettings are used to apply specific settings to the Pod Template used by Skiperator to create Deployments.
  podSettings:
    annotations:
      some-annotation: some-value
    terminationGracePeriodSeconds: 30
    disablePodSpreadTopologyConstraints: false
  # istioSettings are used to configure istio specific resources. Currently, adjusting sampling interval for tracing is 
  # the only supported option.
  istioSettings:
    telemetry:
      tracing:
        - randomSamplingPercentage: 10
      

```

## SKIPJob reference

Below you will find a list of all accepted input parameters to the `SKIPJob`
custom resource. Use `v1beta1` for new SKIPJobs. Only types are shown here. The fields are documented in the API, see [the API](https://skip.kartverket.no/docs/applikasjon-utrulling/skiperator/api-docs#skipjob-1)

```yaml
apiVersion: skiperator.kartverket.no/v1beta1
kind: SKIPJob
metadata:
  namespace: sample
  name: sample-job
spec:
  cron:
    schedule: "* * * * *"
    suspend: false 
    startingDeadlineSeconds: 10
  
  job: 
    activeDeadlineSeconds: 10
    backoffLimit: 10
    suspend: false
    ttlSecondsAfterFinished: 
    
  prometheus:
    path: /metrics
    port: 8080  
  
  # Pod
  image: ""
  command:
    - ""
  resources:
    requests:
      cpu: 10m
      memory: 128Mi
    limits:
      memory: 256Mi

  # Networking
  accessPolicy:
    inbound:
      rules:
        - application: ""
          namespace: ""
    outbound:
      external:
        - host: ""
          ip: ""
          ports:
            - name: ""
              port: 10
              protocol: ""
  additionalPorts:
    - name: ""
      port: 10
      protocol: ""

  # Volumes / environment
  env:
    - name: ""
      value: ""
  envFrom:
    - configMap: ""
    - secret: ""
  filesFrom:
    - mountPath: ""
      # + one of:
      secret: ""
      configMap: ""
      emptyDir: ""
      persistentVolumeClaim: ""

  gcp:
    auth:
      serviceAccount: ""

  # Probes
  startup:
    path: ""
    port: 0
    failureThreshold: 0
    initialDelay: 0
    period: 0
    successThreshold: 0
    timeout: 0
  # Same as startup
  liveness:
    ...
  readiness:
    ...

  # Miscellaneous
  priority: ""
  restartPolicy: ""
```

## Routing reference

Routing is a separate custom resource that can be used to set up path-based routing for Skiperator applications. Use this
instead of the `ingresses` field in the `Application` custom resource. Routes target Applications in the same namespace
as the Routing resource.

Below you will find a list of all accepted input parameters to the `Routing`
custom resource. Only types are shown here. The fields are documented in the API, see [the API](https://skip.kartverket.no/docs/applikasjon-utrulling/skiperator/api-docs#routing)

```yaml
apiVersion: skiperator.kartverket.no/v1alpha1
kind: Routing
metadata:
  name: sample-routing
  namespace: sample
spec:
  hostname: app.example.com
  routes:
  - pathPrefix: /api
    rewriteUri: true
    targetApp: backend-app
  - pathPrefix: /
    rewriteUri: false
    targetApp: frontend-app
```

### Routing fields and defaults

| Field | Required | Default | Description |
| --- | --- | --- | --- |
| `hostname` | Yes | — | Hostname to serve. Append `+secret-name` to use a custom TLS Secret. |
| `routes` | Yes | — | Path-to-Application rules. Each route requires `pathPrefix` and `targetApp`. |
| `routes[].pathPrefix` | Yes | — | URL path prefix, beginning with `/`. Use `/` to match all paths. |
| `routes[].targetApp` | Yes | — | Name of an `Application` in the same namespace. |
| `routes[].port` | No | Application's primary `spec.port` | Service port on the target Application. Set this when routing to another port exposed by that Application. |
| `routes[].rewriteUri` | No | `false` | Rewrite the matched path prefix to `/`; details differ slightly by routing provider below. |
| `routingProvider` | No | `Legacy` | `Legacy` uses Istio Gateway and VirtualService. `Standard` uses Kubernetes Gateway API. |
| `ownership` | No | `Standalone` | `Standalone` serves a hostname exclusively. `Shared` lets multiple Standard Routing resources contribute paths to one hostname. |
| `redirectToHTTPS` | No | `true` | Redirect HTTP requests to HTTPS. For shared ownership, the redirect is shared at the hostname level. |

### Path matching and rewriting

`pathPrefix` is a path-prefix match, not an exact match. For example, `/api`
matches `/api` and `/api/v1`, but not `/apiv2`. A `/` route is a catch-all and
overlaps every other path on that hostname.

Matching precedence depends on the provider:

- With `Legacy` (the default), Istio evaluates routes in the order listed. Put
  specific prefixes before broader ones and the catch-all last. If `/` comes
  first, it matches every request, so later routes such as `/api` will never be
  reached.
- With `Standard`, Gateway API selects the most specific matching path prefix.
  Do not depend on YAML list order to resolve overlapping paths.

`rewriteUri: true` replaces the matched prefix with `/`. On Standard routing,
any unmatched suffix is retained by Gateway API's prefix rewrite. On Legacy
routing, Istio's URI rewrite replaces the matched request URI with `/`. Leave
it unset or false when the target application expects the original path.

If `port` is omitted, Skiperator uses the target Application's primary Service
port (`spec.port`). Set it explicitly to select another Service port exposed by
the Application.

### Routing providers and migration

`routingProvider` is optional and defaults to `Legacy`.

- `Legacy` creates Istio Gateway and VirtualService resources.
- `Standard` creates Kubernetes Gateway API ListenerSet and HTTPRoute
  resources.

To move an existing Routing to Gateway API, change
`spec.routingProvider` to `Standard`. During a migration, Skiperator keeps
existing Legacy routing active until the Standard routing resources are ready,
then removes the Legacy resources. A new Standard Routing has no Legacy
fallback. Check the `Ready`, `StandardRoutingReady`, and
`LegacyRoutingActive` conditions to see migration progress.

### Sharing a hostname

Use `ownership: Shared` when separate Routing resources, possibly in different
namespaces, need to contribute paths under the same hostname. Shared ownership
requires `routingProvider: Standard`. Each contributor still targets
Applications in its own namespace.

```yaml
apiVersion: skiperator.kartverket.no/v1alpha1
kind: Routing
metadata:
  name: team-a-routes
  namespace: team-a
spec:
  hostname: api.example.com
  routingProvider: Standard
  ownership: Shared
  routes:
    - pathPrefix: /team-a
      targetApp: api
```

Teams can create `Routing` resources in separate namespaces, using the same
hostname and distinct path prefixes to direct requests to their Applications.
Skiperator creates one shared set of Gateway API resources for that hostname
and keeps it while any `Routing` using that hostname remains. It removes those
resources only after the last such `Routing` is deleted.

Avoid overlapping prefixes between namespaces. For example, `/team` and
`/team/a` overlap. For Standard routing, Skiperator rejects a new path claim
that overlaps another accepted HTTPRoute. If paths overlap because multiple
routes were accepted concurrently or later changed, Gateway API determines
which route handles each request and Skiperator reports a `RoutePathConflict`
condition. A shared `/` catch-all therefore prevents other `Routing` resources from
claiming paths on that hostname.

`redirectToHTTPS` is also hostname-wide for shared ownership. If any `Routing`
using the hostname enables the redirect, it remains enabled while shared
hostname resources exist; setting it to false in another `Routing` does not
disable that redirect.

### TLS certificates

By default, Skiperator provisions a managed certificate for the hostname.
To use a team-provided certificate, append its Secret name to the hostname,
separated by `+`:

```yaml
spec:
  hostname: api.example.com+api-example-tls
```

The custom Secret must be provisioned in the `istio-gateways` namespace and
contain a certificate valid for the hostname. Skiperator does not create that
Secret. Standard routing uses a ReferenceGrant to allow a namespace-local
ListenerSet to read it; the Secret itself remains in `istio-gateways`.

Shared Routing cannot specify a custom certificate. All `Routing` resources
using the shared hostname use its managed certificate. A missing or unusable custom
certificate leaves Standard routing not ready; inspect the `Ready` and
`StandardRoutingReady` condition messages.

### Checking Routing status

```sh
kubectl get routing -n team-a
kubectl describe routing public-apps -n team-a
```

The `Ready` condition and summary describe reconciliation. Standard Routing
also reports `StandardRoutingReady`; migrations can report
`LegacyRoutingActive`. Shared Routing reports `SharedRoutingResources`, and
path overlaps are reported with `RoutePathConflict`.

## Developing

See [CONTRIBUTING.md](CONTRIBUTING.md) for information on how to develop the
Skiperator.
