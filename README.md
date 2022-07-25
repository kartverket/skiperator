# Skiperator

Skiperator is an operator intended to make the setup of applications simple from
the users' point of view. When using the operator an application developer can
set up all associated resources for an optimal deployment using a simple custom
resource called `Application`.

## Application reference

Below you will find a list of all accepted input parameters to the `Application`
custom resource.

```yaml
apiVersion: skiperator.kartverket.no/v1alpha1
kind: Application
metadata:
  # *Required*: The name of the Application and the created resources
  name: application-frontend
spec:
  # *Required*: A deployment will be created and this image will be run
  image: "kartverket/example:latest"
  # The port the deployment exposes
  port: 8080
  # Override the command set in the Dockerfile. Usually only used when debugging
  # or running third-party containers where you don't have control over the Dockerfile
  command:
    - node
    - ./server.js
  # Any external hostnames that route to this application
  ingresses:
  - testapp.dev.skip.statkart.no
  # Configuration used to automatically scale the deployment based on load
  replicas:
    # Minimum number of replicas when load is low
    min: 3
    # Maximum number of replicas the deployment is allowed to scale to
    max: 5
    # When the average CPU utilization crosses this threshold another replica is started
    targetCpuUtilization: 80
  # Environment variables that will be set inside the Deployment's pod
  env:
  # Alternative 1: Keys and values provided directly
  - name: ENV
    value: PRODUCTION
  # Alternative 2: Keys with dynamic values. valueFrom supports configMaps, secrets
  # and fieldRef, which selects a single key from the deployment object at runtime
  - name: USERNAME
    valueFrom:
      configMapRef:
        name: some-configmap
        key: username
  - name: PASSWORD
    valueFrom:
      secretRef:
        name: some-secret
        key: password
  # Environment variables mounted from files. When specified all the keys of the
  # resource will be assigned as environment variables. Supports both configmaps
  # and secrets. For mounting as files see filesFrom
  envFrom:
  - configmap: some-configmap
  - secret: some-secret
  # Mounting volumes into the Deployment are done using the filesFrom argument
  # filesFrom supports configmaps, secrets and pvcs. The Application resource
  # assumes these have already been created by you
  filesFrom:
  - emptyDir: temp-dir
    mountPath: /tmp
  - configmap: some-configmap
    mountPath: /var/run/configmap
  - secret: some-secret
    mountPath: /var/run/secret
  - persistentVolumeClaim: some-pvc
    mountPath: /var/run/volume
  # Defines an alternative strategy for the Kubernetes deployment is useful for when
  # the deafult which is rolling deployments are not usable. Setting type to
  # Recreate will take down all the pods before starting new pods, whereas the
  # default of RollingUpdate will try to start the new pods before taking down the
  # old ones
  strategy:
    # Valid values: RollingUpdate, Recreate. Default RollingUpdate
    type: RollingUpdate
  # Liveness probes define a resource that returns 200 OK when the app is running
  # as intended. Returning a non-200 code will make kubernetes restart the app.
  # Liveness is optional, but when provided path and port is requred
  liveness:
    # The path to access on the HTTP server
    path: /healthz
    # Number of the port to access on the container
    port: 8080
    # Minimum consecutive failures for the probe to be considered failed after
    # having succeeded. Defaults to 3. Minimum value is 1
    failureThreshold: 3
    # How often (in seconds) to perform the probe. Default to 10 seconds.
    # Minimum value is 1
    periodSeconds: 10
    # Number of seconds after which the probe times out. Defaults to 1 second.
    # Minimum value is 1
    timeout: 1
  # Readiness probes define a resource that returns 200 OK when the app is running
  # as intended. Kubernetes will wait until the resource returns 200 OK before
  # marking the pod as Running and progressing with the deployment strategy.
  # Readiness is optional, but when provided path and port is requred
  readiness:
    # Readiness has the same options as liveness
    path: ..
  # Kubernetes uses startup probes to know when a container application has started.
  # If such a probe is configured, it disables liveness and readiness checks until it
  # succeeds, making sure those probes don't interfere with the application startup.
  # This can be used to adopt liveness checks on slow starting containers, avoiding them
  # getting killed by Kubernetes before they are up and running.
  # Startup is optional, but when provided path and port is requred
  startup:
    # Startup has the same options as liveness
    path: ..
  # Resource limits to apply to the deployment. It's common to set these to
  # prevent the app from swelling in resource usage and consuming all the
  # resources of other apps on the cluster.
  resources:
    # Limits set the maximum the app is allowed to use. Exceeting this will
    # make kubernetes kill the app and restart it.
    limits:
      # A value in millicpus (m)
      cpu: 1000m
      # Number of bytes of RAM
      memory: 1G
    # Requests set the initial allocation that is done for the app and will
    # thus be available to the app on startup. More is allocated on demand
    # until the limit is reached
    requests:
      # A value in millicpus (m)
      cpu: 500m
      # Number of bytes of RAM
      memory: 500M
  # Zero trust dictates that only applications with a reason for being able
  # to access another resource should be able to reach it. This is set up by
  # default by denying all ingress and egress traffic from the pods in the
  # Deployment. The accessPolicy field is an allowlist of other applications
  # that are allowed to talk with this resource and whic resources this app
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
    # internet is the Application allowed to send requests to?
    outbound:
      # The rules list specifies a list of applications that are reachable on
      # the cluster. See rules in inbound for syntax. Note that the application
      # you're trying to reach also must specify that they accept communication
      # from this app in their ingress rules
      rules:
      - application: other-app
      # external specifies which applications on the internet the application
      # can reach. Only host is required unless it is on another port than HTTP
      # on port 80 and HTTPS on port 443. If other ports or protocols are
      # required then port must be specified as well
      external:
        # The allowed hostname. Note that this does not unclude subdomains
      - host: nrk.no
      - host: smtp.mailgrid.com
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

## Developing

See [CONTRIBUTING.md](CONTRIBUTING.md) for information on how to develop the
Skiperator.
