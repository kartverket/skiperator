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

To see explanations and requirements for all inputs, see the documentation under [the API](https://pkg.go.dev/github.com/kartverket/skiperator@v1.0.0/api/v1alpha1).

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
```

## SKIPJob reference

Below you will find a list of all accepted input parameters to the `SKIPJob`
custom resource. Only types are shown here. The fields are documented in the API, see [the API](https://pkg.go.dev/github.com/kartverket/skiperator@v1.0.0/api/v1alpha1)

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
