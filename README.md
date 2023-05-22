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
  name: teamname-frontend
  namespace: yournamespace
spec:
  image: "kartverket/example"
  
  port: 8080
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
  command:
    - node
    - ./server.js
  ingresses:
    - testapp.dev.skip.statkart.no
  replicas:
    min: 3
    max: 5
    targetCpuUtilization: 80
  gcp:
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
    # Valid values: RollingUpdate, Recreate. Default RollingUpdate
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

## Developing

See [CONTRIBUTING.md](CONTRIBUTING.md) for information on how to develop the
Skiperator.
