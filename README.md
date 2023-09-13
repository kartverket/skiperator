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

See also [the documentation directory](https://github.com/kartverket/skiperator/tree/main/doc).

```yaml
apiVersion: skiperator.kartverket.no/v1alpha1
kind: Application
metadata:
  # *Required*: The name of the Application and the created resources
  name: teamname-frontend
  # *Required*: The namespace the Application and the created resources will be in
  namespace: yournamespace
spec:
  # *Required*: A deployment will be created and this image will be run
  image: "kartverket/example"
  # The port the deployment exposes
  port: 8080
  # An optional priority. Supported values are 'low', 'medium' and 'high'.
  # The default value is 'medium'.
  #
  # Most workloads should not have to specify this field. If you think you
  # do, please consult with SKIP beforehand.
  priority: medium
  # An optional list of extra port to expose on a pod level basis,
  # for example so Instana or other APM tools can reach it
  additionalPorts:
    - name: metrics-port
      port: 8181
      protocol: TCP
    - name: another-port
      port: 8282
      protocol: TCP
  # Override the command set in the Dockerfile. Usually only used when debugging
  # or running third-party containers where you don't have control over the Dockerfile
  command:
    - node
    - ./server.js
  # Any external hostnames that route to this application. Using a skip.statkart.no-address
  # will make the application reachable for kartverket-clients (internal), other addresses
  # make the app reachable on the internet. Note that other addresses than skip.statkart.no
  # (also known as pretty hostnames) requires additional DNS setup.
  # The below hostnames will also have TLS certificates issued and be reachable on both
  # HTTP and HTTPS.
  # Ingress must be lowercase, contain no spaces, be a non-empty string, and have a hostname/domain separated by a period
  ingresses:
    - testapp.dev.skip.statkart.no
  # The number of replicas can either be specified as a static number as follows:
  replicas: 2
  # Or by specifying a range between min and max to enable HorizontalPodAutoscaling
  # The default value for replicas is "min: 2, max: 5, targetCpuUtilization: 80"
  # Using autoscaling is the recommended configuration for replicas
  replicas:
    # Minimum number of replicas when load is low
    min: 2
    # Maximum number of replicas the deployment is allowed to scale to
    max: 5
    # When the average CPU utilization crosses this threshold another replica is started
    targetCpuUtilization: 80
  # To interact with GCP.
  gcp:
    # For authentication with GCP, to use services like Secret Manager and/or Pub/Sub we need
    # to set the GCP service account to identify as. To allow this, we need an IAM role binding in
    # GCP adding the role Workload Identity User for the Kubernetes SA on the GCP SA.
    # Documentation on how this is done can be found here:
    # https://kartverket.atlassian.net/wiki/spaces/SKIPDOK/pages/422346824/Autentisering+mot+GCP+som+Kubernetes+SA
    auth:
      serviceAccount: some-serviceaccount@some-project-id.iam.gserviceaccount.com
  # Environment variables that will be set inside the Deployment's pod
  env:
    # Alternative 1: Keys and values provided directly
    - name: ENV
      value: PRODUCTION
    # Alternative 2: Keys with dynamic values. valueFrom supports configMaps, secrets
    # and fieldRef, which selects a single key from the deployment object at runtime
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
  # Environment variables mounted from files. When specified all the keys of the
  # resource will be assigned as environment variables. Supports both configmaps
  # and secrets. For mounting as files see filesFrom
  envFrom:
    - configMap: some-configmap
    - secret: some-secret
  # Mounting volumes into the Deployment are done using the filesFrom argument
  # filesFrom supports configmaps, secrets and pvcs. The Application resource
  # assumes these have already been created by you
  # NB. Out-of-the-box, skiperator provides a writable 'emptyDir'-volume at '/tmp'
  filesFrom:
    - emptyDir: temp-dir              # <-- provided
      mountPath: /tmp
    - configMap: some-configmap
      mountPath: /var/run/configmap
    - secret: some-secret
      mountPath: /var/run/secret
    - persistentVolumeClaim: some-pvc
      mountPath: /var/run/volume
  # Defines an alternative strategy for the Kubernetes deployment is useful for when
  # the default which is rolling deployments are not usable. Setting type to
  # Recreate will take down all the pods before starting new pods, whereas the
  # default of RollingUpdate will try to start the new pods before taking down the
  # old ones
  strategy:
    # Valid values: RollingUpdate, Recreate. Default RollingUpdate
    type: RollingUpdate
  # Liveness probes define a resource that returns 200 OK when the app is running
  # as intended. Returning a non-200 code will make kubernetes restart the app.
  # Liveness is optional, but when provided path and port is required
  liveness:
    # The path to access on the HTTP server
    path: /healthz
    # Number or name of the port to access on the container
    port: 8080
    # Minimum consecutive failures for the probe to be considered failed after
    # having succeeded. Defaults to 3. Minimum value is 1
    failureThreshold: 3
    # Number of seconds after which the probe times out. Defaults to 1 second.
    # Minimum value is 1
    timeout: 1
    # Delay sending the first probe by X seconds. Can be useful for applications that
    # are slow to start.
    initialDelay: 0
  # Readiness probes define a resource that returns 200 OK when the app is running
  # as intended. Kubernetes will wait until the resource returns 200 OK before
  # marking the pod as Running and progressing with the deployment strategy.
  # Readiness is optional, but when provided path and port is required
  readiness:
    # Readiness has the same options as liveness
    path: ..
    port: 8181
  # Kubernetes uses startup probes to know when a container application has started.
  # If such a probe is configured, it disables liveness and readiness checks until it
  # succeeds, making sure those probes don't interfere with the application startup.
  # This can be used to adopt liveness checks on slow starting containers, avoiding them
  # getting killed by Kubernetes before they are up and running.
  # Startup is optional, but when provided path and port is required
  startup:
    # Startup has the same options as liveness
    path: ..
    port: 8282
  # Labels can be used if you want every resource created by your application to
  # have the same labels, including your application. This could for example be useful for
  # metrics, where a certain label and the corresponding resources liveliness can be combined.
  # Any amount of labels can be added as wanted, and they will all cascade down to all resources.
  labels:
    someLabel: some-label
  # Resource Labels can be used if you want to add a label to a specific resources created by
  # the application. One such label could for example be set on a Deployment, such that
  # the deployment avoids certain rules from Gatekeeper, or similar. Any amount of labels may be added per resourceLabel item.
  resourceLabels:
    Deployment:
      labelKey: A value for the label
    Service:
      labelKeyOne: A value for the one label
      labelKeyTwo: A value for the two label
  # Optional settings for how Prometheus compatible metrics should be scraped.
  prometheus:
    port: 8181         # *Required* Pod port number or name
    path: "/metrics"   # *Optional* Defaults to /metrics
  # Settings for overriding the default deny of actuator endpoints. AllowAll set to true will allow any
  # endpoint to be exposed. Use AllowList to only allow specific endpoints.
  #
  # Please be aware that HTTP endpoints, such as actuator, may expose information about your application which you do not want to expose.
  # Before allow listing HTTP endpoints, make note of what these endpoints will expose, especially if your application is served via an external ingress.
  authorizationSettings:
    # Default false
    allowAll: false
    # Default empty
    # Endpoints must be prefixed with /
    # Note that endpoints are matched specifically on the input, so if you for example allow /actuator/health, you will *not* allow /actuator/health/
    allowList:
      - "/actuator/health"
      - "/actuator/info"
  # Resource limits to apply to the deployment. It's common to set these to
  # prevent the app from swelling in resource usage and consuming all the
  # resources of other apps on the cluster.
  resources:
    # Limits set the maximum the app is allowed to use. Exceeting this will
    # make kubernetes kill the app and restart it.
    limits:
      # A value in millicpus (m)
      # NOTE: This is not recommended to set.
      # See: https://home.robusta.dev/blog/stop-using-cpu-limits
      cpu: 1000m
      # Number of bytes of RAM
      memory: 1G
    # Requests set the initial allocation that is done for the app and will
    # thus be available to the app on startup. More is allocated on demand
    # until the limit is reached
    requests:
      # A value in millicpus (m)
      cpu: 25m
      # Number of bytes of RAM
      memory: 250M
  # Whether to enable automatic Pod Disruption Budget creation for this application. Defaults to true and may be omitted.
  enablePDB: true
  # Zero trust dictates that only applications with a reason for being able
  # to access another resource should be able to reach it. This is set up by
  # default by denying all ingress and egress traffic from the pods in the
  # Deployment. The accessPolicy field is an allowlist of other applications
  # that are allowed to talk with this resource and which resources this app
  # can talk to
  accessPolicy:
    # inbound specifies the ingress rules. Which apps on the cluster can talk
    # to this app?
    inbound:
      # The rules list specifies a list of applications. When no namespace is
      # specified it refers to an app in the current namespace. For apps in
      # other namespaces namespace is required
      rules:
        - application: other-app
        - application: third-app
          namespace: other-namespace
    # outbound specifies egress rules. Which apps on the cluster and the
    # internet are the Application allowed to send requests to?
    outbound:
      # The rules list specifies a list of applications that are reachable on
      # the cluster. See rules in inbound for syntax. Note that the application
      # you're trying to reach also must specify that they accept communication
      # from this app in their ingress rules
      rules:
        - application: other-app
      # external specifies which applications on the internet the application
      # can reach. Only host is required unless it is on another port than HTTPS
      # on port 443. If other ports or protocols are required then `ports` must
      # be specified as well
      external:
        # The allowed hostname. Note that this does not include subdomains
        - host: nrk.no
          # Non-HTTP requests (i.e. using the TCP protocol) need to use IP in
          # addition to hostname
        - host: smtp.mailgrid.com
          # IP address. Only required for TCP requests.
          # Note: Hostname must always be defined even if IP is set statically
          ip: "123.123.123.123"
          # The ports to allow for the above hostname. When not specified HTTP and
          # HTTPS on port 80 and 443 respectively are put into the allowlist
          ports:
            # Name is required and is an arbitrary name. Must be unique within
            # this array
            - name: smtp
              # Supported protocols are: TCP, HTTP, HTTPS
              protocol: TCP
              port: 587
```

## SKIPJob reference

Below you will find a list of all accepted input parameters to the `SKIPJob`
custom resource. Only types are shown here. The fields are documented in the API, see [skipjob_types.go](api/v1alpha1/skipjob_types.go)

```yaml
apiVersion: skiperator.kartverket.no/v1alpha1
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
    ttlSecondsAfterFinished: 10
    
  container:
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

## Developing

See [CONTRIBUTING.md](CONTRIBUTING.md) for information on how to develop the
Skiperator.
