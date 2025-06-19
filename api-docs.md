# API Reference

Packages:

- [skiperator.kartverket.no/v1alpha1](#skiperatorkartverketnov1alpha1)

# skiperator.kartverket.no/v1alpha1

Resource Types:

- [Application](#application)

- [Routing](#routing)

- [SKIPJob](#skipjob)




## Application
<sup><sup>[↩ Parent](#skiperatorkartverketnov1alpha1 )</sup></sup>






Application

Root object for Application resource. An application resource is a resource for easily managing a Dockerized container within the context of a Kartverket cluster.
This allows product teams to avoid the need to set up networking on the cluster, as well as a lot of out of the box security features.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>skiperator.kartverket.no/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Application</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#applicationspec">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationstatus">status</a></b></td>
        <td>object</td>
        <td>
          SkiperatorStatus

A status field shown on a Skiperator resource which contains information regarding deployment of the resource.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec
<sup><sup>[↩ Parent](#application)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          The image the application will run. This image will be added to a Deployment resource<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          The port the deployment exposes<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#applicationspecaccesspolicy">accessPolicy</a></b></td>
        <td>object</td>
        <td>
          The root AccessPolicy for managing zero trust access to your Application. See AccessPolicy for more information.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecadditionalportsindex">additionalPorts</a></b></td>
        <td>[]object</td>
        <td>
          An optional list of extra port to expose on a pod level basis,
for example so Instana or other APM tools can reach it<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>appProtocol</b></td>
        <td>enum</td>
        <td>
          Protocol that the application speaks.<br/>
          <br/>
            <i>Enum</i>: http, tcp, udp<br/>
            <i>Default</i>: http<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecauthorizationsettings">authorizationSettings</a></b></td>
        <td>object</td>
        <td>
          Used for allow listing certain default blocked endpoints, such as /actuator/ end points<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          Override the command set in the Dockerfile. Usually only used when debugging
or running third-party containers where you don't have control over the Dockerfile<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enablePDB</b></td>
        <td>boolean</td>
        <td>
          Whether to enable automatic Pod Disruption Budget creation for this application.<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecenvindex">env</a></b></td>
        <td>[]object</td>
        <td>
          Environment variables that will be set inside the Deployment's Pod. See https://pkg.go.dev/k8s.io/api/core/v1#EnvVar for examples.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecenvfromindex">envFrom</a></b></td>
        <td>[]object</td>
        <td>
          Environment variables mounted from files. When specified all the keys of the
resource will be assigned as environment variables. Supports both configmaps
and secrets.

For mounting as files see FilesFrom.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecfilesfromindex">filesFrom</a></b></td>
        <td>[]object</td>
        <td>
          Mounting volumes into the Deployment are done using the FilesFrom argument

FilesFrom supports ConfigMaps, Secrets and PVCs. The Application resource
assumes these have already been created by you, and will fail if this is not the case.

For mounting environment variables see EnvFrom.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecgcp">gcp</a></b></td>
        <td>object</td>
        <td>
          GCP is used to configure Google Cloud Platform specific settings for the application.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecidporten">idporten</a></b></td>
        <td>object</td>
        <td>
          Settings for IDPorten integration with Digitaliseringsdirektoratet<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ingresses</b></td>
        <td>[]string</td>
        <td>
          Any external hostnames that route to this application. Using a skip.statkart.no-address
will make the application reachable for kartverket-clients (internal), other addresses
make the app reachable on the internet. Note that other addresses than skip.statkart.no
(also known as pretty hostnames) requires additional DNS setup.
The below hostnames will also have TLS certificates issued and be reachable on both
HTTP and HTTPS.

Ingresses must be lowercase, contain no spaces, be a non-empty string, and have a hostname/domain separated by a period
They can optionally be suffixed with a plus and name of a custom TLS secret located in the istio-gateways namespace.
E.g. "foo.atkv3-dev.kartverket-intern.cloud+env-wildcard-cert"<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          Labels can be used if you want every resource created by your application to
have the same labels, including your application. This could for example be useful for
metrics, where a certain label and the corresponding resources liveliness can be combined.
Any amount of labels can be added as wanted, and they will all cascade down to all resources.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecliveness">liveness</a></b></td>
        <td>object</td>
        <td>
          Liveness probes define a resource that returns 200 OK when the app is running
as intended. Returning a non-200 code will make kubernetes restart the app.
Liveness is optional, but when provided, path and port are required

See Probe for structure definition.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecmaskinporten">maskinporten</a></b></td>
        <td>object</td>
        <td>
          Settings for Maskinporten integration with Digitaliseringsdirektoratet<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecpodsettings">podSettings</a></b></td>
        <td>object</td>
        <td>
          PodSettings are used to apply specific settings to the Pod Template used by Skiperator to create Deployments. This allows you to set
things like annotations on the Pod to change the behaviour of sidecars, and set relevant Pod options such as TerminationGracePeriodSeconds.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>priority</b></td>
        <td>enum</td>
        <td>
          An optional priority. Supported values are 'low', 'medium' and 'high'.
The default value is 'medium'.

Most workloads should not have to specify this field. If you think you
do, please consult with SKIP beforehand.<br/>
          <br/>
            <i>Enum</i>: low, medium, high<br/>
            <i>Default</i>: medium<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecprometheus">prometheus</a></b></td>
        <td>object</td>
        <td>
          Optional settings for how Prometheus compatible metrics should be scraped.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecreadiness">readiness</a></b></td>
        <td>object</td>
        <td>
          Readiness probes define a resource that returns 200 OK when the app is running
as intended. Kubernetes will wait until the resource returns 200 OK before
marking the pod as Running and progressing with the deployment strategy.
Readiness is optional, but when provided, path and port are required<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>redirectToHTTPS</b></td>
        <td>boolean</td>
        <td>
          Controls whether the application will automatically redirect all HTTP calls to HTTPS via the istio VirtualService.
This redirect does not happen on the route /.well-known/acme-challenge/, as the ACME challenge can only be done on port 80.<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>replicas</b></td>
        <td>JSON</td>
        <td>
          The number of replicas can either be specified as a static number as follows:

	replicas: 2

Or by specifying a range between min and max to enable HorizontalPodAutoscaling.
The default value for replicas is:
	replicas:
		min: 2
		max: 5
		targetCpuUtilization: 80
Using autoscaling is the recommended configuration for replicas.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resourceLabels</b></td>
        <td>map[string]map[string]string</td>
        <td>
          ResourceLabels can be used if you want to add a label to a specific resources created by
the application. One such label could for example be set on a Deployment, such that
the deployment avoids certain rules from Gatekeeper, or similar. Any amount of labels may be added per ResourceLabels item.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecresources">resources</a></b></td>
        <td>object</td>
        <td>
          ResourceRequirements to apply to the deployment. It's common to set some of these to
prevent the app from swelling in resource usage and consuming all the
resources of other apps on the cluster.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecstartup">startup</a></b></td>
        <td>object</td>
        <td>
          Kubernetes uses startup probes to know when a container application has started.
If such a probe is configured, it disables liveness and readiness checks until it
succeeds, making sure those probes don't interfere with the application startup.
This can be used to adopt liveness checks on slow starting containers, avoiding them
getting killed by Kubernetes before they are up and running.
Startup is optional, but when provided, path and port are required<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecstrategy">strategy</a></b></td>
        <td>object</td>
        <td>
          Defines an alternative strategy for the Kubernetes deployment. This is useful when
the default strategy, RollingUpdate, is not usable. Setting type to
Recreate will take down all the pods before starting new pods, whereas the
default of RollingUpdate will try to start the new pods before taking down the
old ones.

Valid values are: RollingUpdate, Recreate. Default is RollingUpdate<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>team</b></td>
        <td>string</td>
        <td>
          Team specifies the team who owns this particular app.
Usually sourced from the namespace label.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.accessPolicy
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



The root AccessPolicy for managing zero trust access to your Application. See AccessPolicy for more information.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#applicationspecaccesspolicyinbound">inbound</a></b></td>
        <td>object</td>
        <td>
          Inbound specifies the ingress rules. Which apps on the cluster can talk to this app?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecaccesspolicyoutbound">outbound</a></b></td>
        <td>object</td>
        <td>
          Outbound specifies egress rules. Which apps on the cluster and the
internet is the Application allowed to send requests to?<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.accessPolicy.inbound
<sup><sup>[↩ Parent](#applicationspecaccesspolicy)</sup></sup>



Inbound specifies the ingress rules. Which apps on the cluster can talk to this app?

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#applicationspecaccesspolicyinboundrulesindex">rules</a></b></td>
        <td>[]object</td>
        <td>
          The rules list specifies a list of applications. When no namespace is
specified it refers to an app in the current namespace. For apps in
other namespaces namespace is required<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Application.spec.accessPolicy.inbound.rules[index]
<sup><sup>[↩ Parent](#applicationspecaccesspolicyinbound)</sup></sup>



InternalRule

The rules list specifies a list of applications. When no namespace is
specified it refers to an app in the current namespace. For apps in
other namespaces, namespace is required.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>application</b></td>
        <td>string</td>
        <td>
          The name of the Application you are allowing traffic to/from. If you wish to allow traffic from a SKIPJob, this field should
be suffixed with -skipjob<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          The namespace in which the Application you are allowing traffic to/from resides. If unset, uses namespace of Application.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespacesByLabel</b></td>
        <td>map[string]string</td>
        <td>
          Namespace label value-pair in which the Application you are allowing traffic to/from resides. If both namespace and namespacesByLabel are set, namespace takes precedence and namespacesByLabel is omitted.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecaccesspolicyinboundrulesindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above application.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.accessPolicy.inbound.rules[index].ports[index]
<sup><sup>[↩ Parent](#applicationspecaccesspolicyinboundrulesindex)</sup></sup>



NetworkPolicyPort describes a port to allow traffic on

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>endPort</b></td>
        <td>integer</td>
        <td>
          endPort indicates that the range of ports from port to endPort if set, inclusive,
should be allowed by the policy. This field cannot be defined if the port field
is not defined or if the port field is defined as a named (string) port.
The endPort must be equal or greater than port.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          port represents the port on the given protocol. This can either be a numerical or named
port on a pod. If this field is not provided, this matches all port names and
numbers.
If present, only traffic on the specified protocol AND port will be matched.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          protocol represents the protocol (TCP, UDP, or SCTP) which traffic must match.
If not specified, this field defaults to TCP.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.accessPolicy.outbound
<sup><sup>[↩ Parent](#applicationspecaccesspolicy)</sup></sup>



Outbound specifies egress rules. Which apps on the cluster and the
internet is the Application allowed to send requests to?

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#applicationspecaccesspolicyoutboundexternalindex">external</a></b></td>
        <td>[]object</td>
        <td>
          External specifies which applications on the internet the application
can reach. Only host is required unless it is on another port than HTTPS port 443.
If other ports or protocols are required then `ports` must be specified as well<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecaccesspolicyoutboundrulesindex">rules</a></b></td>
        <td>[]object</td>
        <td>
          Rules apply the same in-cluster rules as InboundPolicy<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.accessPolicy.outbound.external[index]
<sup><sup>[↩ Parent](#applicationspecaccesspolicyoutbound)</sup></sup>



ExternalRule

Describes a rule for allowing your Application to route traffic to external applications and hosts.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>ip</b></td>
        <td>string</td>
        <td>
          Non-HTTP requests (i.e. using the TCP protocol) need to use IP in addition to hostname
Only required for TCP requests.

Note: Hostname must always be defined even if IP is set statically<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecaccesspolicyoutboundexternalindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above hostname. When not specified HTTP and
HTTPS on port 80 and 443 respectively are put into the allowlist<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.accessPolicy.outbound.external[index].ports[index]
<sup><sup>[↩ Parent](#applicationspecaccesspolicyoutboundexternalindex)</sup></sup>



ExternalPort

A custom port describing an external host

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is required and is an arbitrary name. Must be unique within all ExternalRule ports.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          The port number of the external host<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>enum</td>
        <td>
          The protocol to use for communication with the host. Only HTTP, HTTPS and TCP are supported.<br/>
          <br/>
            <i>Enum</i>: HTTP, HTTPS, TCP<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Application.spec.accessPolicy.outbound.rules[index]
<sup><sup>[↩ Parent](#applicationspecaccesspolicyoutbound)</sup></sup>



InternalRule

The rules list specifies a list of applications. When no namespace is
specified it refers to an app in the current namespace. For apps in
other namespaces, namespace is required.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>application</b></td>
        <td>string</td>
        <td>
          The name of the Application you are allowing traffic to/from. If you wish to allow traffic from a SKIPJob, this field should
be suffixed with -skipjob<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          The namespace in which the Application you are allowing traffic to/from resides. If unset, uses namespace of Application.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespacesByLabel</b></td>
        <td>map[string]string</td>
        <td>
          Namespace label value-pair in which the Application you are allowing traffic to/from resides. If both namespace and namespacesByLabel are set, namespace takes precedence and namespacesByLabel is omitted.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecaccesspolicyoutboundrulesindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above application.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.accessPolicy.outbound.rules[index].ports[index]
<sup><sup>[↩ Parent](#applicationspecaccesspolicyoutboundrulesindex)</sup></sup>



NetworkPolicyPort describes a port to allow traffic on

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>endPort</b></td>
        <td>integer</td>
        <td>
          endPort indicates that the range of ports from port to endPort if set, inclusive,
should be allowed by the policy. This field cannot be defined if the port field
is not defined or if the port field is defined as a named (string) port.
The endPort must be equal or greater than port.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          port represents the port on the given protocol. This can either be a numerical or named
port on a pod. If this field is not provided, this matches all port names and
numbers.
If present, only traffic on the specified protocol AND port will be matched.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          protocol represents the protocol (TCP, UDP, or SCTP) which traffic must match.
If not specified, this field defaults to TCP.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.additionalPorts[index]
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>enum</td>
        <td>
          Protocol defines network protocols supported for things like container ports.<br/>
          <br/>
            <i>Enum</i>: TCP, UDP, SCTP<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Application.spec.authorizationSettings
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



Used for allow listing certain default blocked endpoints, such as /actuator/ end points

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>allowAll</b></td>
        <td>boolean</td>
        <td>
          Allows all endpoints by not creating an AuthorizationPolicy, and ignores the content of AllowList.
If field is false, the contents of AllowList will be used instead if AllowList is set.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowList</b></td>
        <td>[]string</td>
        <td>
          Allows specific endpoints. Common endpoints one might want to allow include /actuator/health, /actuator/startup, /actuator/info.

Note that endpoints are matched specifically on the input, so if you allow /actuator/health, you will *not* allow /actuator/health/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.env[index]
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



EnvVar represents an environment variable present in a Container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the environment variable. Must be a C_IDENTIFIER.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Variable references $(VAR_NAME) are expanded
using the previously defined environment variables in the container and
any service environment variables. If a variable cannot be resolved,
the reference in the input string will be unchanged. Double $$ are reduced
to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e.
"$$(VAR_NAME)" will produce the string literal "$(VAR_NAME)".
Escaped references will never be expanded, regardless of whether the variable
exists or not.
Defaults to "".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecenvindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Source for the environment variable's value. Cannot be used if value is not empty.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.env[index].valueFrom
<sup><sup>[↩ Parent](#applicationspecenvindex)</sup></sup>



Source for the environment variable's value. Cannot be used if value is not empty.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#applicationspecenvindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecenvindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`,
spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecenvindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecenvindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a secret in the pod's namespace<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.env[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#applicationspecenvindexvaluefrom)</sup></sup>



Selects a key of a ConfigMap.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.env[index].valueFrom.fieldRef
<sup><sup>[↩ Parent](#applicationspecenvindexvaluefrom)</sup></sup>



Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`,
spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          Path of the field to select in the specified API version.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          Version of the schema the FieldPath is written in terms of, defaults to "v1".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.env[index].valueFrom.resourceFieldRef
<sup><sup>[↩ Parent](#applicationspecenvindexvaluefrom)</sup></sup>



Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Required: resource to select<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Container name: required for volumes, optional for env vars<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          Specifies the output format of the exposed resources, defaults to "1"<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.env[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#applicationspecenvindexvaluefrom)</sup></sup>



Selects a key of a secret in the pod's namespace

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.envFrom[index]
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>configMap</b></td>
        <td>string</td>
        <td>
          Name of Kubernetes ConfigMap in which the deployment should mount environment variables from. Must be in the same namespace as the Application<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          Name of Kubernetes Secret in which the deployment should mount environment variables from. Must be in the same namespace as the Application<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.filesFrom[index]
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



FilesFrom

Struct representing information needed to mount a Kubernetes resource as a file to a Pod's directory.
One of ConfigMap, Secret, EmptyDir or PersistentVolumeClaim must be present, and just represent the name of the resource in question
NB. Out-of-the-box, skiperator provides a writable 'emptyDir'-volume at '/tmp'

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>mountPath</b></td>
        <td>string</td>
        <td>
          The path to mount the file in the Pods directory. Required.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>configMap</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>emptyDir</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>persistentVolumeClaim</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.gcp
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



GCP is used to configure Google Cloud Platform specific settings for the application.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#applicationspecgcpauth">auth</a></b></td>
        <td>object</td>
        <td>
          Configuration for authenticating a Pod with Google Cloud Platform
For authentication with GCP, to use services like Secret Manager and/or Pub/Sub we need
to set the GCP Service Account Pods should identify as. To allow this, we need the IAM role iam.workloadIdentityUser set on a GCP
service account and bind this to the Pod's Kubernetes SA.
Documentation on how this is done can be found here (Closed Wiki):
https://kartverket.atlassian.net/wiki/spaces/SKIPDOK/pages/422346824/Autentisering+mot+GCP+som+Kubernetes+SA<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecgcpcloudsqlproxy">cloudSqlProxy</a></b></td>
        <td>object</td>
        <td>
          CloudSQL is used to deploy a CloudSQL proxy sidecar in the pod.
This is useful for connecting to CloudSQL databases that require Cloud SQL Auth Proxy.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.gcp.auth
<sup><sup>[↩ Parent](#applicationspecgcp)</sup></sup>



Configuration for authenticating a Pod with Google Cloud Platform
For authentication with GCP, to use services like Secret Manager and/or Pub/Sub we need
to set the GCP Service Account Pods should identify as. To allow this, we need the IAM role iam.workloadIdentityUser set on a GCP
service account and bind this to the Pod's Kubernetes SA.
Documentation on how this is done can be found here (Closed Wiki):
https://kartverket.atlassian.net/wiki/spaces/SKIPDOK/pages/422346824/Autentisering+mot+GCP+som+Kubernetes+SA

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>serviceAccount</b></td>
        <td>string</td>
        <td>
          Name of the service account in which you are trying to authenticate your pod with
Generally takes the form of some-name@some-project-id.iam.gserviceaccount.com<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Application.spec.gcp.cloudSqlProxy
<sup><sup>[↩ Parent](#applicationspecgcp)</sup></sup>



CloudSQL is used to deploy a CloudSQL proxy sidecar in the pod.
This is useful for connecting to CloudSQL databases that require Cloud SQL Auth Proxy.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>connectionName</b></td>
        <td>string</td>
        <td>
          Connection name for the CloudSQL instance. Found in the Google Cloud Console under your CloudSQL resource.
The format is "projectName:region:instanceName" E.g. "skip-prod-bda1:europe-north1:my-db".<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>ip</b></td>
        <td>string</td>
        <td>
          The IP address of the CloudSQL instance. This is used to create a serviceentry for the CloudSQL proxy.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>serviceAccount</b></td>
        <td>string</td>
        <td>
          Service account used by cloudsql auth proxy. This service account must have the roles/cloudsql.client role.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>version</b></td>
        <td>string</td>
        <td>
          Image version for the CloudSQL proxy sidecar.<br/>
          <br/>
            <i>Default</i>: 2.8.0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.idporten
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



Settings for IDPorten integration with Digitaliseringsdirektoratet

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          Whether to enable provisioning of an ID-porten client.
If enabled, an ID-porten client be provisioned.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>accessTokenLifetime</b></td>
        <td>integer</td>
        <td>
          AccessTokenLifetime is the lifetime in seconds for any issued access token from ID-porten.

If unspecified, defaults to `3600` seconds (1 hour).<br/>
          <br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 3600<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clientName</b></td>
        <td>string</td>
        <td>
          The name of the Client as shown in Digitaliseringsdirektoratet's Samarbeidsportal
Meant to be a human-readable name for separating clients in the portal<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clientURI</b></td>
        <td>string</td>
        <td>
          ClientURI is the URL shown to the user at ID-porten when displaying a 'back' button or on errors.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>frontchannelLogoutPath</b></td>
        <td>string</td>
        <td>
          FrontchannelLogoutPath is a valid path for your application where ID-porten sends a request to whenever the user has
initiated a logout elsewhere as part of a single logout (front channel logout) process.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>integrationType</b></td>
        <td>enum</td>
        <td>
          IntegrationType is used to make sensible choices for your client.
Which type of integration you choose will provide guidance on which scopes you can use with the client.
A client can only have one integration type.

NB! It is not possible to change the integration type after creation.<br/>
          <br/>
            <i>Enum</i>: krr, idporten, api_klient<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>postLogoutRedirectPath</b></td>
        <td>string</td>
        <td>
          PostLogoutRedirectPath is a simpler verison of PostLogoutRedirectURIs
that will be appended to the ingress<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>postLogoutRedirectURIs</b></td>
        <td>[]string</td>
        <td>
          PostLogoutRedirectURIs are valid URIs that ID-porten will allow redirecting the end-user to after a single logout
has been initiated and performed by the application.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>redirectPath</b></td>
        <td>string</td>
        <td>
          RedirectPath is a valid path that ID-porten redirects back to after a successful authorization request.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scopes</b></td>
        <td>[]string</td>
        <td>
          Register different oauth2 Scopes on your client.
You will not be able to add a scope to your client that conflicts with the client's IntegrationType.
For example, you can not add a scope that is limited to the IntegrationType `krr` of IntegrationType `idporten`, and vice versa.

Default for IntegrationType `krr` = ("krr:global/kontaktinformasjon.read", "krr:global/digitalpost.read")
Default for IntegrationType `idporten` = ("openid", "profile")
IntegrationType `api_klient` have no Default, checkout Digdir documentation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sessionLifetime</b></td>
        <td>integer</td>
        <td>
          SessionLifetime is the maximum lifetime in seconds for any given user's session in your application.
The timeout starts whenever the user is redirected from the `authorization_endpoint` at ID-porten.

If unspecified, defaults to `7200` seconds (2 hours).
Note: Attempting to refresh the user's `access_token` beyond this timeout will yield an error.<br/>
          <br/>
            <i>Minimum</i>: 3600<br/>
            <i>Maximum</i>: 7200<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.liveness
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



Liveness probes define a resource that returns 200 OK when the app is running
as intended. Returning a non-200 code will make kubernetes restart the app.
Liveness is optional, but when provided, path and port are required

See Probe for structure definition.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that
are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 10<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.
Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.maskinporten
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



Settings for Maskinporten integration with Digitaliseringsdirektoratet

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          If enabled, provisions and configures a Maskinporten client with consumed scopes and/or Exposed scopes with DigDir.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>clientName</b></td>
        <td>string</td>
        <td>
          The name of the Client as shown in Digitaliseringsdirektoratet's Samarbeidsportal
Meant to be a human-readable name for separating clients in the portal<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecmaskinportenscopes">scopes</a></b></td>
        <td>object</td>
        <td>
          Schema to configure Maskinporten clients with consumed scopes and/or exposed scopes.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.maskinporten.scopes
<sup><sup>[↩ Parent](#applicationspecmaskinporten)</sup></sup>



Schema to configure Maskinporten clients with consumed scopes and/or exposed scopes.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#applicationspecmaskinportenscopesconsumesindex">consumes</a></b></td>
        <td>[]object</td>
        <td>
          This is the Schema for the consumes and exposes API.
`consumes` is a list of scopes that your client can request access to.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecmaskinportenscopesexposesindex">exposes</a></b></td>
        <td>[]object</td>
        <td>
          `exposes` is a list of scopes your application want to expose to other organization where access to the scope is based on organization number.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.maskinporten.scopes.consumes[index]
<sup><sup>[↩ Parent](#applicationspecmaskinportenscopes)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          The scope consumed by the application to gain access to an external organization API.
Ensure that the NAV organization has been granted access to the scope prior to requesting access.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Application.spec.maskinporten.scopes.exposes[index]
<sup><sup>[↩ Parent](#applicationspecmaskinportenscopes)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          If Enabled the configured scope is available to be used and consumed by organizations granted access.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          The actual subscope combined with `Product`.
Ensure that `<Product><Name>` matches `Pattern`.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>product</b></td>
        <td>string</td>
        <td>
          The product-area your application belongs to e.g. arbeid, helse ...
This will be included in the final scope `nav:<Product><Name>`.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>accessibleForAll</b></td>
        <td>boolean</td>
        <td>
          Allow any organization to access the scope.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowedIntegrations</b></td>
        <td>[]string</td>
        <td>
          Whitelisting of integration's allowed.
Default is `maskinporten`<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>atMaxAge</b></td>
        <td>integer</td>
        <td>
          Max time in seconds for a issued access_token.
Default is `30` sec.<br/>
          <br/>
            <i>Minimum</i>: 30<br/>
            <i>Maximum</i>: 680<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#applicationspecmaskinportenscopesexposesindexconsumersindex">consumers</a></b></td>
        <td>[]object</td>
        <td>
          External consumers granted access to this scope and able to request access_token.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>delegationSource</b></td>
        <td>enum</td>
        <td>
          Delegation source for the scope. Default is empty, which means no delegation is allowed.<br/>
          <br/>
            <i>Enum</i>: altinn<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>separator</b></td>
        <td>string</td>
        <td>
          Separator is the character that separates `product` and `name` in the final scope:
`scope := <prefix>:<product><separator><name>`
This overrides the default separator.
The default separator is `:`. If `name` contains `/`, the default separator is instead `/`.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.maskinporten.scopes.exposes[index].consumers[index]
<sup><sup>[↩ Parent](#applicationspecmaskinportenscopesexposesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>orgno</b></td>
        <td>string</td>
        <td>
          The external business/organization number.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          This is a describing field intended for clarity not used for any other purpose.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.podSettings
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



PodSettings are used to apply specific settings to the Pod Template used by Skiperator to create Deployments. This allows you to set
things like annotations on the Pod to change the behaviour of sidecars, and set relevant Pod options such as TerminationGracePeriodSeconds.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          Annotations that are set on Pods created by Skiperator. These annotations can for example be used to change the behaviour of sidecars and similar.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>disablePodSpreadTopologyConstraints</b></td>
        <td>boolean</td>
        <td>
          DisablePodSpreadTopologyConstraints specifies whether to disable the addition of Pod Topology Spread Constraints to
a given pod.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          TerminationGracePeriodSeconds determines how long Kubernetes waits after a SIGTERM signal sent to a Pod before terminating the pod. If your application uses longer than
30 seconds to terminate, you should increase TerminationGracePeriodSeconds.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Default</i>: 30<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.prometheus
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



Optional settings for how Prometheus compatible metrics should be scraped.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          The port number or name where metrics are exposed (at the Pod level).<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>allowAllMetrics</b></td>
        <td>boolean</td>
        <td>
          Setting AllowAllMetrics to true will ensure all exposed metrics are scraped. Otherwise, a list of predefined
metrics will be dropped by default. See util/constants.go for the default list.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The HTTP path where Prometheus compatible metrics exists<br/>
          <br/>
            <i>Default</i>: /metrics<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.readiness
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



Readiness probes define a resource that returns 200 OK when the app is running
as intended. Kubernetes will wait until the resource returns 200 OK before
marking the pod as Running and progressing with the deployment strategy.
Readiness is optional, but when provided, path and port are required

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that
are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 10<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.
Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.resources
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



ResourceRequirements to apply to the deployment. It's common to set some of these to
prevent the app from swelling in resource usage and consuming all the
resources of other apps on the cluster.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits set the maximum the app is allowed to use. Exceeding this limit will
make kubernetes kill the app and restart it.

Limits can be set on the CPU and memory, but it is not recommended to put a limit on CPU, see: https://home.robusta.dev/blog/stop-using-cpu-limits<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests set the initial allocation that is done for the app and will
thus be available to the app on startup. More is allocated on demand
until the limit is reached.

Requests can be set on the CPU and memory.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.startup
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



Kubernetes uses startup probes to know when a container application has started.
If such a probe is configured, it disables liveness and readiness checks until it
succeeds, making sure those probes don't interfere with the application startup.
This can be used to adopt liveness checks on slow starting containers, avoiding them
getting killed by Kubernetes before they are up and running.
Startup is optional, but when provided, path and port are required

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that
are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 10<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.
Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.spec.strategy
<sup><sup>[↩ Parent](#applicationspec)</sup></sup>



Defines an alternative strategy for the Kubernetes deployment. This is useful when
the default strategy, RollingUpdate, is not usable. Setting type to
Recreate will take down all the pods before starting new pods, whereas the
default of RollingUpdate will try to start the new pods before taking down the
old ones.

Valid values are: RollingUpdate, Recreate. Default is RollingUpdate

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Valid values are: RollingUpdate, Recreate. Default is RollingUpdate<br/>
          <br/>
            <i>Enum</i>: RollingUpdate, Recreate<br/>
            <i>Default</i>: RollingUpdate<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.status
<sup><sup>[↩ Parent](#application)</sup></sup>



SkiperatorStatus

A status field shown on a Skiperator resource which contains information regarding deployment of the resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>accessPolicies</b></td>
        <td>string</td>
        <td>
          Indicates if access policies are valid<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#applicationstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#applicationstatussubresourceskey">subresources</a></b></td>
        <td>map[string]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#applicationstatussummary">summary</a></b></td>
        <td>object</td>
        <td>
          Status<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Application.status.conditions[index]
<sup><sup>[↩ Parent](#applicationstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Application.status.subresources[key]
<sup><sup>[↩ Parent](#applicationstatus)</sup></sup>



Status

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: hello<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: Synced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: hello<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Application.status.summary
<sup><sup>[↩ Parent](#applicationstatus)</sup></sup>



Status

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: hello<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: Synced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: hello<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>

## Routing
<sup><sup>[↩ Parent](#skiperatorkartverketnov1alpha1 )</sup></sup>








<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>skiperator.kartverket.no/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Routing</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#routingspec">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#routingstatus">status</a></b></td>
        <td>object</td>
        <td>
          SkiperatorStatus

A status field shown on a Skiperator resource which contains information regarding deployment of the resource.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Routing.spec
<sup><sup>[↩ Parent](#routing)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>hostname</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#routingspecroutesindex">routes</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>redirectToHTTPS</b></td>
        <td>boolean</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Routing.spec.routes[index]
<sup><sup>[↩ Parent](#routingspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>pathPrefix</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>targetApp</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>rewriteUri</b></td>
        <td>boolean</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Routing.status
<sup><sup>[↩ Parent](#routing)</sup></sup>



SkiperatorStatus

A status field shown on a Skiperator resource which contains information regarding deployment of the resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>accessPolicies</b></td>
        <td>string</td>
        <td>
          Indicates if access policies are valid<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#routingstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#routingstatussubresourceskey">subresources</a></b></td>
        <td>map[string]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#routingstatussummary">summary</a></b></td>
        <td>object</td>
        <td>
          Status<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Routing.status.conditions[index]
<sup><sup>[↩ Parent](#routingstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Routing.status.subresources[key]
<sup><sup>[↩ Parent](#routingstatus)</sup></sup>



Status

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: hello<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: Synced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: hello<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Routing.status.summary
<sup><sup>[↩ Parent](#routingstatus)</sup></sup>



Status

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: hello<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: Synced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: hello<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>

## SKIPJob
<sup><sup>[↩ Parent](#skiperatorkartverketnov1alpha1 )</sup></sup>






SKIPJob is the Schema for the skipjobs API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>skiperator.kartverket.no/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>SKIPJob</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#skipjobspec">spec</a></b></td>
        <td>object</td>
        <td>
          SKIPJobSpec defines the desired state of SKIPJob

A SKIPJob is either defined as a one-off or a scheduled job. If the Cron field is set for SKIPJob, it may not be removed. If the Cron field is unset, it may not be added.
The Container field of a SKIPJob is only mutable if the Cron field is set. If unset, you must delete your SKIPJob to change container settings.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#skipjobstatus">status</a></b></td>
        <td>object</td>
        <td>
          SkiperatorStatus

A status field shown on a Skiperator resource which contains information regarding deployment of the resource.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec
<sup><sup>[↩ Parent](#skipjob)</sup></sup>



SKIPJobSpec defines the desired state of SKIPJob

A SKIPJob is either defined as a one-off or a scheduled job. If the Cron field is set for SKIPJob, it may not be removed. If the Cron field is unset, it may not be added.
The Container field of a SKIPJob is only mutable if the Cron field is set. If unset, you must delete your SKIPJob to change container settings.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#skipjobspeccontainer">container</a></b></td>
        <td>object</td>
        <td>
          Settings for the Pods running in the job. Fields are mostly the same as an Application, and are (probably) better documented there. Some fields are omitted, but none added.
Once set, you may not change Container without deleting your current SKIPJob<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccron">cron</a></b></td>
        <td>object</td>
        <td>
          Settings for the Job if you are running a scheduled job. Optional as Jobs may be one-off.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspecjob">job</a></b></td>
        <td>object</td>
        <td>
          Settings for the actual Job. If you use a scheduled job, the settings in here will also specify the template of the job.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspecprometheus">prometheus</a></b></td>
        <td>object</td>
        <td>
          Prometheus settings for pod running in job. Fields are identical to Application and if set,
a podmonitoring object is created.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container
<sup><sup>[↩ Parent](#skipjobspec)</sup></sup>



Settings for the Pods running in the job. Fields are mostly the same as an Application, and are (probably) better documented there. Some fields are omitted, but none added.
Once set, you may not change Container without deleting your current SKIPJob

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicy">accessPolicy</a></b></td>
        <td>object</td>
        <td>
          AccessPolicy

Zero trust dictates that only applications with a reason for being able
to access another resource should be able to reach it. This is set up by
default by denying all ingress and egress traffic from the Pods in the
Deployment. The AccessPolicy field is an allowlist of other applications and hostnames
that are allowed to talk with this Application and which resources this app can talk to<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontaineradditionalportsindex">additionalPorts</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainerenvindex">env</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainerenvfromindex">envFrom</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainerfilesfromindex">filesFrom</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainergcp">gcp</a></b></td>
        <td>object</td>
        <td>
          GCP

Configuration for interacting with Google Cloud Platform<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainerliveness">liveness</a></b></td>
        <td>object</td>
        <td>
          Probe

Type configuration for all types of Kubernetes probes.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainerpodsettings">podSettings</a></b></td>
        <td>object</td>
        <td>
          PodSettings<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>priority</b></td>
        <td>enum</td>
        <td>
          <br/>
          <br/>
            <i>Enum</i>: low, medium, high<br/>
            <i>Default</i>: medium<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainerreadiness">readiness</a></b></td>
        <td>object</td>
        <td>
          Probe

Type configuration for all types of Kubernetes probes.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainerresources">resources</a></b></td>
        <td>object</td>
        <td>
          ResourceRequirements

A simplified version of the Kubernetes native ResourceRequirement field, in which only Limits and Requests are present.
For the units used for resources, see https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-units-in-kubernetes<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>restartPolicy</b></td>
        <td>enum</td>
        <td>
          RestartPolicy describes how the container should be restarted.
Only one of the following restart policies may be specified.
If none of the following policies is specified, the default one
is RestartPolicyAlways.<br/>
          <br/>
            <i>Enum</i>: OnFailure, Never<br/>
            <i>Default</i>: Never<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainerstartup">startup</a></b></td>
        <td>object</td>
        <td>
          Probe

Type configuration for all types of Kubernetes probes.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.accessPolicy
<sup><sup>[↩ Parent](#skipjobspeccontainer)</sup></sup>



AccessPolicy

Zero trust dictates that only applications with a reason for being able
to access another resource should be able to reach it. This is set up by
default by denying all ingress and egress traffic from the Pods in the
Deployment. The AccessPolicy field is an allowlist of other applications and hostnames
that are allowed to talk with this Application and which resources this app can talk to

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyinbound">inbound</a></b></td>
        <td>object</td>
        <td>
          Inbound specifies the ingress rules. Which apps on the cluster can talk to this app?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyoutbound">outbound</a></b></td>
        <td>object</td>
        <td>
          Outbound specifies egress rules. Which apps on the cluster and the
internet is the Application allowed to send requests to?<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.accessPolicy.inbound
<sup><sup>[↩ Parent](#skipjobspeccontaineraccesspolicy)</sup></sup>



Inbound specifies the ingress rules. Which apps on the cluster can talk to this app?

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyinboundrulesindex">rules</a></b></td>
        <td>[]object</td>
        <td>
          The rules list specifies a list of applications. When no namespace is
specified it refers to an app in the current namespace. For apps in
other namespaces namespace is required<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.accessPolicy.inbound.rules[index]
<sup><sup>[↩ Parent](#skipjobspeccontaineraccesspolicyinbound)</sup></sup>



InternalRule

The rules list specifies a list of applications. When no namespace is
specified it refers to an app in the current namespace. For apps in
other namespaces, namespace is required.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>application</b></td>
        <td>string</td>
        <td>
          The name of the Application you are allowing traffic to/from. If you wish to allow traffic from a SKIPJob, this field should
be suffixed with -skipjob<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          The namespace in which the Application you are allowing traffic to/from resides. If unset, uses namespace of Application.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespacesByLabel</b></td>
        <td>map[string]string</td>
        <td>
          Namespace label value-pair in which the Application you are allowing traffic to/from resides. If both namespace and namespacesByLabel are set, namespace takes precedence and namespacesByLabel is omitted.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyinboundrulesindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above application.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.accessPolicy.inbound.rules[index].ports[index]
<sup><sup>[↩ Parent](#skipjobspeccontaineraccesspolicyinboundrulesindex)</sup></sup>



NetworkPolicyPort describes a port to allow traffic on

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>endPort</b></td>
        <td>integer</td>
        <td>
          endPort indicates that the range of ports from port to endPort if set, inclusive,
should be allowed by the policy. This field cannot be defined if the port field
is not defined or if the port field is defined as a named (string) port.
The endPort must be equal or greater than port.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          port represents the port on the given protocol. This can either be a numerical or named
port on a pod. If this field is not provided, this matches all port names and
numbers.
If present, only traffic on the specified protocol AND port will be matched.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          protocol represents the protocol (TCP, UDP, or SCTP) which traffic must match.
If not specified, this field defaults to TCP.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.accessPolicy.outbound
<sup><sup>[↩ Parent](#skipjobspeccontaineraccesspolicy)</sup></sup>



Outbound specifies egress rules. Which apps on the cluster and the
internet is the Application allowed to send requests to?

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyoutboundexternalindex">external</a></b></td>
        <td>[]object</td>
        <td>
          External specifies which applications on the internet the application
can reach. Only host is required unless it is on another port than HTTPS port 443.
If other ports or protocols are required then `ports` must be specified as well<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyoutboundrulesindex">rules</a></b></td>
        <td>[]object</td>
        <td>
          Rules apply the same in-cluster rules as InboundPolicy<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.accessPolicy.outbound.external[index]
<sup><sup>[↩ Parent](#skipjobspeccontaineraccesspolicyoutbound)</sup></sup>



ExternalRule

Describes a rule for allowing your Application to route traffic to external applications and hosts.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>ip</b></td>
        <td>string</td>
        <td>
          Non-HTTP requests (i.e. using the TCP protocol) need to use IP in addition to hostname
Only required for TCP requests.

Note: Hostname must always be defined even if IP is set statically<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyoutboundexternalindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above hostname. When not specified HTTP and
HTTPS on port 80 and 443 respectively are put into the allowlist<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.accessPolicy.outbound.external[index].ports[index]
<sup><sup>[↩ Parent](#skipjobspeccontaineraccesspolicyoutboundexternalindex)</sup></sup>



ExternalPort

A custom port describing an external host

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is required and is an arbitrary name. Must be unique within all ExternalRule ports.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          The port number of the external host<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>enum</td>
        <td>
          The protocol to use for communication with the host. Only HTTP, HTTPS and TCP are supported.<br/>
          <br/>
            <i>Enum</i>: HTTP, HTTPS, TCP<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.accessPolicy.outbound.rules[index]
<sup><sup>[↩ Parent](#skipjobspeccontaineraccesspolicyoutbound)</sup></sup>



InternalRule

The rules list specifies a list of applications. When no namespace is
specified it refers to an app in the current namespace. For apps in
other namespaces, namespace is required.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>application</b></td>
        <td>string</td>
        <td>
          The name of the Application you are allowing traffic to/from. If you wish to allow traffic from a SKIPJob, this field should
be suffixed with -skipjob<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          The namespace in which the Application you are allowing traffic to/from resides. If unset, uses namespace of Application.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespacesByLabel</b></td>
        <td>map[string]string</td>
        <td>
          Namespace label value-pair in which the Application you are allowing traffic to/from resides. If both namespace and namespacesByLabel are set, namespace takes precedence and namespacesByLabel is omitted.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyoutboundrulesindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above application.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.accessPolicy.outbound.rules[index].ports[index]
<sup><sup>[↩ Parent](#skipjobspeccontaineraccesspolicyoutboundrulesindex)</sup></sup>



NetworkPolicyPort describes a port to allow traffic on

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>endPort</b></td>
        <td>integer</td>
        <td>
          endPort indicates that the range of ports from port to endPort if set, inclusive,
should be allowed by the policy. This field cannot be defined if the port field
is not defined or if the port field is defined as a named (string) port.
The endPort must be equal or greater than port.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          port represents the port on the given protocol. This can either be a numerical or named
port on a pod. If this field is not provided, this matches all port names and
numbers.
If present, only traffic on the specified protocol AND port will be matched.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          protocol represents the protocol (TCP, UDP, or SCTP) which traffic must match.
If not specified, this field defaults to TCP.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.additionalPorts[index]
<sup><sup>[↩ Parent](#skipjobspeccontainer)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>enum</td>
        <td>
          Protocol defines network protocols supported for things like container ports.<br/>
          <br/>
            <i>Enum</i>: TCP, UDP, SCTP<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.env[index]
<sup><sup>[↩ Parent](#skipjobspeccontainer)</sup></sup>



EnvVar represents an environment variable present in a Container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the environment variable. Must be a C_IDENTIFIER.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Variable references $(VAR_NAME) are expanded
using the previously defined environment variables in the container and
any service environment variables. If a variable cannot be resolved,
the reference in the input string will be unchanged. Double $$ are reduced
to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e.
"$$(VAR_NAME)" will produce the string literal "$(VAR_NAME)".
Escaped references will never be expanded, regardless of whether the variable
exists or not.
Defaults to "".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainerenvindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Source for the environment variable's value. Cannot be used if value is not empty.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.env[index].valueFrom
<sup><sup>[↩ Parent](#skipjobspeccontainerenvindex)</sup></sup>



Source for the environment variable's value. Cannot be used if value is not empty.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#skipjobspeccontainerenvindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainerenvindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`,
spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainerenvindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainerenvindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a secret in the pod's namespace<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.env[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#skipjobspeccontainerenvindexvaluefrom)</sup></sup>



Selects a key of a ConfigMap.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.env[index].valueFrom.fieldRef
<sup><sup>[↩ Parent](#skipjobspeccontainerenvindexvaluefrom)</sup></sup>



Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`,
spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          Path of the field to select in the specified API version.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          Version of the schema the FieldPath is written in terms of, defaults to "v1".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.env[index].valueFrom.resourceFieldRef
<sup><sup>[↩ Parent](#skipjobspeccontainerenvindexvaluefrom)</sup></sup>



Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Required: resource to select<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Container name: required for volumes, optional for env vars<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          Specifies the output format of the exposed resources, defaults to "1"<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.env[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#skipjobspeccontainerenvindexvaluefrom)</sup></sup>



Selects a key of a secret in the pod's namespace

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.envFrom[index]
<sup><sup>[↩ Parent](#skipjobspeccontainer)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>configMap</b></td>
        <td>string</td>
        <td>
          Name of Kubernetes ConfigMap in which the deployment should mount environment variables from. Must be in the same namespace as the Application<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          Name of Kubernetes Secret in which the deployment should mount environment variables from. Must be in the same namespace as the Application<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.filesFrom[index]
<sup><sup>[↩ Parent](#skipjobspeccontainer)</sup></sup>



FilesFrom

Struct representing information needed to mount a Kubernetes resource as a file to a Pod's directory.
One of ConfigMap, Secret, EmptyDir or PersistentVolumeClaim must be present, and just represent the name of the resource in question
NB. Out-of-the-box, skiperator provides a writable 'emptyDir'-volume at '/tmp'

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>mountPath</b></td>
        <td>string</td>
        <td>
          The path to mount the file in the Pods directory. Required.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>configMap</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>emptyDir</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>persistentVolumeClaim</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.gcp
<sup><sup>[↩ Parent](#skipjobspeccontainer)</sup></sup>



GCP

Configuration for interacting with Google Cloud Platform

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#skipjobspeccontainergcpauth">auth</a></b></td>
        <td>object</td>
        <td>
          Configuration for authenticating a Pod with Google Cloud Platform
For authentication with GCP, to use services like Secret Manager and/or Pub/Sub we need
to set the GCP Service Account Pods should identify as. To allow this, we need the IAM role iam.workloadIdentityUser set on a GCP
service account and bind this to the Pod's Kubernetes SA.
Documentation on how this is done can be found here (Closed Wiki):
https://kartverket.atlassian.net/wiki/spaces/SKIPDOK/pages/422346824/Autentisering+mot+GCP+som+Kubernetes+SA<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#skipjobspeccontainergcpcloudsqlproxy">cloudSqlProxy</a></b></td>
        <td>object</td>
        <td>
          CloudSQL is used to deploy a CloudSQL proxy sidecar in the pod.
This is useful for connecting to CloudSQL databases that require Cloud SQL Auth Proxy.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.gcp.auth
<sup><sup>[↩ Parent](#skipjobspeccontainergcp)</sup></sup>



Configuration for authenticating a Pod with Google Cloud Platform
For authentication with GCP, to use services like Secret Manager and/or Pub/Sub we need
to set the GCP Service Account Pods should identify as. To allow this, we need the IAM role iam.workloadIdentityUser set on a GCP
service account and bind this to the Pod's Kubernetes SA.
Documentation on how this is done can be found here (Closed Wiki):
https://kartverket.atlassian.net/wiki/spaces/SKIPDOK/pages/422346824/Autentisering+mot+GCP+som+Kubernetes+SA

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>serviceAccount</b></td>
        <td>string</td>
        <td>
          Name of the service account in which you are trying to authenticate your pod with
Generally takes the form of some-name@some-project-id.iam.gserviceaccount.com<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.gcp.cloudSqlProxy
<sup><sup>[↩ Parent](#skipjobspeccontainergcp)</sup></sup>



CloudSQL is used to deploy a CloudSQL proxy sidecar in the pod.
This is useful for connecting to CloudSQL databases that require Cloud SQL Auth Proxy.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>connectionName</b></td>
        <td>string</td>
        <td>
          Connection name for the CloudSQL instance. Found in the Google Cloud Console under your CloudSQL resource.
The format is "projectName:region:instanceName" E.g. "skip-prod-bda1:europe-north1:my-db".<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>ip</b></td>
        <td>string</td>
        <td>
          The IP address of the CloudSQL instance. This is used to create a serviceentry for the CloudSQL proxy.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>serviceAccount</b></td>
        <td>string</td>
        <td>
          Service account used by cloudsql auth proxy. This service account must have the roles/cloudsql.client role.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>version</b></td>
        <td>string</td>
        <td>
          Image version for the CloudSQL proxy sidecar.<br/>
          <br/>
            <i>Default</i>: 2.8.0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.liveness
<sup><sup>[↩ Parent](#skipjobspeccontainer)</sup></sup>



Probe

Type configuration for all types of Kubernetes probes.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that
are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 10<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.
Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.podSettings
<sup><sup>[↩ Parent](#skipjobspeccontainer)</sup></sup>



PodSettings

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          Annotations that are set on Pods created by Skiperator. These annotations can for example be used to change the behaviour of sidecars and similar.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>disablePodSpreadTopologyConstraints</b></td>
        <td>boolean</td>
        <td>
          DisablePodSpreadTopologyConstraints specifies whether to disable the addition of Pod Topology Spread Constraints to
a given pod.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          TerminationGracePeriodSeconds determines how long Kubernetes waits after a SIGTERM signal sent to a Pod before terminating the pod. If your application uses longer than
30 seconds to terminate, you should increase TerminationGracePeriodSeconds.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Default</i>: 30<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.readiness
<sup><sup>[↩ Parent](#skipjobspeccontainer)</sup></sup>



Probe

Type configuration for all types of Kubernetes probes.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that
are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 10<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.
Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.resources
<sup><sup>[↩ Parent](#skipjobspeccontainer)</sup></sup>



ResourceRequirements

A simplified version of the Kubernetes native ResourceRequirement field, in which only Limits and Requests are present.
For the units used for resources, see https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-units-in-kubernetes

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits set the maximum the app is allowed to use. Exceeding this limit will
make kubernetes kill the app and restart it.

Limits can be set on the CPU and memory, but it is not recommended to put a limit on CPU, see: https://home.robusta.dev/blog/stop-using-cpu-limits<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests set the initial allocation that is done for the app and will
thus be available to the app on startup. More is allocated on demand
until the limit is reached.

Requests can be set on the CPU and memory.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.container.startup
<sup><sup>[↩ Parent](#skipjobspeccontainer)</sup></sup>



Probe

Type configuration for all types of Kubernetes probes.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 3<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that
are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 10<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.
Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.cron
<sup><sup>[↩ Parent](#skipjobspec)</sup></sup>



Settings for the Job if you are running a scheduled job. Optional as Jobs may be one-off.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>schedule</b></td>
        <td>string</td>
        <td>
          A CronJob string for denoting the schedule of this job. See https://crontab.guru/ for help creating CronJob strings.
Kubernetes CronJobs also include the extended "Vixie cron" step values: https://man.freebsd.org/cgi/man.cgi?crontab%285%29.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>allowConcurrency</b></td>
        <td>enum</td>
        <td>
          Denotes how Kubernetes should react to multiple instances of the Job being started at the same time.
Allow will allow concurrent jobs. Forbid will not allow this, and instead skip the newer schedule Job.
Replace will replace the current active Job with the newer scheduled Job.<br/>
          <br/>
            <i>Enum</i>: Allow, Forbid, Replace<br/>
            <i>Default</i>: Allow<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>startingDeadlineSeconds</b></td>
        <td>integer</td>
        <td>
          Denotes the deadline in seconds for starting a job on its schedule, if for some reason the Job's controller was not ready upon the scheduled time.
If unset, Jobs missing their deadline will be considered failed jobs and will not start.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          If set to true, this tells Kubernetes to suspend this Job till the field is set to false. If the Job is active while this field is set to true,
all running Pods will be terminated.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.job
<sup><sup>[↩ Parent](#skipjobspec)</sup></sup>



Settings for the actual Job. If you use a scheduled job, the settings in here will also specify the template of the job.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>activeDeadlineSeconds</b></td>
        <td>integer</td>
        <td>
          ActiveDeadlineSeconds denotes a duration in seconds started from when the job is first active. If the deadline is reached during the job's workload
the job and its Pods are terminated. If the job is suspended using the Suspend field, this timer is stopped and reset when unsuspended.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>backoffLimit</b></td>
        <td>integer</td>
        <td>
          Specifies the number of retry attempts before determining the job as failed. Defaults to 6.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          If set to true, this tells Kubernetes to suspend this Job till the field is set to false. If the Job is active while this field is set to false,
all running Pods will be terminated.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ttlSecondsAfterFinished</b></td>
        <td>integer</td>
        <td>
          The number of seconds to wait before removing the Job after it has finished. If unset, Job will not be cleaned up.
It is recommended to set this to avoid clutter in your resource tree.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.spec.prometheus
<sup><sup>[↩ Parent](#skipjobspec)</sup></sup>



Prometheus settings for pod running in job. Fields are identical to Application and if set,
a podmonitoring object is created.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          The port number or name where metrics are exposed (at the Pod level).<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>allowAllMetrics</b></td>
        <td>boolean</td>
        <td>
          Setting AllowAllMetrics to true will ensure all exposed metrics are scraped. Otherwise, a list of predefined
metrics will be dropped by default. See util/constants.go for the default list.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The HTTP path where Prometheus compatible metrics exists<br/>
          <br/>
            <i>Default</i>: /metrics<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.status
<sup><sup>[↩ Parent](#skipjob)</sup></sup>



SkiperatorStatus

A status field shown on a Skiperator resource which contains information regarding deployment of the resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>accessPolicies</b></td>
        <td>string</td>
        <td>
          Indicates if access policies are valid<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#skipjobstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#skipjobstatussubresourceskey">subresources</a></b></td>
        <td>map[string]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#skipjobstatussummary">summary</a></b></td>
        <td>object</td>
        <td>
          Status<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### SKIPJob.status.conditions[index]
<sup><sup>[↩ Parent](#skipjobstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SKIPJob.status.subresources[key]
<sup><sup>[↩ Parent](#skipjobstatus)</sup></sup>



Status

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: hello<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: Synced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: hello<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### SKIPJob.status.summary
<sup><sup>[↩ Parent](#skipjobstatus)</sup></sup>



Status

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: hello<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: Synced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: hello<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>