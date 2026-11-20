<!--
MIT License

Copyright (c) 2022 NAV

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
-->

# Access Policy

## Access Policy

Access policies express which applications and services you are able to communicate with, both inbound and outbound. The default policy is to **deny all incoming and outgoing traffic** for your application, meaning you must be conscious of which services/application you consume, and who your consumers are.

> NOTE:
> 
> The Access policies only apply when communicating interally within the cluster with [service discovery](../clusters/service-discovery.md).
> Outbound requests to ingresses are regarded as external hosts, even if these ingresses exist in the same cluster.
> Analogously, inbound access policies are thus _not_ enforced for requests coming through exposed ingresses.

### Inbound rules

Inbound rules specifies what other applications _in the same cluster_ your application receives traffic from.

#### Receive requests from other app in the same namespace

For app `app-a` to be able to receive incoming requests from `app-b` in the same cluster and the same namespace, this specification is needed for `app-a`:

```yaml
apiVersion: "skiperator.kartverket.no/v1alpha1"
kind: "Application"
metadata:
  name: app-a
...
spec:
  ...
  accessPolicy:
    inbound:
      rules:
        - application: app-b
```

#### Receive requests from other app in the another namespace

For app `app-a` to be able to receive incoming requests from `app-b` in the same cluster but another namespace \(`othernamespace`\), this specification is needed for `app-a`:

```yaml
apiVersion: "skiperator.kartverket.no/v1alpha1"
kind: "Application"
metadata:
  name: app-a
...
spec:
  ...
  accessPolicy:
    inbound:
      rules:
        - application: app-b
          namespace: othernamespace
```

#### Receive traffic on more ports

Skiperator opens the application port for all inbound rules. If a rule has `ports`, Skiperator also opens these ports for this rule only. These ports are always TCP.

For example, Keycloak pods in a cluster send cache data to each other on the JGroups ports 7800 and 57800. For the pods of `keycloak` to reach each other on these ports, this specification is necessary:

```yaml
apiVersion: "skiperator.kartverket.no/v1alpha1"
kind: "Application"
metadata:
  name: keycloak
...
spec:
  ...
  additionalPorts:
    - name: tcp-jgroups
      port: 7800
      protocol: TCP
    - name: tcp-jgroups-fd
      port: 57800
      protocol: TCP
  accessPolicy:
    inbound:
      rules:
        - application: keycloak
          ports:
            - port: 7800
              protocol: TCP
            - port: 57800
              protocol: TCP
    outbound:
      rules:
        - application: keycloak
          ports:
            - port: 7800
              protocol: TCP
            - port: 57800
              protocol: TCP
```

The `ports` field only changes the Kubernetes network policy. If the application has an Istio sidecar, Istio blocks traffic to a pod address that it does not know. For a stateful application, add the ports to `additionalPorts`. Skiperator then adds the ports to the headless Service, and Istio knows the address of each pod. A port name that starts with `tcp-` tells Istio to send the traffic as TCP.

##### Limit: applications that are not stateful

With a sidecar, this example works only for a stateful application (`spec.stateful.enabled: true`). An application that is not stateful has no headless Service, so Istio does not know the address of each pod. The pods can reach each other through the normal Service, but not directly on a pod address. Istio drops this traffic without an error. Software such as JGroups connects directly to pod addresses, so it does not work.

You can fix this in one of these ways:

- Make the application stateful. `spec.stateful.enabled` is immutable, so an existing Application that is not stateful cannot be updated. Delete the Application and create it again with stateful enabled.
- Create your own headless Service (`clusterIP: None`) with the same pod selector and ports. If the pods become ready only after they find their peers, set `publishNotReadyAddresses: true` on it, as Skiperator does for its own headless Service.

Tell the SKIP team if you need Skiperator to support this for applications that are not stateful.

### Outbound rules

Outbound rules specify what other applications your application sends traffic to. `spec.accessPolicy.outbound.rules` specifies which applications in the same cluster to open for. To open for external applications, use the field `spec.accessPolicy.outbound.external`.

#### Send requests to other app in the same namespace

For app `app-a` to be able to send requests to `app-b` in the same cluster and the same namespace, this specification is needed for `app-a`:

```yaml
apiVersion: "skiperator.kartverket.no/v1alpha1"
kind: "Application"
metadata:
  name: app-a
...
spec:
  ...
  accessPolicy:
    outbound:
      rules:
        - application: app-b
```

#### Send requests to other app in the another namespace

For app `app-a` to be able to send requests requests to `app-b` in the same cluster but in another namespace \(`othernamespace`\), this specification is needed for `app-a`:

```yaml
apiVersion: "skiperator.kartverket.no/v1alpha1"
kind: "Application"
metadata:
  name: app-a
...
spec:
  ...
  accessPolicy:
    outbound:
      rules:
        - application: app-b
          namespace: othernamespace
```

#### External services

In order to send requests to services outside of the cluster, `external.host` is needed:

```yaml
apiVersion: "skiperator.kartverket.no/v1alpha1"
kind: "Application"
metadata:
  name: app-a
...
spec:
  ...
  accessPolicy:
    outbound:
      external: 
        - host: www.external-application.com
```

### Advanced: Resources created by Skiperator

The previous application manifest examples will create Kubernetes Network Policies.

#### Kubernetes Network Policy

**Default policy**

Every app created will have some entries added to their default network policy that allows traffic to Anthos Service Mesh (ASM) and kube-dns.
Apps that specify ingresses will also have an ingress rule that allows traffic from the istio ingress to the pod in question.
These policies will be created for every app, also those who don't have any access policies specified.

```yaml
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: appname-ingress
  namespace: teamname
spec:
  ingress:
  - from:
    - podSelector:
        matchLabels:
          ingress: external
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: istio-system
  podSelector:
    matchLabels:
      app: appname
  policyTypes:
    - Ingress
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: appname-egress
  namespace: teamname
spec:
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: kube-system
    ports:
    - port: 53
      protocol: UDP
    - port: 53
      protocol: TCP
  - to:
    - namespaceSelector:
        matchLabels:
          istio: system
      podSelector:
        matchLabels:
          istio: istiod
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: istio-system
      podSelector:
        matchLabels:
          egress: external
  # Allow traffic to internet and re-disallows local traffic
  # This only applies to traffic that does not use the egress gateway, i.e.
  # sidecar pods, which allows meshca to issue certificates
  - to:
    - ipBlock:
        cidr: 0.0.0.0/0
        except:
        # IANA standard local network
        - 10.0.0.0/8
        - 192.168.0.0/16
        - 172.16.0.0/20
  - to: # For GKE data plane v2
    - ipBlock:
        cidr: 169.254.169.254/32
  podSelector:
    matchLabels:
      app: appname
  policyTypes:
    - Egress
```

**Kubernetes network policies**

The applications specified in `spec.accessPolicy.inbound.rules` and `spec.accessPolicy.outbound.rules` will append these fields to the corresponding ingress- and egress Network Policies:

```yaml
apiVersion: extensions/v1beta1
kind: NetworkPolicy
...
  - to:
    - namespaceSelector:
        matchLabels:
          name: othernamespace
      podSelector:
        matchLabels:
          app: app-b
    - podSelector:
        matchLabels:
          app: app-b
  ...
  - from:
    - namespaceSelector:
        matchLabels:
          name: othernamespace
      podSelector:
        matchLabels:
          app: app-b
    - podSelector:
        matchLabels:
          app: app-b
```
