# API Reference

## Packages

| Package | Resource types |
| --- | --- |
| `skiperator.kartverket.no/v1alpha1` | [Application](#application)<br/>[Routing](#routing)<br/>[SKIPJob](#skipjob)<br/> |
| `skiperator.kartverket.no/v1beta1` | [SKIPJob](#skipjob-1)<br/> |


## Package `skiperator.kartverket.no/v1alpha1`

Resource types in this package:

- [Application](#application)

- [Routing](#routing)

- [SKIPJob](#skipjob)



<a id="application"></a>
### Application

| Field | Value |
| --- | --- |
| Package | `skiperator.kartverket.no/v1alpha1` |
| API version | `skiperator.kartverket.no/v1alpha1` |
| Kind | `Application` |

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
    <tbody>
      <tr>
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
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspec">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationstatus">status</a></b></td>
        <td>object</td>
        <td>
          ApplicationStatus is a specialized status specific to the Application kind.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspec"></a>
#### Application.spec

<sup>[Parent](#application)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          The image the application will run. This image will be added to a Deployment resource<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          The port the deployment exposes<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecaccesspolicy">accessPolicy</a></b></td>
        <td>object</td>
        <td>
          The root AccessPolicy for managing zero trust access to your Application. See AccessPolicy for more information.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecadditionalportsindex">additionalPorts</a></b></td>
        <td>[]object</td>
        <td>
          An optional list of extra port to expose on a pod level basis,<br/>for example so Instana or other APM tools can reach it<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>appProtocol</b></td>
        <td>enum</td>
        <td>
          Protocol that the application speaks.<br/>
          <br/>
            <i>Enum</i>: http, tcp, udp<br/>
            <i>Default</i>: `http`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecauthorizationsettings">authorizationSettings</a></b></td>
        <td>object</td>
        <td>
          Used for allow listing certain default blocked endpoints, such as /actuator/ end points<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          Override the command set in the Dockerfile. Usually only used when debugging<br/>or running third-party containers where you don&#39;t have control over the Dockerfile<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>enablePDB</b></td>
        <td>boolean</td>
        <td>
          Whether to enable automatic Pod Disruption Budget creation for this application.<br/>
          <br/>
            <i>Default</i>: `true`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecenvindex">env</a></b></td>
        <td>[]object</td>
        <td>
          Environment variables that will be set inside the Deployment&#39;s Pod. See https://pkg.go.dev/k8s.io/api/core/v1#EnvVar for examples.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecenvfromindex">envFrom</a></b></td>
        <td>[]object</td>
        <td>
          Environment variables mounted from files. When specified all the keys of the<br/>resource will be assigned as environment variables. Supports both configmaps<br/>and secrets.<br/><br/>For mounting as files see FilesFrom.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindex">extraContainers</a></b></td>
        <td>[]object</td>
        <td>
          Extra containers to run in the pod alongside the main application<br/>container. Each entry is either a regular container (type: standard, the<br/>default) running next to the main container, or an init container<br/>(type: init) that starts first and stays running for the pod lifetime<br/>(a native sidecar). The operator enforces a least-privilege security<br/>context on these containers; it cannot be overridden.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecfilesfromindex">filesFrom</a></b></td>
        <td>[]object</td>
        <td>
          Mounting volumes into the Deployment are done using the FilesFrom argument<br/><br/>FilesFrom supports ConfigMaps, Secrets and PVCs. The Application resource<br/>assumes these have already been created by you, and will fail if this is not the case.<br/><br/>For mounting environment variables see EnvFrom.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecgcp">gcp</a></b></td>
        <td>object</td>
        <td>
          GCP is used to configure Google Cloud Platform specific settings for the application.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecidporten">idporten</a></b></td>
        <td>object</td>
        <td>
          Settings for IDPorten integration with Digitaliseringsdirektoratet<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>ingresses</b></td>
        <td>[]string</td>
        <td>
          Any external hostnames that route to this application. Using a skip.statkart.no-address<br/>will make the application reachable for kartverket-clients (internal), other addresses<br/>make the app reachable on the internet. Note that other addresses than skip.statkart.no<br/>(also known as pretty hostnames) requires additional DNS setup.<br/>The below hostnames will also have TLS certificates issued and be reachable on both<br/>HTTP and HTTPS.<br/><br/>Ingresses must be lowercase, contain no spaces, be a non-empty string, and have a hostname/domain separated by a period<br/>They can optionally be suffixed with a plus and name of a custom TLS secret located in the istio-gateways namespace.<br/>E.g. &#34;foo.atkv3-dev.kartverket-intern.cloud+env-wildcard-cert&#34;<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecistiosettings">istioSettings</a></b></td>
        <td>object</td>
        <td>
          IstioSettings are used to configure istio specific resources such as telemetry. Currently, adjusting sampling<br/>interval for tracing is the only supported option.<br/>By default, tracing is enabled with a random sampling percentage of 10%.<br/>
          <br/>
            <i>Default</i>: `map[telemetry:map[tracing:[map[randomSamplingPercentage:10]]]]`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          Labels can be used if you want every resource created by your application to<br/>have the same labels, including your application. This could for example be useful for<br/>metrics, where a certain label and the corresponding resources liveliness can be combined.<br/>Any amount of labels can be added as wanted, and they will all cascade down to all resources.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecliveness">liveness</a></b></td>
        <td>object</td>
        <td>
          Liveness probes define a resource that returns 200 OK when the app is running<br/>as intended. Returning a non-200 code will make kubernetes restart the app.<br/>Liveness is optional, but when provided, path and port are required<br/><br/>See Probe for structure definition.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecmaskinporten">maskinporten</a></b></td>
        <td>object</td>
        <td>
          Settings for Maskinporten integration with Digitaliseringsdirektoratet<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecpodsettings">podSettings</a></b></td>
        <td>object</td>
        <td>
          PodSettings are used to apply specific settings to the Pod Template used by Skiperator to create Deployments. This allows you to set<br/>things like annotations on the Pod to change the behaviour of sidecars, and set relevant Pod options such as TerminationGracePeriodSeconds.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>priority</b></td>
        <td>enum</td>
        <td>
          An optional priority. Supported values are &#39;low&#39;, &#39;medium&#39; and &#39;high&#39;.<br/>The default value is &#39;medium&#39;.<br/><br/>Most workloads should not have to specify this field. If you think you<br/>do, please consult with SKIP beforehand.<br/>
          <br/>
            <i>Enum</i>: low, medium, high<br/>
            <i>Default</i>: `medium`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecprometheus">prometheus</a></b></td>
        <td>object</td>
        <td>
          Optional settings for how Prometheus compatible metrics should be scraped.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecreadiness">readiness</a></b></td>
        <td>object</td>
        <td>
          Readiness probes define a resource that returns 200 OK when the app is running<br/>as intended. Kubernetes will wait until the resource returns 200 OK before<br/>marking the pod as Running and progressing with the deployment strategy.<br/>Readiness is optional, but when provided, path and port are required<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>redirectToHTTPS</b></td>
        <td>boolean</td>
        <td>
          Controls whether the application will automatically redirect all HTTP calls to HTTPS via the istio VirtualService.<br/>This redirect does not happen on the route /.well-known/acme-challenge/, as the ACME challenge can only be done on port 80.<br/>
          <br/>
            <i>Default</i>: `true`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>replicas</b></td>
        <td>JSON</td>
        <td>
          The number of replicas can either be specified as a static number as follows:<br/><br/>	replicas: 2<br/><br/>Or by specifying a range between min and max to enable HorizontalPodAutoscaling.<br/>The default value for replicas is:<br/>	replicas:<br/>		min: 2<br/>		max: 5<br/>		targetCpuUtilization: 80<br/>     targetMemoryUtilization: 80<br/>Using autoscaling is the recommended configuration for replicas.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>resourceLabels</b></td>
        <td>map[string]map[string]string</td>
        <td>
          ResourceLabels can be used if you want to add a label to a specific resources created by<br/>the application. One such label could for example be set on a Deployment, such that<br/>the deployment avoids certain rules from Gatekeeper, or similar. Any amount of labels may be added per ResourceLabels item.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecresources">resources</a></b></td>
        <td>object</td>
        <td>
          ResourceRequirements to apply to the deployment. It&#39;s common to set some of these to<br/>prevent the app from swelling in resource usage and consuming all the<br/>resources of other apps on the cluster.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecstartup">startup</a></b></td>
        <td>object</td>
        <td>
          Kubernetes uses startup probes to know when a container application has started.<br/>If such a probe is configured, it disables liveness and readiness checks until it<br/>succeeds, making sure those probes don&#39;t interfere with the application startup.<br/>This can be used to adopt liveness checks on slow starting containers, avoiding them<br/>getting killed by Kubernetes before they are up and running.<br/>Startup is optional, but when provided, path and port are required<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecstateful">stateful</a></b></td>
        <td>object</td>
        <td>
          Stateful, when set with enabled=true, generates a StatefulSet instead of a Deployment.<br/>Requires VolumeClaimTemplates. Disallows Strategy.Type=Recreate and HPA-range replicas.<br/>The enabled flag is immutable - delete and recreate the Application to change.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecstrategy">strategy</a></b></td>
        <td>object</td>
        <td>
          Defines an alternative strategy for the Kubernetes deployment. This is useful when<br/>the default strategy, RollingUpdate, is not usable. Setting type to<br/>Recreate will take down all the pods before starting new pods, whereas the<br/>default of RollingUpdate will try to start the new pods before taking down the<br/>old ones.<br/><br/>Valid values are: RollingUpdate, Recreate. Default is RollingUpdate<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>team</b></td>
        <td>string</td>
        <td>
          Team specifies the team who owns this particular app.<br/>Usually sourced from the namespace label.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecaccesspolicy"></a>
#### Application.spec.accessPolicy

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#applicationspecaccesspolicyinbound">inbound</a></b></td>
        <td>object</td>
        <td>
          Inbound specifies the ingress rules. Which apps on the cluster can talk to this app?<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecaccesspolicyoutbound">outbound</a></b></td>
        <td>object</td>
        <td>
          Outbound specifies egress rules. Which apps on the cluster and the<br/>internet is the Application allowed to send requests to?<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecaccesspolicyinbound"></a>
#### Application.spec.accessPolicy.inbound

<sup>[Parent](#applicationspecaccesspolicy)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#applicationspecaccesspolicyinboundrulesindex">rules</a></b></td>
        <td>[]object</td>
        <td>
          The rules list specifies a list of applications. When no namespace is<br/>specified it refers to an app in the current namespace. For apps in<br/>other namespaces namespace is required<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecaccesspolicyinboundrulesindex"></a>
#### Application.spec.accessPolicy.inbound.rules[index]

<sup>[Parent](#applicationspecaccesspolicyinbound)</sup>

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
    <tbody>
      <tr>
        <td><b>application</b></td>
        <td>string</td>
        <td>
          The name of the Application you are allowing traffic to/from. If you wish to allow traffic from a SKIPJob, this field should<br/>be suffixed with -skipjob<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          The namespace in which the Application you are allowing traffic to/from resides. If unset, uses namespace of Application.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>namespacesByLabel</b></td>
        <td>map[string]string</td>
        <td>
          Namespace label value-pair in which the Application you are allowing traffic to/from resides. If both namespace and namespacesByLabel are set, namespace takes precedence and namespacesByLabel is omitted.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecaccesspolicyinboundrulesindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above application.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecaccesspolicyinboundrulesindexportsindex"></a>
#### Application.spec.accessPolicy.inbound.rules[index].ports[index]

<sup>[Parent](#applicationspecaccesspolicyinboundrulesindex)</sup>

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
    <tbody>
      <tr>
        <td><b>endPort</b></td>
        <td>integer</td>
        <td>
          endPort indicates that the range of ports from port to endPort if set, inclusive,<br/>should be allowed by the policy. This field cannot be defined if the port field<br/>is not defined or if the port field is defined as a named (string) port.<br/>The endPort must be equal or greater than port.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          port represents the port on the given protocol. This can either be a numerical or named<br/>port on a pod. If this field is not provided, this matches all port names and<br/>numbers.<br/>If present, only traffic on the specified protocol AND port will be matched.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          protocol represents the protocol (TCP, UDP, or SCTP) which traffic must match.<br/>If not specified, this field defaults to TCP.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecaccesspolicyoutbound"></a>
#### Application.spec.accessPolicy.outbound

<sup>[Parent](#applicationspecaccesspolicy)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#applicationspecaccesspolicyoutboundexternalindex">external</a></b></td>
        <td>[]object</td>
        <td>
          External specifies which applications on the internet the application<br/>can reach. Only host is required unless it is on another port than HTTPS port 443.<br/>If other ports or protocols are required then `ports` must be specified as well<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecaccesspolicyoutboundrulesindex">rules</a></b></td>
        <td>[]object</td>
        <td>
          Rules apply the same in-cluster rules as InboundPolicy<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecaccesspolicyoutboundexternalindex"></a>
#### Application.spec.accessPolicy.outbound.external[index]

<sup>[Parent](#applicationspecaccesspolicyoutbound)</sup>

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
    <tbody>
      <tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          The allowed hostname. Note that this does not include subdomains.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>ip</b></td>
        <td>string</td>
        <td>
          Non-HTTP requests (i.e. using the TCP protocol) need to use IP in addition to hostname<br/>Only required for TCP requests.<br/><br/>Note: Hostname must always be defined even if IP is set statically<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecaccesspolicyoutboundexternalindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above hostname. When not specified HTTP and<br/>HTTPS on port 80 and 443 respectively are put into the allowlist<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecaccesspolicyoutboundexternalindexportsindex"></a>
#### Application.spec.accessPolicy.outbound.external[index].ports[index]

<sup>[Parent](#applicationspecaccesspolicyoutboundexternalindex)</sup>

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
    <tbody>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is required and is an arbitrary name. Must be unique within all ExternalRule ports.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          The port number of the external host<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>protocol</b></td>
        <td>enum</td>
        <td>
          The protocol to use for communication with the host. Supported protocols are: HTTP, HTTPS, TCP and TLS.<br/>
          <br/>
            <i>Enum</i>: HTTP, HTTPS, TCP, TLS<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecaccesspolicyoutboundrulesindex"></a>
#### Application.spec.accessPolicy.outbound.rules[index]

<sup>[Parent](#applicationspecaccesspolicyoutbound)</sup>

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
    <tbody>
      <tr>
        <td><b>application</b></td>
        <td>string</td>
        <td>
          The name of the Application you are allowing traffic to/from. If you wish to allow traffic from a SKIPJob, this field should<br/>be suffixed with -skipjob<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          The namespace in which the Application you are allowing traffic to/from resides. If unset, uses namespace of Application.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>namespacesByLabel</b></td>
        <td>map[string]string</td>
        <td>
          Namespace label value-pair in which the Application you are allowing traffic to/from resides. If both namespace and namespacesByLabel are set, namespace takes precedence and namespacesByLabel is omitted.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecaccesspolicyoutboundrulesindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above application.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecaccesspolicyoutboundrulesindexportsindex"></a>
#### Application.spec.accessPolicy.outbound.rules[index].ports[index]

<sup>[Parent](#applicationspecaccesspolicyoutboundrulesindex)</sup>

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
    <tbody>
      <tr>
        <td><b>endPort</b></td>
        <td>integer</td>
        <td>
          endPort indicates that the range of ports from port to endPort if set, inclusive,<br/>should be allowed by the policy. This field cannot be defined if the port field<br/>is not defined or if the port field is defined as a named (string) port.<br/>The endPort must be equal or greater than port.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          port represents the port on the given protocol. This can either be a numerical or named<br/>port on a pod. If this field is not provided, this matches all port names and<br/>numbers.<br/>If present, only traffic on the specified protocol AND port will be matched.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          protocol represents the protocol (TCP, UDP, or SCTP) which traffic must match.<br/>If not specified, this field defaults to TCP.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecadditionalportsindex"></a>
#### Application.spec.additionalPorts[index]

<sup>[Parent](#applicationspec)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>protocol</b></td>
        <td>enum</td>
        <td>
          Protocol defines network protocols supported for things like container ports.<br/>
          <br/>
            <i>Enum</i>: TCP, UDP, SCTP<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecauthorizationsettings"></a>
#### Application.spec.authorizationSettings

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b>allowAll</b></td>
        <td>boolean</td>
        <td>
          Allows all endpoints by not creating an AuthorizationPolicy, and ignores the content of AllowList.<br/>If field is false, the contents of AllowList will be used instead if AllowList is set.<br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>allowList</b></td>
        <td>[]string</td>
        <td>
          Allows specific endpoints. Common endpoints one might want to allow include /actuator/health, /actuator/startup, /actuator/info.<br/><br/>Note that endpoints are matched specifically on the input, so if you allow /actuator/health, you will *not* allow /actuator/health/<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecenvindex"></a>
#### Application.spec.env[index]

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the environment variable.<br/>May consist of any printable ASCII characters except &#39;=&#39;.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Variable references $(VAR_NAME) are expanded<br/>using the previously defined environment variables in the container and<br/>any service environment variables. If a variable cannot be resolved,<br/>the reference in the input string will be unchanged. Double $$ are reduced<br/>to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e.<br/>&#34;$$(VAR_NAME)&#34; will produce the string literal &#34;$(VAR_NAME)&#34;.<br/>Escaped references will never be expanded, regardless of whether the variable<br/>exists or not.<br/>Defaults to &#34;&#34;.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecenvindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Source for the environment variable&#39;s value. Cannot be used if value is not empty.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecenvindexvaluefrom"></a>
#### Application.spec.env[index].valueFrom

<sup>[Parent](#applicationspecenvindex)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#applicationspecenvindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecenvindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels[&#39;&lt;KEY&gt;&#39;]`, `metadata.annotations[&#39;&lt;KEY&gt;&#39;]`,<br/>spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecenvindexvaluefromfilekeyref">fileKeyRef</a></b></td>
        <td>object</td>
        <td>
          FileKeyRef selects a key of the env file.<br/>Requires the EnvFiles feature gate to be enabled.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecenvindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests<br/>(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecenvindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a secret in the pod&#39;s namespace<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecenvindexvaluefromconfigmapkeyref"></a>
#### Application.spec.env[index].valueFrom.configMapKeyRef

<sup>[Parent](#applicationspecenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.<br/>This field is effectively required, but due to backwards compatibility is<br/>allowed to be empty. Instances of this type with an empty value here are<br/>almost certainly wrong.<br/>More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: ``<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecenvindexvaluefromfieldref"></a>
#### Application.spec.env[index].valueFrom.fieldRef

<sup>[Parent](#applicationspecenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          Path of the field to select in the specified API version.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          Version of the schema the FieldPath is written in terms of, defaults to &#34;v1&#34;.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecenvindexvaluefromfilekeyref"></a>
#### Application.spec.env[index].valueFrom.fileKeyRef

<sup>[Parent](#applicationspecenvindexvaluefrom)</sup>

FileKeyRef selects a key of the env file.
Requires the EnvFiles feature gate to be enabled.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key within the env file. An invalid key will prevent the pod from starting.<br/>The keys defined within a source may consist of any printable ASCII characters except &#39;=&#39;.<br/>During Alpha stage of the EnvFiles feature gate, the key size is limited to 128 characters.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path within the volume from which to select the file.<br/>Must be relative and may not contain the &#39;..&#39; path or start with &#39;..&#39;.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          The name of the volume mount containing the env file.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the file or its key must be defined. If the file or key<br/>does not exist, then the env var is not published.<br/>If optional is set to true and the specified key does not exist,<br/>the environment variable will not be set in the Pod&#39;s containers.<br/><br/>If optional is set to false and the specified key does not exist,<br/>an error will be returned during Pod creation.<br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecenvindexvaluefromresourcefieldref"></a>
#### Application.spec.env[index].valueFrom.resourceFieldRef

<sup>[Parent](#applicationspecenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Required: resource to select<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Container name: required for volumes, optional for env vars<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          Specifies the output format of the exposed resources, defaults to &#34;1&#34;<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecenvindexvaluefromsecretkeyref"></a>
#### Application.spec.env[index].valueFrom.secretKeyRef

<sup>[Parent](#applicationspecenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.<br/>This field is effectively required, but due to backwards compatibility is<br/>allowed to be empty. Instances of this type with an empty value here are<br/>almost certainly wrong.<br/>More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: ``<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecenvfromindex"></a>
#### Application.spec.envFrom[index]

<sup>[Parent](#applicationspec)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>configMap</b></td>
        <td>string</td>
        <td>
          Name of Kubernetes ConfigMap in which the deployment should mount environment variables from. Must be in the same namespace as the Application<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          Name of Kubernetes Secret in which the deployment should mount environment variables from. Must be in the same namespace as the Application<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindex"></a>
#### Application.spec.extraContainers[index]

<sup>[Parent](#applicationspec)</sup>

ContainerSpec describes an extra container to run in the workload's pod
alongside the main application container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          The container image to run.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the container. Must be unique within the pod and must not collide<br/>with the application name or a reserved name (e.g. cloudsql-proxy,<br/>istio-proxy, istio-validation, istio-init).<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexadditionalportsindex">additionalPorts</a></b></td>
        <td>[]object</td>
        <td>
          Additional ports exposed by the container.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>args</b></td>
        <td>[]string</td>
        <td>
          Arguments to the container entrypoint.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          Override the command set in the image.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexenvindex">env</a></b></td>
        <td>[]object</td>
        <td>
          Environment variables set inside the container.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexenvfromindex">envFrom</a></b></td>
        <td>[]object</td>
        <td>
          Environment variables mounted from ConfigMaps or Secrets. When specified<br/>all keys of the resource are assigned as environment variables.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexfilesfromindex">filesFrom</a></b></td>
        <td>[]object</td>
        <td>
          Files mounted into the container from ConfigMaps, Secrets, PVCs or<br/>emptyDirs. The referenced resources are assumed to already exist.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>ingressPort</b></td>
        <td>integer</td>
        <td>
          When set, the application&#39;s ingress traffic enters the pod through this<br/>container instead of the main container: the generated Service keeps its<br/>external port (spec.port) but routes its target port to this container&#39;s<br/>IngressPort. This suits any container that should sit in front of the<br/>application and receive incoming traffic first - an auth proxy, an API<br/>gateway, a TLS-terminating or rate-limiting proxy, etc. — which then<br/>forwards to the application (e.g. it listens on ingressPort and forwards<br/>to the app on spec.port via localhost).<br/><br/>The IngressPort value must be declared in this container&#39;s additionalPorts.<br/>At most one extra container may set this, and the value must differ from<br/>spec.port.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexliveness">liveness</a></b></td>
        <td>object</td>
        <td>
          Liveness probe. When provided, path and port are required.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexreadiness">readiness</a></b></td>
        <td>object</td>
        <td>
          Readiness probe. When provided, path and port are required.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexresources">resources</a></b></td>
        <td>object</td>
        <td>
          ResourceRequirements to apply to the container.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexstartup">startup</a></b></td>
        <td>object</td>
        <td>
          Startup probe. When provided, path and port are required.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type selects how the container runs:<br/>  - &#34;standard&#34; or omitted: a regular container running alongside the main<br/>    container for the lifetime of the pod.<br/>  - &#34;init&#34;: an init container that starts before the main container and<br/>    keeps running for the lifetime of the pod.<br/>
          <br/>
            <i>Enum</i>: standard, init<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexadditionalportsindex"></a>
#### Application.spec.extraContainers[index].additionalPorts[index]

<sup>[Parent](#applicationspecextracontainersindex)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>protocol</b></td>
        <td>enum</td>
        <td>
          Protocol defines network protocols supported for things like container ports.<br/>
          <br/>
            <i>Enum</i>: TCP, UDP, SCTP<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexenvindex"></a>
#### Application.spec.extraContainers[index].env[index]

<sup>[Parent](#applicationspecextracontainersindex)</sup>

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
    <tbody>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the environment variable.<br/>May consist of any printable ASCII characters except &#39;=&#39;.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Variable references $(VAR_NAME) are expanded<br/>using the previously defined environment variables in the container and<br/>any service environment variables. If a variable cannot be resolved,<br/>the reference in the input string will be unchanged. Double $$ are reduced<br/>to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e.<br/>&#34;$$(VAR_NAME)&#34; will produce the string literal &#34;$(VAR_NAME)&#34;.<br/>Escaped references will never be expanded, regardless of whether the variable<br/>exists or not.<br/>Defaults to &#34;&#34;.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexenvindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Source for the environment variable&#39;s value. Cannot be used if value is not empty.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexenvindexvaluefrom"></a>
#### Application.spec.extraContainers[index].env[index].valueFrom

<sup>[Parent](#applicationspecextracontainersindexenvindex)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexenvindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexenvindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels[&#39;&lt;KEY&gt;&#39;]`, `metadata.annotations[&#39;&lt;KEY&gt;&#39;]`,<br/>spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexenvindexvaluefromfilekeyref">fileKeyRef</a></b></td>
        <td>object</td>
        <td>
          FileKeyRef selects a key of the env file.<br/>Requires the EnvFiles feature gate to be enabled.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexenvindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests<br/>(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecextracontainersindexenvindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a secret in the pod&#39;s namespace<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexenvindexvaluefromconfigmapkeyref"></a>
#### Application.spec.extraContainers[index].env[index].valueFrom.configMapKeyRef

<sup>[Parent](#applicationspecextracontainersindexenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.<br/>This field is effectively required, but due to backwards compatibility is<br/>allowed to be empty. Instances of this type with an empty value here are<br/>almost certainly wrong.<br/>More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: ``<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexenvindexvaluefromfieldref"></a>
#### Application.spec.extraContainers[index].env[index].valueFrom.fieldRef

<sup>[Parent](#applicationspecextracontainersindexenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          Path of the field to select in the specified API version.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          Version of the schema the FieldPath is written in terms of, defaults to &#34;v1&#34;.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexenvindexvaluefromfilekeyref"></a>
#### Application.spec.extraContainers[index].env[index].valueFrom.fileKeyRef

<sup>[Parent](#applicationspecextracontainersindexenvindexvaluefrom)</sup>

FileKeyRef selects a key of the env file.
Requires the EnvFiles feature gate to be enabled.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key within the env file. An invalid key will prevent the pod from starting.<br/>The keys defined within a source may consist of any printable ASCII characters except &#39;=&#39;.<br/>During Alpha stage of the EnvFiles feature gate, the key size is limited to 128 characters.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path within the volume from which to select the file.<br/>Must be relative and may not contain the &#39;..&#39; path or start with &#39;..&#39;.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          The name of the volume mount containing the env file.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the file or its key must be defined. If the file or key<br/>does not exist, then the env var is not published.<br/>If optional is set to true and the specified key does not exist,<br/>the environment variable will not be set in the Pod&#39;s containers.<br/><br/>If optional is set to false and the specified key does not exist,<br/>an error will be returned during Pod creation.<br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexenvindexvaluefromresourcefieldref"></a>
#### Application.spec.extraContainers[index].env[index].valueFrom.resourceFieldRef

<sup>[Parent](#applicationspecextracontainersindexenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Required: resource to select<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Container name: required for volumes, optional for env vars<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          Specifies the output format of the exposed resources, defaults to &#34;1&#34;<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexenvindexvaluefromsecretkeyref"></a>
#### Application.spec.extraContainers[index].env[index].valueFrom.secretKeyRef

<sup>[Parent](#applicationspecextracontainersindexenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.<br/>This field is effectively required, but due to backwards compatibility is<br/>allowed to be empty. Instances of this type with an empty value here are<br/>almost certainly wrong.<br/>More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: ``<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexenvfromindex"></a>
#### Application.spec.extraContainers[index].envFrom[index]

<sup>[Parent](#applicationspecextracontainersindex)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>configMap</b></td>
        <td>string</td>
        <td>
          Name of Kubernetes ConfigMap in which the deployment should mount environment variables from. Must be in the same namespace as the Application<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          Name of Kubernetes Secret in which the deployment should mount environment variables from. Must be in the same namespace as the Application<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexfilesfromindex"></a>
#### Application.spec.extraContainers[index].filesFrom[index]

<sup>[Parent](#applicationspecextracontainersindex)</sup>

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
    <tbody>
      <tr>
        <td><b>mountPath</b></td>
        <td>string</td>
        <td>
          The path to mount the file in the Pods directory. Required.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>configMap</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          defaultMode is optional: mode bits used to set permissions on created files by default.<br/>Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.<br/>YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.<br/>Defaults to 0644.<br/>Directories within the path are not affected by this setting.<br/>This might be in conflict with other options that affect the file<br/>mode, like fsGroup, and the result can be other mode bits set.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>emptyDir</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>persistentVolumeClaim</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexliveness"></a>
#### Application.spec.extraContainers[index].liveness

<sup>[Parent](#applicationspecextracontainersindex)</sup>

Liveness probe. When provided, path and port are required.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after<br/>having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `3`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that<br/>are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `0`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `10`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.<br/>Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.<br/>Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexreadiness"></a>
#### Application.spec.extraContainers[index].readiness

<sup>[Parent](#applicationspecextracontainersindex)</sup>

Readiness probe. When provided, path and port are required.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after<br/>having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `3`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that<br/>are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `0`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `10`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.<br/>Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.<br/>Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexresources"></a>
#### Application.spec.extraContainers[index].resources

<sup>[Parent](#applicationspecextracontainersindex)</sup>

ResourceRequirements to apply to the container.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits set the maximum the app is allowed to use. Exceeding this limit will<br/>make kubernetes kill the app and restart it.<br/><br/>Limits can be set on the CPU and memory, but it is not recommended to put a limit on CPU, see: https://home.robusta.dev/blog/stop-using-cpu-limits<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests set the initial allocation that is done for the app and will<br/>thus be available to the app on startup. More is allocated on demand<br/>until the limit is reached.<br/><br/>Requests can be set on the CPU and memory.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecextracontainersindexstartup"></a>
#### Application.spec.extraContainers[index].startup

<sup>[Parent](#applicationspecextracontainersindex)</sup>

Startup probe. When provided, path and port are required.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after<br/>having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `3`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that<br/>are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `0`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `10`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.<br/>Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.<br/>Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecfilesfromindex"></a>
#### Application.spec.filesFrom[index]

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b>mountPath</b></td>
        <td>string</td>
        <td>
          The path to mount the file in the Pods directory. Required.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>configMap</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          defaultMode is optional: mode bits used to set permissions on created files by default.<br/>Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.<br/>YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.<br/>Defaults to 0644.<br/>Directories within the path are not affected by this setting.<br/>This might be in conflict with other options that affect the file<br/>mode, like fsGroup, and the result can be other mode bits set.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>emptyDir</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>persistentVolumeClaim</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecgcp"></a>
#### Application.spec.gcp

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#applicationspecgcpauth">auth</a></b></td>
        <td>object</td>
        <td>
          Configuration for authenticating a Pod with Google Cloud Platform<br/>For authentication with GCP, to use services like Secret Manager and/or Pub/Sub we need<br/>to set the GCP Service Account Pods should identify as. To allow this, we need the IAM role iam.workloadIdentityUser set on a GCP<br/>service account and bind this to the Pod&#39;s Kubernetes SA.<br/>Documentation on how this is done can be found here (Closed Wiki):<br/>https://kartverket.atlassian.net/wiki/spaces/SKIPDOK/pages/422346824/Autentisering+mot+GCP+som+Kubernetes+SA<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecgcpcloudsqlproxy">cloudSqlProxy</a></b></td>
        <td>object</td>
        <td>
          CloudSQL is used to deploy a CloudSQL proxy sidecar in the pod.<br/>This is useful for connecting to CloudSQL databases that require Cloud SQL Auth Proxy.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecgcpauth"></a>
#### Application.spec.gcp.auth

<sup>[Parent](#applicationspecgcp)</sup>

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
    <tbody>
      <tr>
        <td><b>serviceAccount</b></td>
        <td>string</td>
        <td>
          Name of the service account in which you are trying to authenticate your pod with<br/>Generally takes the form of some-name@some-project-id.iam.gserviceaccount.com<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecgcpcloudsqlproxy"></a>
#### Application.spec.gcp.cloudSqlProxy

<sup>[Parent](#applicationspecgcp)</sup>

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
    <tbody>
      <tr>
        <td><b>connectionName</b></td>
        <td>string</td>
        <td>
          Connection name for the CloudSQL instance. Found in the Google Cloud Console under your CloudSQL resource.<br/>The format is &#34;projectName:region:instanceName&#34; E.g. &#34;skip-prod-bda1:europe-north1:my-db&#34;.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>ip</b></td>
        <td>string</td>
        <td>
          The IP address of the CloudSQL instance. This is used to create a serviceentry for the CloudSQL proxy.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>serviceAccount</b></td>
        <td>string</td>
        <td>
          Service account used by cloudsql auth proxy. This service account must have the roles/cloudsql.client role.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>publicIP</b></td>
        <td>boolean</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>version</b></td>
        <td>string</td>
        <td>
          Image version for the CloudSQL proxy sidecar.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecidporten"></a>
#### Application.spec.idporten

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          Whether to enable provisioning of an ID-porten client.<br/>If enabled, an ID-porten client will be provisioned.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>accessTokenLifetime</b></td>
        <td>integer</td>
        <td>
          AccessTokenLifetime is the lifetime in seconds for any issued access token from ID-porten.<br/><br/>If unspecified, defaults to `3600` seconds (1 hour).<br/>
          <br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 3600<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>clientName</b></td>
        <td>string</td>
        <td>
          The name of the Client as shown in Digitaliseringsdirektoratet&#39;s Samarbeidsportal<br/>Meant to be a human-readable name for separating clients in the portal.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>clientURI</b></td>
        <td>string</td>
        <td>
          ClientURI is the URL shown to the user at ID-porten when displaying a &#39;back&#39; button or on errors.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>frontchannelLogoutPath</b></td>
        <td>string</td>
        <td>
          FrontchannelLogoutPath is a valid path for your application where ID-porten sends a request to whenever the user has<br/>initiated a logout elsewhere as part of a single logout (front channel logout) process.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>integrationType</b></td>
        <td>enum</td>
        <td>
          IntegrationType is used to make sensible choices for your client.<br/>Which type of integration you choose will provide guidance on which scopes you can use with the client.<br/>A client can only have one integration type.<br/><br/>NB! It is not possible to change the integration type after creation.<br/>
          <br/>
            <i>Enum</i>: krr, idporten, api_klient<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>postLogoutRedirectPath</b></td>
        <td>string</td>
        <td>
          PostLogoutRedirectPath is a simpler verison of PostLogoutRedirectURIs<br/>that will be appended to the ingress<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>postLogoutRedirectURIs</b></td>
        <td>[]string</td>
        <td>
          PostLogoutRedirectURIs are valid URIs that ID-porten will allow redirecting the end-user to after a single logout<br/>has been initiated and performed by the application.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>redirectPath</b></td>
        <td>string</td>
        <td>
          RedirectPath is a valid path that ID-porten redirects back to after a successful authorization request.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecidportenrequestauthentication">requestAuthentication</a></b></td>
        <td>object</td>
        <td>
          RequestAuthentication specifies how incoming JWTs should be validated.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>scopes</b></td>
        <td>[]string</td>
        <td>
          Register different oauth2 Scopes on your client.<br/>You will not be able to add a scope to your client that conflicts with the client&#39;s IntegrationType.<br/>For example, you can not add a scope that is limited to the IntegrationType `krr` of IntegrationType `idporten`, and vice versa.<br/><br/>Default for IntegrationType `krr` = (&#34;krr:global/kontaktinformasjon.read&#34;, &#34;krr:global/digitalpost.read&#34;)<br/>Default for IntegrationType `idporten` = (&#34;openid&#34;, &#34;profile&#34;)<br/>IntegrationType `api_klient` have no Default, checkout Digdir documentation.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>sessionLifetime</b></td>
        <td>integer</td>
        <td>
          SessionLifetime is the maximum lifetime in seconds for any given user&#39;s session in your application.<br/>The timeout starts whenever the user is redirected from the `authorization_endpoint` at ID-porten.<br/><br/>If unspecified, defaults to `7200` seconds (2 hours).<br/>Note: Attempting to refresh the user&#39;s `access_token` beyond this timeout will yield an error.<br/>
          <br/>
            <i>Minimum</i>: 3600<br/>
            <i>Maximum</i>: 7200<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecidportenrequestauthentication"></a>
#### Application.spec.idporten.requestAuthentication

<sup>[Parent](#applicationspecidporten)</sup>

RequestAuthentication specifies how incoming JWTs should be validated.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          Whether to enable JWT validation.<br/>If enabled, incoming JWTs will be validated against the issuer specified in the app registration and the generated audience.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>forwardJwt</b></td>
        <td>boolean</td>
        <td>
          If set to `true`, the original token will be kept for the upstream request. Defaults to `true`.<br/>
          <br/>
            <i>Default</i>: `true`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>ignorePaths</b></td>
        <td>[]string</td>
        <td>
          IgnorePaths specifies paths that do not require an authenticated JWT.<br/><br/>The specified paths must be a valid URI path. It has to start with &#39;/&#39; and cannot end with &#39;/&#39;.<br/>The paths can also contain the wildcard operator &#39;*&#39;, but only at the end.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecidportenrequestauthenticationoutputclaimtoheadersindex">outputClaimToHeaders</a></b></td>
        <td>[]object</td>
        <td>
          This field specifies a list of operations to copy the claim to HTTP headers on a successfully verified token.<br/>The header specified in each operation in the list must be unique. Nested claims of type string/int/bool is supported as well.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>paths</b></td>
        <td>[]string</td>
        <td>
          Paths specifies paths that require an authenticated JWT.<br/><br/>The specified paths must be a valid URI path. It has to start with &#39;/&#39; and cannot end with &#39;/&#39;.<br/>The paths can also contain the wildcard operator &#39;*&#39;, but only at the end.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>secretName</b></td>
        <td>string</td>
        <td>
          The name of the Kubernetes Secret containing OAuth2 credentials.<br/><br/>If omitted, the associated client registration in the application manifest is used for JWT validation.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>tokenLocation</b></td>
        <td>enum</td>
        <td>
          Where to find the JWT in the incoming request<br/><br/>An enum value of `header` means that the JWT is present in the `Authorization` header as a `Bearer` token.<br/>An enum value of `cookie` means that the JWT is present as a cookie called `BearerToken`.<br/><br/>If omitted, its default value depends on the provider type:<br/>  Defaults to &#34;cookie&#34; for providers supporting user login (e.g. IDPorten).<br/>  Defaults to &#34;header&#34; for providers not supporting user login (e.g. Maskinporten).<br/>
          <br/>
            <i>Enum</i>: header, cookie<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecidportenrequestauthenticationoutputclaimtoheadersindex"></a>
#### Application.spec.idporten.requestAuthentication.outputClaimToHeaders[index]

<sup>[Parent](#applicationspecidportenrequestauthentication)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>claim</b></td>
        <td>string</td>
        <td>
          The claim to be copied.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>header</b></td>
        <td>string</td>
        <td>
          The name of the HTTP header for which the specified claim will be copied to.<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecistiosettings"></a>
#### Application.spec.istioSettings

<sup>[Parent](#applicationspec)</sup>

IstioSettings are used to configure istio specific resources such as telemetry. Currently, adjusting sampling
interval for tracing is the only supported option.
By default, tracing is enabled with a random sampling percentage of 10%.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b><a href="#applicationspecistiosettingsretries">retries</a></b></td>
        <td>object</td>
        <td>
          Retries is configurable automatic retries for requests towards the application.<br/>By default requests falling under: &#34;connect-failure,refused-stream,unavailable,cancelled&#34; will be retried.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecistiosettingstelemetry">telemetry</a></b></td>
        <td>object</td>
        <td>
          Telemetry is a placeholder for all relevant telemetry types, and may be extended in the future to configure additional telemetry settings.<br/>
          <br/>
            <i>Default</i>: `map[tracing:[map[randomSamplingPercentage:10]]]`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecistiosettingsretries"></a>
#### Application.spec.istioSettings.retries

<sup>[Parent](#applicationspecistiosettings)</sup>

Retries is configurable automatic retries for requests towards the application.
By default requests falling under: "connect-failure,refused-stream,unavailable,cancelled" will be retried.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>attempts</b></td>
        <td>integer</td>
        <td>
          Attempts is the number of retries to be allowed for a given request before giving up. The interval between retries will be determined automatically (25ms+).<br/>Default is 2<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>perTryTimeout</b></td>
        <td>string</td>
        <td>
          PerTryTimeout is the timeout per attempt for a given request, including the initial call and any retries. Format: 1h/1m/1s/1ms. MUST be &gt;=1ms.<br/>Default: no timeout<br/>
          <br/>
            <i>Format</i>: duration<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>retryOnHttpResponseCodes</b></td>
        <td>[]int or string</td>
        <td>
          RetryOnHttpResponseCodes HTTP response codes that should trigger a retry. A typical value is [503].<br/>You may also use 5xx and retriable-4xx (only 409).<br/>mixed types are allowed such as [503, &#34;retriable-4xx&#34;]<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecistiosettingstelemetry"></a>
#### Application.spec.istioSettings.telemetry

<sup>[Parent](#applicationspecistiosettings)</sup>

Telemetry is a placeholder for all relevant telemetry types, and may be extended in the future to configure additional telemetry settings.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b><a href="#applicationspecistiosettingstelemetrytracingindex">tracing</a></b></td>
        <td>[]object</td>
        <td>
          Tracing is a list of tracing configurations for the telemetry resource. Normally only one tracing configuration is needed.<br/>
          <br/>
            <i>Default</i>: `[map[randomSamplingPercentage:10]]`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecistiosettingstelemetrytracingindex"></a>
#### Application.spec.istioSettings.telemetry.tracing[index]

<sup>[Parent](#applicationspecistiosettingstelemetry)</sup>

Tracing contains relevant settings for tracing in the telemetry configuration

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>randomSamplingPercentage</b></td>
        <td>integer</td>
        <td>
          RandomSamplingPercentage is the percentage of requests that should be sampled for tracing, specified by a whole number between 0-100.<br/>Setting RandomSamplingPercentage to 0 will disable tracing.<br/>
          <br/>
            <i>Default</i>: `10`<br/>
            <i>Minimum</i>: 0<br/>
            <i>Maximum</i>: 100<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecliveness"></a>
#### Application.spec.liveness

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after<br/>having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `3`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that<br/>are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `0`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `10`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.<br/>Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.<br/>Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecmaskinporten"></a>
#### Application.spec.maskinporten

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          If enabled, provisions and configures a Maskinporten client with consumed scopes and/or Exposed scopes with DigDir.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>clientName</b></td>
        <td>string</td>
        <td>
          The name of the Client as shown in Digitaliseringsdirektoratet&#39;s Samarbeidsportal<br/>Meant to be a human-readable name for separating clients in the portal<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecmaskinportenrequestauthentication">requestAuthentication</a></b></td>
        <td>object</td>
        <td>
          RequestAuthentication specifies how incoming JWTs should be validated.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecmaskinportenscopes">scopes</a></b></td>
        <td>object</td>
        <td>
          Schema to configure Maskinporten clients with consumed scopes and/or exposed scopes.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecmaskinportenrequestauthentication"></a>
#### Application.spec.maskinporten.requestAuthentication

<sup>[Parent](#applicationspecmaskinporten)</sup>

RequestAuthentication specifies how incoming JWTs should be validated.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          Whether to enable JWT validation.<br/>If enabled, incoming JWTs will be validated against the issuer specified in the app registration and the generated audience.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>forwardJwt</b></td>
        <td>boolean</td>
        <td>
          If set to `true`, the original token will be kept for the upstream request. Defaults to `true`.<br/>
          <br/>
            <i>Default</i>: `true`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>ignorePaths</b></td>
        <td>[]string</td>
        <td>
          IgnorePaths specifies paths that do not require an authenticated JWT.<br/><br/>The specified paths must be a valid URI path. It has to start with &#39;/&#39; and cannot end with &#39;/&#39;.<br/>The paths can also contain the wildcard operator &#39;*&#39;, but only at the end.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecmaskinportenrequestauthenticationoutputclaimtoheadersindex">outputClaimToHeaders</a></b></td>
        <td>[]object</td>
        <td>
          This field specifies a list of operations to copy the claim to HTTP headers on a successfully verified token.<br/>The header specified in each operation in the list must be unique. Nested claims of type string/int/bool is supported as well.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>paths</b></td>
        <td>[]string</td>
        <td>
          Paths specifies paths that require an authenticated JWT.<br/><br/>The specified paths must be a valid URI path. It has to start with &#39;/&#39; and cannot end with &#39;/&#39;.<br/>The paths can also contain the wildcard operator &#39;*&#39;, but only at the end.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>secretName</b></td>
        <td>string</td>
        <td>
          The name of the Kubernetes Secret containing OAuth2 credentials.<br/><br/>If omitted, the associated client registration in the application manifest is used for JWT validation.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>tokenLocation</b></td>
        <td>enum</td>
        <td>
          Where to find the JWT in the incoming request<br/><br/>An enum value of `header` means that the JWT is present in the `Authorization` header as a `Bearer` token.<br/>An enum value of `cookie` means that the JWT is present as a cookie called `BearerToken`.<br/><br/>If omitted, its default value depends on the provider type:<br/>  Defaults to &#34;cookie&#34; for providers supporting user login (e.g. IDPorten).<br/>  Defaults to &#34;header&#34; for providers not supporting user login (e.g. Maskinporten).<br/>
          <br/>
            <i>Enum</i>: header, cookie<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecmaskinportenrequestauthenticationoutputclaimtoheadersindex"></a>
#### Application.spec.maskinporten.requestAuthentication.outputClaimToHeaders[index]

<sup>[Parent](#applicationspecmaskinportenrequestauthentication)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>claim</b></td>
        <td>string</td>
        <td>
          The claim to be copied.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>header</b></td>
        <td>string</td>
        <td>
          The name of the HTTP header for which the specified claim will be copied to.<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecmaskinportenscopes"></a>
#### Application.spec.maskinporten.scopes

<sup>[Parent](#applicationspecmaskinporten)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#applicationspecmaskinportenscopesconsumesindex">consumes</a></b></td>
        <td>[]object</td>
        <td>
          This is the Schema for the consumes and exposes API.<br/>`consumes` is a list of scopes that your client can request access to.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecmaskinportenscopesexposesindex">exposes</a></b></td>
        <td>[]object</td>
        <td>
          `exposes` is a list of scopes your application want to expose to other organization where access to the scope is based on organization number.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecmaskinportenscopesconsumesindex"></a>
#### Application.spec.maskinporten.scopes.consumes[index]

<sup>[Parent](#applicationspecmaskinportenscopes)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          The scope consumed by the application to gain access to an external organization API.<br/>Ensure that the NAV organization has been granted access to the scope prior to requesting access.<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecmaskinportenscopesexposesindex"></a>
#### Application.spec.maskinporten.scopes.exposes[index]

<sup>[Parent](#applicationspecmaskinportenscopes)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          If Enabled the configured scope is available to be used and consumed by organizations granted access.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          The actual subscope combined with `Product`.<br/>Ensure that `&lt;Product&gt;&lt;Name&gt;` matches `Pattern`.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>product</b></td>
        <td>string</td>
        <td>
          The product-area your application belongs to e.g. arbeid, helse ...<br/>This will be included in the final scope `nav:&lt;Product&gt;&lt;Name&gt;`.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>accessibleForAll</b></td>
        <td>boolean</td>
        <td>
          Allow any organization to access the scope.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>allowedIntegrations</b></td>
        <td>[]string</td>
        <td>
          Whitelisting of integration&#39;s allowed.<br/>Default is `maskinporten`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>atMaxAge</b></td>
        <td>integer</td>
        <td>
          Max time in seconds for a issued access_token.<br/>Default is `30` sec.<br/>
          <br/>
            <i>Minimum</i>: 30<br/>
            <i>Maximum</i>: 680<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecmaskinportenscopesexposesindexconsumersindex">consumers</a></b></td>
        <td>[]object</td>
        <td>
          External consumers granted access to this scope and able to request access_token.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>delegationSource</b></td>
        <td>enum</td>
        <td>
          Delegation source for the scope. Default is empty, which means no delegation is allowed.<br/>
          <br/>
            <i>Enum</i>: altinn<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>separator</b></td>
        <td>string</td>
        <td>
          Separator is the character that separates `product` and `name` in the final scope:<br/>`scope := &lt;prefix&gt;:&lt;product&gt;&lt;separator&gt;&lt;name&gt;`<br/>This overrides the default separator.<br/>The default separator is `:`. If `name` contains `/`, the default separator is instead `/`.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>visibility</b></td>
        <td>enum</td>
        <td>
          Visibility controls the scope&#39;s visibility.<br/>Public scopes are visible for everyone.<br/>Private scopes are only visible for the organization that owns the scope as well as<br/>organizations that have been granted consumer access.<br/>
          <br/>
            <i>Enum</i>: private, public<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecmaskinportenscopesexposesindexconsumersindex"></a>
#### Application.spec.maskinporten.scopes.exposes[index].consumers[index]

<sup>[Parent](#applicationspecmaskinportenscopesexposesindex)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>orgno</b></td>
        <td>string</td>
        <td>
          The external business/organization number.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          This is a describing field intended for clarity not used for any other purpose.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecpodsettings"></a>
#### Application.spec.podSettings

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          Annotations that are set on Pods created by Skiperator. These annotations can for example be used to change the behaviour of sidecars and similar.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>disablePodSpreadTopologyConstraints</b></td>
        <td>boolean</td>
        <td>
          DisablePodSpreadTopologyConstraints specifies whether to disable the addition of Pod Topology Spread Constraints to<br/>a given pod.<br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          TerminationGracePeriodSeconds determines how long Kubernetes waits after a SIGTERM signal sent to a Pod before terminating the pod. If your application uses longer than<br/>30 seconds to terminate, you should increase TerminationGracePeriodSeconds.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Default</i>: `30`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecprometheus"></a>
#### Application.spec.prometheus

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          The port number or name where metrics are exposed (at the Pod level).<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>allowAllMetrics</b></td>
        <td>boolean</td>
        <td>
          Setting AllowAllMetrics to true will ensure all exposed metrics are scraped. Otherwise, a list of predefined<br/>metrics will be dropped by default. See util/constants.go for the default list.<br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The HTTP path where Prometheus compatible metrics exists<br/>
          <br/>
            <i>Default</i>: `/metrics`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>scrapeInterval</b></td>
        <td>string</td>
        <td>
          ScrapeInterval specifies the interval at which Prometheus should scrape the metrics.<br/>The interval must be at least 15 seconds (if using &#34;Xs&#34;) and divisible by 5.<br/>If minutes (&#34;Xm&#34;) are used, the value must be at least 1m.<br/>
          <br/>
            <i>Default</i>: `60s`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecreadiness"></a>
#### Application.spec.readiness

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after<br/>having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `3`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that<br/>are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `0`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `10`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.<br/>Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.<br/>Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecresources"></a>
#### Application.spec.resources

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits set the maximum the app is allowed to use. Exceeding this limit will<br/>make kubernetes kill the app and restart it.<br/><br/>Limits can be set on the CPU and memory, but it is not recommended to put a limit on CPU, see: https://home.robusta.dev/blog/stop-using-cpu-limits<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests set the initial allocation that is done for the app and will<br/>thus be available to the app on startup. More is allocated on demand<br/>until the limit is reached.<br/><br/>Requests can be set on the CPU and memory.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecstartup"></a>
#### Application.spec.startup

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after<br/>having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `3`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that<br/>are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `0`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `10`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.<br/>Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.<br/>Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecstateful"></a>
#### Application.spec.stateful

<sup>[Parent](#applicationspec)</sup>

Stateful, when set with enabled=true, generates a StatefulSet instead of a Deployment.
Requires VolumeClaimTemplates. Disallows Strategy.Type=Recreate and HPA-range replicas.
The enabled flag is immutable - delete and recreate the Application to change.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          When true, generates a StatefulSet instead of a Deployment.<br/>This value is immutable - delete and recreate the Application to change<br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>partition</b></td>
        <td>integer</td>
        <td>
          Staged rollouts - only pods with ordinal &gt;= Partition are updated.<br/>Set Partition equal to replicas to pause updates.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>podManagementPolicy</b></td>
        <td>enum</td>
        <td>
          Controls pod creation and update order. OrderedReady creates pods one at a time, Parallel creates them simultaneously.<br/>
          <br/>
            <i>Enum</i>: OrderedReady, Parallel<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>pvcRetentionWhenDeleted</b></td>
        <td>enum</td>
        <td>
          PVC fate when the StatefulSet is deleted. Defaults to Retain.<br/>
          <br/>
            <i>Enum</i>: Retain, Delete<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>pvcRetentionWhenScaled</b></td>
        <td>enum</td>
        <td>
          PVC fate when the StatefulSet is scaled down. Defaults to Retain.<br/>
          <br/>
            <i>Enum</i>: Retain, Delete<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecstatefulvolumeclaimtemplatesindex">volumeClaimTemplates</a></b></td>
        <td>[]object</td>
        <td>
          Per-pod PersistentVolumeClaims provisioned by the StatefulSet controller.<br/>Each replica gets its own PVC named `&lt;template.metadata.name&gt;-&lt;app&gt;-&lt;ordinal&gt;`.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecstatefulvolumeclaimtemplatesindex"></a>
#### Application.spec.stateful.volumeClaimTemplates[index]

<sup>[Parent](#applicationspecstateful)</sup>

VolumeClaimTemplate describes a per-pod PersistentVolumeClaim provisioned by the StatefulSet
controller. Name serves as both the pod volume reference and the PVC prefix

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>mountPath</b></td>
        <td>string</td>
        <td>
          Where the volume is mounted inside the container<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Pod volume name and PVC name prefix. Resulting PVCs are named `&lt;name&gt;-&lt;app&gt;-&lt;ordinal&gt;`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecstatefulvolumeclaimtemplatesindexspec">spec</a></b></td>
        <td>object</td>
        <td>
          PVC spec<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          Optional annotations applied to PVCs<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          Optional labels applied to PVCs<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>subPath</b></td>
        <td>string</td>
        <td>
          Subpath within the volume to mount instead of its root<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecstatefulvolumeclaimtemplatesindexspec"></a>
#### Application.spec.stateful.volumeClaimTemplates[index].spec

<sup>[Parent](#applicationspecstatefulvolumeclaimtemplatesindex)</sup>

PVC spec

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>accessModes</b></td>
        <td>[]string</td>
        <td>
          accessModes contains the desired access modes the volume should have.<br/>More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecstatefulvolumeclaimtemplatesindexspecdatasource">dataSource</a></b></td>
        <td>object</td>
        <td>
          dataSource field can be used to specify either:<br/>* An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot)<br/>* An existing PVC (PersistentVolumeClaim)<br/>If the provisioner or an external controller can support the specified data source,<br/>it will create a new volume based on the contents of the specified data source.<br/>When the AnyVolumeDataSource feature gate is enabled, dataSource contents will be copied to dataSourceRef,<br/>and dataSourceRef contents will be copied to dataSource when dataSourceRef.namespace is not specified.<br/>If the namespace is specified, then dataSourceRef will not be copied to dataSource.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecstatefulvolumeclaimtemplatesindexspecdatasourceref">dataSourceRef</a></b></td>
        <td>object</td>
        <td>
          dataSourceRef specifies the object from which to populate the volume with data, if a non-empty<br/>volume is desired. This may be any object from a non-empty API group (non<br/>core object) or a PersistentVolumeClaim object.<br/>When this field is specified, volume binding will only succeed if the type of<br/>the specified object matches some installed volume populator or dynamic<br/>provisioner.<br/>This field will replace the functionality of the dataSource field and as such<br/>if both fields are non-empty, they must have the same value. For backwards<br/>compatibility, when namespace isn&#39;t specified in dataSourceRef,<br/>both fields (dataSource and dataSourceRef) will be set to the same<br/>value automatically if one of them is empty and the other is non-empty.<br/>When namespace is specified in dataSourceRef,<br/>dataSource isn&#39;t set to the same value and must be empty.<br/>There are three important differences between dataSource and dataSourceRef:<br/>* While dataSource only allows two specific types of objects, dataSourceRef<br/>  allows any non-core object, as well as PersistentVolumeClaim objects.<br/>* While dataSource ignores disallowed values (dropping them), dataSourceRef<br/>  preserves all values, and generates an error if a disallowed value is<br/>  specified.<br/>* While dataSource only allows local objects, dataSourceRef allows objects<br/>  in any namespaces.<br/>(Beta) Using this field requires the AnyVolumeDataSource feature gate to be enabled.<br/>(Alpha) Using the namespace field of dataSourceRef requires the CrossNamespaceVolumeDataSource feature gate to be enabled.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecstatefulvolumeclaimtemplatesindexspecresources">resources</a></b></td>
        <td>object</td>
        <td>
          resources represents the minimum resources the volume should have.<br/>Users are allowed to specify resource requirements<br/>that are lower than previous value but must still be higher than capacity recorded in the<br/>status field of the claim.<br/>More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#applicationspecstatefulvolumeclaimtemplatesindexspecselector">selector</a></b></td>
        <td>object</td>
        <td>
          selector is a label query over volumes to consider for binding.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>storageClassName</b></td>
        <td>string</td>
        <td>
          storageClassName is the name of the StorageClass required by the claim.<br/>More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>volumeAttributesClassName</b></td>
        <td>string</td>
        <td>
          volumeAttributesClassName may be used to set the VolumeAttributesClass used by this claim.<br/>If specified, the CSI driver will create or update the volume with the attributes defined<br/>in the corresponding VolumeAttributesClass. This has a different purpose than storageClassName,<br/>it can be changed after the claim is created. An empty string or nil value indicates that no<br/>VolumeAttributesClass will be applied to the claim. If the claim enters an Infeasible error state,<br/>this field can be reset to its previous value (including nil) to cancel the modification.<br/>If the resource referred to by volumeAttributesClass does not exist, this PersistentVolumeClaim will be<br/>set to a Pending state, as reflected by the modifyVolumeStatus field, until such as a resource<br/>exists.<br/>More info: https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>volumeMode</b></td>
        <td>string</td>
        <td>
          volumeMode defines what type of volume is required by the claim.<br/>Value of Filesystem is implied when not included in claim spec.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          volumeName is the binding reference to the PersistentVolume backing this claim.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecstatefulvolumeclaimtemplatesindexspecdatasource"></a>
#### Application.spec.stateful.volumeClaimTemplates[index].spec.dataSource

<sup>[Parent](#applicationspecstatefulvolumeclaimtemplatesindexspec)</sup>

dataSource field can be used to specify either:
* An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot)
* An existing PVC (PersistentVolumeClaim)
If the provisioner or an external controller can support the specified data source,
it will create a new volume based on the contents of the specified data source.
When the AnyVolumeDataSource feature gate is enabled, dataSource contents will be copied to dataSourceRef,
and dataSourceRef contents will be copied to dataSource when dataSourceRef.namespace is not specified.
If the namespace is specified, then dataSourceRef will not be copied to dataSource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is the type of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>apiGroup</b></td>
        <td>string</td>
        <td>
          APIGroup is the group for the resource being referenced.<br/>If APIGroup is not specified, the specified Kind must be in the core API group.<br/>For any other third-party types, APIGroup is required.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecstatefulvolumeclaimtemplatesindexspecdatasourceref"></a>
#### Application.spec.stateful.volumeClaimTemplates[index].spec.dataSourceRef

<sup>[Parent](#applicationspecstatefulvolumeclaimtemplatesindexspec)</sup>

dataSourceRef specifies the object from which to populate the volume with data, if a non-empty
volume is desired. This may be any object from a non-empty API group (non
core object) or a PersistentVolumeClaim object.
When this field is specified, volume binding will only succeed if the type of
the specified object matches some installed volume populator or dynamic
provisioner.
This field will replace the functionality of the dataSource field and as such
if both fields are non-empty, they must have the same value. For backwards
compatibility, when namespace isn't specified in dataSourceRef,
both fields (dataSource and dataSourceRef) will be set to the same
value automatically if one of them is empty and the other is non-empty.
When namespace is specified in dataSourceRef,
dataSource isn't set to the same value and must be empty.
There are three important differences between dataSource and dataSourceRef:
* While dataSource only allows two specific types of objects, dataSourceRef
  allows any non-core object, as well as PersistentVolumeClaim objects.
* While dataSource ignores disallowed values (dropping them), dataSourceRef
  preserves all values, and generates an error if a disallowed value is
  specified.
* While dataSource only allows local objects, dataSourceRef allows objects
  in any namespaces.
(Beta) Using this field requires the AnyVolumeDataSource feature gate to be enabled.
(Alpha) Using the namespace field of dataSourceRef requires the CrossNamespaceVolumeDataSource feature gate to be enabled.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is the type of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>apiGroup</b></td>
        <td>string</td>
        <td>
          APIGroup is the group for the resource being referenced.<br/>If APIGroup is not specified, the specified Kind must be in the core API group.<br/>For any other third-party types, APIGroup is required.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace is the namespace of resource being referenced<br/>Note that when a namespace is specified, a gateway.networking.k8s.io/ReferenceGrant object is required in the referent namespace to allow that namespace&#39;s owner to accept the reference. See the ReferenceGrant documentation for details.<br/>(Alpha) This field requires the CrossNamespaceVolumeDataSource feature gate to be enabled.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecstatefulvolumeclaimtemplatesindexspecresources"></a>
#### Application.spec.stateful.volumeClaimTemplates[index].spec.resources

<sup>[Parent](#applicationspecstatefulvolumeclaimtemplatesindexspec)</sup>

resources represents the minimum resources the volume should have.
Users are allowed to specify resource requirements
that are lower than previous value but must still be higher than capacity recorded in the
status field of the claim.
More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits describes the maximum amount of compute resources allowed.<br/>More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests describes the minimum amount of compute resources required.<br/>If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,<br/>otherwise to an implementation-defined value. Requests cannot exceed Limits.<br/>More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecstatefulvolumeclaimtemplatesindexspecselector"></a>
#### Application.spec.stateful.volumeClaimTemplates[index].spec.selector

<sup>[Parent](#applicationspecstatefulvolumeclaimtemplatesindexspec)</sup>

selector is a label query over volumes to consider for binding.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b><a href="#applicationspecstatefulvolumeclaimtemplatesindexspecselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of &#123;key,value&#125; pairs. A single &#123;key,value&#125; in the matchLabels<br/>map is equivalent to an element of matchExpressions, whose key field is &#34;key&#34;, the<br/>operator is &#34;In&#34;, and the values array contains only &#34;value&#34;. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecstatefulvolumeclaimtemplatesindexspecselectormatchexpressionsindex"></a>
#### Application.spec.stateful.volumeClaimTemplates[index].spec.selector.matchExpressions[index]

<sup>[Parent](#applicationspecstatefulvolumeclaimtemplatesindexspecselector)</sup>

A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key&#39;s relationship to a set of values.<br/>Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,<br/>the values array must be non-empty. If the operator is Exists or DoesNotExist,<br/>the values array must be empty. This array is replaced during a strategic<br/>merge patch.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationspecstrategy"></a>
#### Application.spec.strategy

<sup>[Parent](#applicationspec)</sup>

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
    <tbody>
      <tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Valid values are: RollingUpdate, Recreate. Default is RollingUpdate<br/>
          <br/>
            <i>Enum</i>: RollingUpdate, Recreate<br/>
            <i>Default</i>: `RollingUpdate`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationstatus"></a>
#### Application.status

<sup>[Parent](#application)</sup>

ApplicationStatus is a specialized status specific to the Application kind.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>accessPolicies</b></td>
        <td>string</td>
        <td>
          Indicates if access policies are valid<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#applicationstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#applicationstatussubresourceskey">subresources</a></b></td>
        <td>map[string]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#applicationstatussummary">summary</a></b></td>
        <td>object</td>
        <td>
          Status<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>applicationKind</b></td>
        <td>string</td>
        <td>
          Kind generated for this Application after a successful reconcile.<br/>Used to prevent switching between Deployment and StatefulSet.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationstatusconditionsindex"></a>
#### Application.status.conditions[index]

<sup>[Parent](#applicationstatus)</sup>

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
    <tbody>
      <tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.<br/>This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.<br/>This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition&#39;s last transition.<br/>Producers of specific condition types may define expected values and meanings for this field,<br/>and whether the values are considered a guaranteed API.<br/>The value should be a CamelCase string.<br/>This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.<br/>For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date<br/>with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="applicationstatussubresourceskey"></a>
#### Application.status.subresources[key]

<sup>[Parent](#applicationstatus)</sup>

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
    <tbody>
      <tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Resource accepted by Kubernetes. Waiting for Skiperator to become aware of the resource and start processing.`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Pending`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="applicationstatussummary"></a>
#### Application.status.summary

<sup>[Parent](#applicationstatus)</sup>

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
    <tbody>
      <tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Resource accepted by Kubernetes. Waiting for Skiperator to become aware of the resource and start processing.`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Pending`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="routing"></a>
### Routing

| Field | Value |
| --- | --- |
| Package | `skiperator.kartverket.no/v1alpha1` |
| API version | `skiperator.kartverket.no/v1alpha1` |
| Kind | `Routing` |



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
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
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#routingspec">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#routingstatus">status</a></b></td>
        <td>object</td>
        <td>
          SkiperatorStatus<br/><br/>A status field shown on a Skiperator resource which contains information regarding deployment of the resource.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="routingspec"></a>
#### Routing.spec

<sup>[Parent](#routing)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>hostname</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#routingspecroutesindex">routes</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>redirectToHTTPS</b></td>
        <td>boolean</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `true`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="routingspecroutesindex"></a>
#### Routing.spec.routes[index]

<sup>[Parent](#routingspec)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>pathPrefix</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>targetApp</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>rewriteUri</b></td>
        <td>boolean</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="routingstatus"></a>
#### Routing.status

<sup>[Parent](#routing)</sup>

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
    <tbody>
      <tr>
        <td><b>accessPolicies</b></td>
        <td>string</td>
        <td>
          Indicates if access policies are valid<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#routingstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#routingstatussubresourceskey">subresources</a></b></td>
        <td>map[string]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#routingstatussummary">summary</a></b></td>
        <td>object</td>
        <td>
          Status<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="routingstatusconditionsindex"></a>
#### Routing.status.conditions[index]

<sup>[Parent](#routingstatus)</sup>

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
    <tbody>
      <tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.<br/>This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.<br/>This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition&#39;s last transition.<br/>Producers of specific condition types may define expected values and meanings for this field,<br/>and whether the values are considered a guaranteed API.<br/>The value should be a CamelCase string.<br/>This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.<br/>For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date<br/>with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="routingstatussubresourceskey"></a>
#### Routing.status.subresources[key]

<sup>[Parent](#routingstatus)</sup>

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
    <tbody>
      <tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Resource accepted by Kubernetes. Waiting for Skiperator to become aware of the resource and start processing.`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Pending`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="routingstatussummary"></a>
#### Routing.status.summary

<sup>[Parent](#routingstatus)</sup>

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
    <tbody>
      <tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Resource accepted by Kubernetes. Waiting for Skiperator to become aware of the resource and start processing.`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Pending`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="skipjob"></a>
### SKIPJob

| Field | Value |
| --- | --- |
| Package | `skiperator.kartverket.no/v1alpha1` |
| API version | `skiperator.kartverket.no/v1alpha1` |
| Kind | `SKIPJob` |

SKIPJob is the deprecated schema for the SKIPJobs API. Please migrate to v1beta1.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
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
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspec">spec</a></b></td>
        <td>object</td>
        <td>
          SKIPJobSpec defines the desired state of SKIPJob<br/><br/>A SKIPJob is either defined as a one-off or a scheduled job. If the Cron field is set for SKIPJob, it may not be removed. If the Cron field is unset, it may not be added.<br/>The Container field of a SKIPJob is only mutable if the Cron field is set. If unset, you must delete your SKIPJob to change container settings.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobstatus">status</a></b></td>
        <td>object</td>
        <td>
          SkiperatorStatus<br/><br/>A status field shown on a Skiperator resource which contains information regarding deployment of the resource.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspec"></a>
#### SKIPJob.spec

<sup>[Parent](#skipjob)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#skipjobspeccontainer">container</a></b></td>
        <td>object</td>
        <td>
          Settings for the Pods running in the job. Fields are mostly the same as an Application, and are (probably) better documented there. Some fields are omitted, but none added.<br/>Once set, you may not change Container without deleting your current SKIPJob<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccron">cron</a></b></td>
        <td>object</td>
        <td>
          Settings for the Job if you are running a scheduled job. Optional as Jobs may be one-off.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecistiosettings">istioSettings</a></b></td>
        <td>object</td>
        <td>
          IstioSettings are used to configure istio specific resources such as telemetry. Currently, adjusting sampling<br/>interval for tracing is the only supported option.<br/>By default, tracing is enabled with a random sampling percentage of 10%.<br/>
          <br/>
            <i>Default</i>: `map[telemetry:map[tracing:[map[randomSamplingPercentage:10]]]]`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecjob">job</a></b></td>
        <td>object</td>
        <td>
          Settings for the actual Job. If you use a scheduled job, the settings in here will also specify the template of the job.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          Labels can be used if you want every resource created by your SKIPJob to<br/>have the same labels, including the Job/CronJob itself. This could for example be useful for<br/>metrics, where a certain label and the corresponding resources liveliness can be combined.<br/>Any amount of labels can be added as wanted, and they will all cascade down to all resources.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecprometheus">prometheus</a></b></td>
        <td>object</td>
        <td>
          Prometheus settings for pod running in job. Fields are identical to Application and if set,<br/>a podmonitoring object is created.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>team</b></td>
        <td>string</td>
        <td>
          Team specifies the team who owns this particular SKIPJob.<br/>Usually sourced from the namespace label.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainer"></a>
#### SKIPJob.spec.container

<sup>[Parent](#skipjobspec)</sup>

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
    <tbody>
      <tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicy">accessPolicy</a></b></td>
        <td>object</td>
        <td>
          AccessPolicy<br/><br/>Zero trust dictates that only applications with a reason for being able<br/>to access another resource should be able to reach it. This is set up by<br/>default by denying all ingress and egress traffic from the Pods in the<br/>Deployment. The AccessPolicy field is an allowlist of other applications and hostnames<br/>that are allowed to talk with this Application and which resources this app can talk to<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontaineradditionalportsindex">additionalPorts</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainerenvindex">env</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainerenvfromindex">envFrom</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainerfilesfromindex">filesFrom</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainergcp">gcp</a></b></td>
        <td>object</td>
        <td>
          GCP<br/><br/>Configuration for interacting with Google Cloud Platform<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainerliveness">liveness</a></b></td>
        <td>object</td>
        <td>
          Probe<br/><br/>Type configuration for all types of Kubernetes probes.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainerpodsettings">podSettings</a></b></td>
        <td>object</td>
        <td>
          PodSettings<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>priority</b></td>
        <td>enum</td>
        <td>
          <br/>
          <br/>
            <i>Enum</i>: low, medium, high<br/>
            <i>Default</i>: `medium`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainerreadiness">readiness</a></b></td>
        <td>object</td>
        <td>
          Probe<br/><br/>Type configuration for all types of Kubernetes probes.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainerresources">resources</a></b></td>
        <td>object</td>
        <td>
          ResourceRequirements<br/><br/>A simplified version of the Kubernetes native ResourceRequirement field, in which only Limits and Requests are present.<br/>For the units used for resources, see https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-units-in-kubernetes<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>restartPolicy</b></td>
        <td>enum</td>
        <td>
          RestartPolicy describes how the container should be restarted.<br/>Only one of the following restart policies may be specified.<br/>If none of the following policies is specified, the default one<br/>is RestartPolicyAlways.<br/>
          <br/>
            <i>Enum</i>: OnFailure, Never<br/>
            <i>Default</i>: `Never`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainerstartup">startup</a></b></td>
        <td>object</td>
        <td>
          Probe<br/><br/>Type configuration for all types of Kubernetes probes.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontaineraccesspolicy"></a>
#### SKIPJob.spec.container.accessPolicy

<sup>[Parent](#skipjobspeccontainer)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyinbound">inbound</a></b></td>
        <td>object</td>
        <td>
          Inbound specifies the ingress rules. Which apps on the cluster can talk to this app?<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyoutbound">outbound</a></b></td>
        <td>object</td>
        <td>
          Outbound specifies egress rules. Which apps on the cluster and the<br/>internet is the Application allowed to send requests to?<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontaineraccesspolicyinbound"></a>
#### SKIPJob.spec.container.accessPolicy.inbound

<sup>[Parent](#skipjobspeccontaineraccesspolicy)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyinboundrulesindex">rules</a></b></td>
        <td>[]object</td>
        <td>
          The rules list specifies a list of applications. When no namespace is<br/>specified it refers to an app in the current namespace. For apps in<br/>other namespaces namespace is required<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontaineraccesspolicyinboundrulesindex"></a>
#### SKIPJob.spec.container.accessPolicy.inbound.rules[index]

<sup>[Parent](#skipjobspeccontaineraccesspolicyinbound)</sup>

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
    <tbody>
      <tr>
        <td><b>application</b></td>
        <td>string</td>
        <td>
          The name of the Application you are allowing traffic to/from. If you wish to allow traffic from a SKIPJob, this field should<br/>be suffixed with -skipjob<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          The namespace in which the Application you are allowing traffic to/from resides. If unset, uses namespace of Application.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>namespacesByLabel</b></td>
        <td>map[string]string</td>
        <td>
          Namespace label value-pair in which the Application you are allowing traffic to/from resides. If both namespace and namespacesByLabel are set, namespace takes precedence and namespacesByLabel is omitted.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyinboundrulesindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above application.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontaineraccesspolicyinboundrulesindexportsindex"></a>
#### SKIPJob.spec.container.accessPolicy.inbound.rules[index].ports[index]

<sup>[Parent](#skipjobspeccontaineraccesspolicyinboundrulesindex)</sup>

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
    <tbody>
      <tr>
        <td><b>endPort</b></td>
        <td>integer</td>
        <td>
          endPort indicates that the range of ports from port to endPort if set, inclusive,<br/>should be allowed by the policy. This field cannot be defined if the port field<br/>is not defined or if the port field is defined as a named (string) port.<br/>The endPort must be equal or greater than port.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          port represents the port on the given protocol. This can either be a numerical or named<br/>port on a pod. If this field is not provided, this matches all port names and<br/>numbers.<br/>If present, only traffic on the specified protocol AND port will be matched.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          protocol represents the protocol (TCP, UDP, or SCTP) which traffic must match.<br/>If not specified, this field defaults to TCP.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontaineraccesspolicyoutbound"></a>
#### SKIPJob.spec.container.accessPolicy.outbound

<sup>[Parent](#skipjobspeccontaineraccesspolicy)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyoutboundexternalindex">external</a></b></td>
        <td>[]object</td>
        <td>
          External specifies which applications on the internet the application<br/>can reach. Only host is required unless it is on another port than HTTPS port 443.<br/>If other ports or protocols are required then `ports` must be specified as well<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyoutboundrulesindex">rules</a></b></td>
        <td>[]object</td>
        <td>
          Rules apply the same in-cluster rules as InboundPolicy<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontaineraccesspolicyoutboundexternalindex"></a>
#### SKIPJob.spec.container.accessPolicy.outbound.external[index]

<sup>[Parent](#skipjobspeccontaineraccesspolicyoutbound)</sup>

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
    <tbody>
      <tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          The allowed hostname. Note that this does not include subdomains.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>ip</b></td>
        <td>string</td>
        <td>
          Non-HTTP requests (i.e. using the TCP protocol) need to use IP in addition to hostname<br/>Only required for TCP requests.<br/><br/>Note: Hostname must always be defined even if IP is set statically<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyoutboundexternalindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above hostname. When not specified HTTP and<br/>HTTPS on port 80 and 443 respectively are put into the allowlist<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontaineraccesspolicyoutboundexternalindexportsindex"></a>
#### SKIPJob.spec.container.accessPolicy.outbound.external[index].ports[index]

<sup>[Parent](#skipjobspeccontaineraccesspolicyoutboundexternalindex)</sup>

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
    <tbody>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is required and is an arbitrary name. Must be unique within all ExternalRule ports.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          The port number of the external host<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>protocol</b></td>
        <td>enum</td>
        <td>
          The protocol to use for communication with the host. Supported protocols are: HTTP, HTTPS, TCP and TLS.<br/>
          <br/>
            <i>Enum</i>: HTTP, HTTPS, TCP, TLS<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontaineraccesspolicyoutboundrulesindex"></a>
#### SKIPJob.spec.container.accessPolicy.outbound.rules[index]

<sup>[Parent](#skipjobspeccontaineraccesspolicyoutbound)</sup>

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
    <tbody>
      <tr>
        <td><b>application</b></td>
        <td>string</td>
        <td>
          The name of the Application you are allowing traffic to/from. If you wish to allow traffic from a SKIPJob, this field should<br/>be suffixed with -skipjob<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          The namespace in which the Application you are allowing traffic to/from resides. If unset, uses namespace of Application.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>namespacesByLabel</b></td>
        <td>map[string]string</td>
        <td>
          Namespace label value-pair in which the Application you are allowing traffic to/from resides. If both namespace and namespacesByLabel are set, namespace takes precedence and namespacesByLabel is omitted.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontaineraccesspolicyoutboundrulesindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above application.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontaineraccesspolicyoutboundrulesindexportsindex"></a>
#### SKIPJob.spec.container.accessPolicy.outbound.rules[index].ports[index]

<sup>[Parent](#skipjobspeccontaineraccesspolicyoutboundrulesindex)</sup>

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
    <tbody>
      <tr>
        <td><b>endPort</b></td>
        <td>integer</td>
        <td>
          endPort indicates that the range of ports from port to endPort if set, inclusive,<br/>should be allowed by the policy. This field cannot be defined if the port field<br/>is not defined or if the port field is defined as a named (string) port.<br/>The endPort must be equal or greater than port.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          port represents the port on the given protocol. This can either be a numerical or named<br/>port on a pod. If this field is not provided, this matches all port names and<br/>numbers.<br/>If present, only traffic on the specified protocol AND port will be matched.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          protocol represents the protocol (TCP, UDP, or SCTP) which traffic must match.<br/>If not specified, this field defaults to TCP.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontaineradditionalportsindex"></a>
#### SKIPJob.spec.container.additionalPorts[index]

<sup>[Parent](#skipjobspeccontainer)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>protocol</b></td>
        <td>enum</td>
        <td>
          Protocol defines network protocols supported for things like container ports.<br/>
          <br/>
            <i>Enum</i>: TCP, UDP, SCTP<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerenvindex"></a>
#### SKIPJob.spec.container.env[index]

<sup>[Parent](#skipjobspeccontainer)</sup>

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
    <tbody>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the environment variable.<br/>May consist of any printable ASCII characters except &#39;=&#39;.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Variable references $(VAR_NAME) are expanded<br/>using the previously defined environment variables in the container and<br/>any service environment variables. If a variable cannot be resolved,<br/>the reference in the input string will be unchanged. Double $$ are reduced<br/>to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e.<br/>&#34;$$(VAR_NAME)&#34; will produce the string literal &#34;$(VAR_NAME)&#34;.<br/>Escaped references will never be expanded, regardless of whether the variable<br/>exists or not.<br/>Defaults to &#34;&#34;.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainerenvindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Source for the environment variable&#39;s value. Cannot be used if value is not empty.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerenvindexvaluefrom"></a>
#### SKIPJob.spec.container.env[index].valueFrom

<sup>[Parent](#skipjobspeccontainerenvindex)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#skipjobspeccontainerenvindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainerenvindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels[&#39;&lt;KEY&gt;&#39;]`, `metadata.annotations[&#39;&lt;KEY&gt;&#39;]`,<br/>spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainerenvindexvaluefromfilekeyref">fileKeyRef</a></b></td>
        <td>object</td>
        <td>
          FileKeyRef selects a key of the env file.<br/>Requires the EnvFiles feature gate to be enabled.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainerenvindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests<br/>(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainerenvindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a secret in the pod&#39;s namespace<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerenvindexvaluefromconfigmapkeyref"></a>
#### SKIPJob.spec.container.env[index].valueFrom.configMapKeyRef

<sup>[Parent](#skipjobspeccontainerenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.<br/>This field is effectively required, but due to backwards compatibility is<br/>allowed to be empty. Instances of this type with an empty value here are<br/>almost certainly wrong.<br/>More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: ``<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerenvindexvaluefromfieldref"></a>
#### SKIPJob.spec.container.env[index].valueFrom.fieldRef

<sup>[Parent](#skipjobspeccontainerenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          Path of the field to select in the specified API version.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          Version of the schema the FieldPath is written in terms of, defaults to &#34;v1&#34;.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerenvindexvaluefromfilekeyref"></a>
#### SKIPJob.spec.container.env[index].valueFrom.fileKeyRef

<sup>[Parent](#skipjobspeccontainerenvindexvaluefrom)</sup>

FileKeyRef selects a key of the env file.
Requires the EnvFiles feature gate to be enabled.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key within the env file. An invalid key will prevent the pod from starting.<br/>The keys defined within a source may consist of any printable ASCII characters except &#39;=&#39;.<br/>During Alpha stage of the EnvFiles feature gate, the key size is limited to 128 characters.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path within the volume from which to select the file.<br/>Must be relative and may not contain the &#39;..&#39; path or start with &#39;..&#39;.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          The name of the volume mount containing the env file.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the file or its key must be defined. If the file or key<br/>does not exist, then the env var is not published.<br/>If optional is set to true and the specified key does not exist,<br/>the environment variable will not be set in the Pod&#39;s containers.<br/><br/>If optional is set to false and the specified key does not exist,<br/>an error will be returned during Pod creation.<br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerenvindexvaluefromresourcefieldref"></a>
#### SKIPJob.spec.container.env[index].valueFrom.resourceFieldRef

<sup>[Parent](#skipjobspeccontainerenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Required: resource to select<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Container name: required for volumes, optional for env vars<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          Specifies the output format of the exposed resources, defaults to &#34;1&#34;<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerenvindexvaluefromsecretkeyref"></a>
#### SKIPJob.spec.container.env[index].valueFrom.secretKeyRef

<sup>[Parent](#skipjobspeccontainerenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.<br/>This field is effectively required, but due to backwards compatibility is<br/>allowed to be empty. Instances of this type with an empty value here are<br/>almost certainly wrong.<br/>More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: ``<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerenvfromindex"></a>
#### SKIPJob.spec.container.envFrom[index]

<sup>[Parent](#skipjobspeccontainer)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>configMap</b></td>
        <td>string</td>
        <td>
          Name of Kubernetes ConfigMap in which the deployment should mount environment variables from. Must be in the same namespace as the Application<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          Name of Kubernetes Secret in which the deployment should mount environment variables from. Must be in the same namespace as the Application<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerfilesfromindex"></a>
#### SKIPJob.spec.container.filesFrom[index]

<sup>[Parent](#skipjobspeccontainer)</sup>

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
    <tbody>
      <tr>
        <td><b>mountPath</b></td>
        <td>string</td>
        <td>
          The path to mount the file in the Pods directory. Required.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>configMap</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          defaultMode is optional: mode bits used to set permissions on created files by default.<br/>Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.<br/>YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.<br/>Defaults to 0644.<br/>Directories within the path are not affected by this setting.<br/>This might be in conflict with other options that affect the file<br/>mode, like fsGroup, and the result can be other mode bits set.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>emptyDir</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>persistentVolumeClaim</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainergcp"></a>
#### SKIPJob.spec.container.gcp

<sup>[Parent](#skipjobspeccontainer)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#skipjobspeccontainergcpauth">auth</a></b></td>
        <td>object</td>
        <td>
          Configuration for authenticating a Pod with Google Cloud Platform<br/>For authentication with GCP, to use services like Secret Manager and/or Pub/Sub we need<br/>to set the GCP Service Account Pods should identify as. To allow this, we need the IAM role iam.workloadIdentityUser set on a GCP<br/>service account and bind this to the Pod&#39;s Kubernetes SA.<br/>Documentation on how this is done can be found here (Closed Wiki):<br/>https://kartverket.atlassian.net/wiki/spaces/SKIPDOK/pages/422346824/Autentisering+mot+GCP+som+Kubernetes+SA<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccontainergcpcloudsqlproxy">cloudSqlProxy</a></b></td>
        <td>object</td>
        <td>
          CloudSQL is used to deploy a CloudSQL proxy sidecar in the pod.<br/>This is useful for connecting to CloudSQL databases that require Cloud SQL Auth Proxy.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainergcpauth"></a>
#### SKIPJob.spec.container.gcp.auth

<sup>[Parent](#skipjobspeccontainergcp)</sup>

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
    <tbody>
      <tr>
        <td><b>serviceAccount</b></td>
        <td>string</td>
        <td>
          Name of the service account in which you are trying to authenticate your pod with<br/>Generally takes the form of some-name@some-project-id.iam.gserviceaccount.com<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainergcpcloudsqlproxy"></a>
#### SKIPJob.spec.container.gcp.cloudSqlProxy

<sup>[Parent](#skipjobspeccontainergcp)</sup>

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
    <tbody>
      <tr>
        <td><b>connectionName</b></td>
        <td>string</td>
        <td>
          Connection name for the CloudSQL instance. Found in the Google Cloud Console under your CloudSQL resource.<br/>The format is &#34;projectName:region:instanceName&#34; E.g. &#34;skip-prod-bda1:europe-north1:my-db&#34;.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>ip</b></td>
        <td>string</td>
        <td>
          The IP address of the CloudSQL instance. This is used to create a serviceentry for the CloudSQL proxy.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>serviceAccount</b></td>
        <td>string</td>
        <td>
          Service account used by cloudsql auth proxy. This service account must have the roles/cloudsql.client role.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>publicIP</b></td>
        <td>boolean</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>version</b></td>
        <td>string</td>
        <td>
          Image version for the CloudSQL proxy sidecar.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerliveness"></a>
#### SKIPJob.spec.container.liveness

<sup>[Parent](#skipjobspeccontainer)</sup>

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
    <tbody>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after<br/>having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `3`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that<br/>are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `0`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `10`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.<br/>Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.<br/>Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerpodsettings"></a>
#### SKIPJob.spec.container.podSettings

<sup>[Parent](#skipjobspeccontainer)</sup>

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
    <tbody>
      <tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          Annotations that are set on Pods created by Skiperator. These annotations can for example be used to change the behaviour of sidecars and similar.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>disablePodSpreadTopologyConstraints</b></td>
        <td>boolean</td>
        <td>
          DisablePodSpreadTopologyConstraints specifies whether to disable the addition of Pod Topology Spread Constraints to<br/>a given pod.<br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          TerminationGracePeriodSeconds determines how long Kubernetes waits after a SIGTERM signal sent to a Pod before terminating the pod. If your application uses longer than<br/>30 seconds to terminate, you should increase TerminationGracePeriodSeconds.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Default</i>: `30`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerreadiness"></a>
#### SKIPJob.spec.container.readiness

<sup>[Parent](#skipjobspeccontainer)</sup>

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
    <tbody>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after<br/>having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `3`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that<br/>are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `0`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `10`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.<br/>Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.<br/>Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerresources"></a>
#### SKIPJob.spec.container.resources

<sup>[Parent](#skipjobspeccontainer)</sup>

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
    <tbody>
      <tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits set the maximum the app is allowed to use. Exceeding this limit will<br/>make kubernetes kill the app and restart it.<br/><br/>Limits can be set on the CPU and memory, but it is not recommended to put a limit on CPU, see: https://home.robusta.dev/blog/stop-using-cpu-limits<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests set the initial allocation that is done for the app and will<br/>thus be available to the app on startup. More is allocated on demand<br/>until the limit is reached.<br/><br/>Requests can be set on the CPU and memory.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccontainerstartup"></a>
#### SKIPJob.spec.container.startup

<sup>[Parent](#skipjobspeccontainer)</sup>

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
    <tbody>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after<br/>having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `3`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that<br/>are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `0`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `10`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.<br/>Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.<br/>Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccron"></a>
#### SKIPJob.spec.cron

<sup>[Parent](#skipjobspec)</sup>

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
    <tbody>
      <tr>
        <td><b>schedule</b></td>
        <td>string</td>
        <td>
          A CronJob string for denoting the schedule of this job. See https://crontab.guru/ for help creating CronJob strings.<br/>Kubernetes CronJobs also include the extended &#34;Vixie cron&#34; step values: https://man.freebsd.org/cgi/man.cgi?crontab%285%29.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>allowConcurrency</b></td>
        <td>enum</td>
        <td>
          Denotes how Kubernetes should react to multiple instances of the Job being started at the same time.<br/>Allow will allow concurrent jobs. Forbid will not allow this, and instead skip the newer schedule Job.<br/>Replace will replace the current active Job with the newer scheduled Job.<br/>
          <br/>
            <i>Enum</i>: Allow, Forbid, Replace<br/>
            <i>Default</i>: `Allow`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>startingDeadlineSeconds</b></td>
        <td>integer</td>
        <td>
          Denotes the deadline in seconds for starting a job on its schedule, if for some reason the Job&#39;s controller was not ready upon the scheduled time.<br/>If unset, Jobs missing their deadline will be considered failed jobs and will not start.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          If set to true, this tells Kubernetes to suspend this Job till the field is set to false. If the Job is active while this field is set to true,<br/>all running Pods will be terminated.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeZone</b></td>
        <td>string</td>
        <td>
          The time zone name for the given schedule, see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones. If not specified,<br/>this will default to the time zone of the cluster.<br/><br/>Example: &#34;Europe/Oslo&#34;<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecistiosettings"></a>
#### SKIPJob.spec.istioSettings

<sup>[Parent](#skipjobspec)</sup>

IstioSettings are used to configure istio specific resources such as telemetry. Currently, adjusting sampling
interval for tracing is the only supported option.
By default, tracing is enabled with a random sampling percentage of 10%.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b><a href="#skipjobspecistiosettingstelemetry">telemetry</a></b></td>
        <td>object</td>
        <td>
          Telemetry is a placeholder for all relevant telemetry types, and may be extended in the future to configure additional telemetry settings.<br/>
          <br/>
            <i>Default</i>: `map[tracing:[map[randomSamplingPercentage:10]]]`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecistiosettingstelemetry"></a>
#### SKIPJob.spec.istioSettings.telemetry

<sup>[Parent](#skipjobspecistiosettings)</sup>

Telemetry is a placeholder for all relevant telemetry types, and may be extended in the future to configure additional telemetry settings.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b><a href="#skipjobspecistiosettingstelemetrytracingindex">tracing</a></b></td>
        <td>[]object</td>
        <td>
          Tracing is a list of tracing configurations for the telemetry resource. Normally only one tracing configuration is needed.<br/>
          <br/>
            <i>Default</i>: `[map[randomSamplingPercentage:10]]`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecistiosettingstelemetrytracingindex"></a>
#### SKIPJob.spec.istioSettings.telemetry.tracing[index]

<sup>[Parent](#skipjobspecistiosettingstelemetry)</sup>

Tracing contains relevant settings for tracing in the telemetry configuration

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>randomSamplingPercentage</b></td>
        <td>integer</td>
        <td>
          RandomSamplingPercentage is the percentage of requests that should be sampled for tracing, specified by a whole number between 0-100.<br/>Setting RandomSamplingPercentage to 0 will disable tracing.<br/>
          <br/>
            <i>Default</i>: `10`<br/>
            <i>Minimum</i>: 0<br/>
            <i>Maximum</i>: 100<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecjob"></a>
#### SKIPJob.spec.job

<sup>[Parent](#skipjobspec)</sup>

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
    <tbody>
      <tr>
        <td><b>activeDeadlineSeconds</b></td>
        <td>integer</td>
        <td>
          ActiveDeadlineSeconds denotes a duration in seconds started from when the job is first active. If the deadline is reached during the job&#39;s workload<br/>the job and its Pods are terminated. If the job is suspended using the Suspend field, this timer is stopped and reset when unsuspended.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>backoffLimit</b></td>
        <td>integer</td>
        <td>
          Specifies the number of retry attempts before determining the job as failed. Defaults to 6.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          If set to true, this tells Kubernetes to suspend this Job till the field is set to false. If the Job is active while this field is set to false,<br/>all running Pods will be terminated.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>ttlSecondsAfterFinished</b></td>
        <td>integer</td>
        <td>
          The number of seconds to wait before removing the Job after it has finished. If unset, Job will not be cleaned up.<br/>It is recommended to set this to avoid clutter in your resource tree.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecprometheus"></a>
#### SKIPJob.spec.prometheus

<sup>[Parent](#skipjobspec)</sup>

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
    <tbody>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          The port number or name where metrics are exposed (at the Pod level).<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>allowAllMetrics</b></td>
        <td>boolean</td>
        <td>
          Setting AllowAllMetrics to true will ensure all exposed metrics are scraped. Otherwise, a list of predefined<br/>metrics will be dropped by default. See util/constants.go for the default list.<br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The HTTP path where Prometheus compatible metrics exists<br/>
          <br/>
            <i>Default</i>: `/metrics`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>scrapeInterval</b></td>
        <td>string</td>
        <td>
          ScrapeInterval specifies the interval at which Prometheus should scrape the metrics.<br/>The interval must be at least 15 seconds (if using &#34;Xs&#34;) and divisible by 5.<br/>If minutes (&#34;Xm&#34;) are used, the value must be at least 1m.<br/>
          <br/>
            <i>Default</i>: `60s`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobstatus"></a>
#### SKIPJob.status

<sup>[Parent](#skipjob)</sup>

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
    <tbody>
      <tr>
        <td><b>accessPolicies</b></td>
        <td>string</td>
        <td>
          Indicates if access policies are valid<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobstatussubresourceskey">subresources</a></b></td>
        <td>map[string]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobstatussummary">summary</a></b></td>
        <td>object</td>
        <td>
          Status<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="skipjobstatusconditionsindex"></a>
#### SKIPJob.status.conditions[index]

<sup>[Parent](#skipjobstatus)</sup>

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
    <tbody>
      <tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.<br/>This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.<br/>This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition&#39;s last transition.<br/>Producers of specific condition types may define expected values and meanings for this field,<br/>and whether the values are considered a guaranteed API.<br/>The value should be a CamelCase string.<br/>This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.<br/>For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date<br/>with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobstatussubresourceskey"></a>
#### SKIPJob.status.subresources[key]

<sup>[Parent](#skipjobstatus)</sup>

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
    <tbody>
      <tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Resource accepted by Kubernetes. Waiting for Skiperator to become aware of the resource and start processing.`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Pending`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="skipjobstatussummary"></a>
#### SKIPJob.status.summary

<sup>[Parent](#skipjobstatus)</sup>

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
    <tbody>
      <tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Resource accepted by Kubernetes. Waiting for Skiperator to become aware of the resource and start processing.`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Pending`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
## Package `skiperator.kartverket.no/v1beta1`

Resource types in this package:

- [SKIPJob](#skipjob-1)



<a id="skipjob-1"></a>
### SKIPJob

| Field | Value |
| --- | --- |
| Package | `skiperator.kartverket.no/v1beta1` |
| API version | `skiperator.kartverket.no/v1beta1` |
| Kind | `SKIPJob` |

SKIPJob is the supported schema for the SKIPJobs API.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>skiperator.kartverket.no/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>SKIPJob</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspec-1">spec</a></b></td>
        <td>object</td>
        <td>
          SKIPJobSpec defines the desired state of SKIPJob<br/><br/>A SKIPJob is either defined as a one-off or a scheduled job. If the Cron field is set for SKIPJob, it may not be removed. If the Cron field is unset, it may not be added.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobstatus-1">status</a></b></td>
        <td>object</td>
        <td>
          SkiperatorStatus<br/><br/>A status field shown on a Skiperator resource which contains information regarding deployment of the resource.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspec-1"></a>
#### SKIPJob.spec

<sup>[Parent](#skipjob-1)</sup>

SKIPJobSpec defines the desired state of SKIPJob

A SKIPJob is either defined as a one-off or a scheduled job. If the Cron field is set for SKIPJob, it may not be removed. If the Cron field is unset, it may not be added.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecaccesspolicy">accessPolicy</a></b></td>
        <td>object</td>
        <td>
          AccessPolicy<br/><br/>Zero trust dictates that only applications with a reason for being able<br/>to access another resource should be able to reach it. This is set up by<br/>default by denying all ingress and egress traffic from the Pods in the<br/>Deployment. The AccessPolicy field is an allowlist of other applications and hostnames<br/>that are allowed to talk with this Application and which resources this app can talk to<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecadditionalportsindex">additionalPorts</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspeccron-1">cron</a></b></td>
        <td>object</td>
        <td>
          Settings for the Job if you are running a scheduled job. Optional as Jobs may be one-off.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecenvindex">env</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecenvfromindex">envFrom</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecfilesfromindex">filesFrom</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecgcp">gcp</a></b></td>
        <td>object</td>
        <td>
          GCP<br/><br/>Configuration for interacting with Google Cloud Platform<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecistiosettings-1">istioSettings</a></b></td>
        <td>object</td>
        <td>
          IstioSettings are used to configure istio specific resources such as telemetry. Currently, adjusting sampling<br/>interval for tracing is the only supported option.<br/>By default, tracing is enabled with a random sampling percentage of 10%.<br/>
          <br/>
            <i>Default</i>: `map[telemetry:map[tracing:[map[randomSamplingPercentage:10]]]]`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecjob-1">job</a></b></td>
        <td>object</td>
        <td>
          Settings for the actual Job. If you use a scheduled job, the settings in here will also specify the template of the job.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          Labels can be used if you want every resource created by your SKIPJob to<br/>have the same labels, including the Job/CronJob itself. This could for example be useful for<br/>metrics, where a certain label and the corresponding resources liveliness can be combined.<br/>Any amount of labels can be added as wanted, and they will all cascade down to all resources.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecliveness">liveness</a></b></td>
        <td>object</td>
        <td>
          Probe<br/><br/>Type configuration for all types of Kubernetes probes.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecpodsettings">podSettings</a></b></td>
        <td>object</td>
        <td>
          PodSettings<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>priority</b></td>
        <td>enum</td>
        <td>
          <br/>
          <br/>
            <i>Enum</i>: low, medium, high<br/>
            <i>Default</i>: `medium`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecprometheus-1">prometheus</a></b></td>
        <td>object</td>
        <td>
          Prometheus settings for pod running in job. Fields are identical to Application and if set,<br/>a podmonitoring object is created.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecreadiness">readiness</a></b></td>
        <td>object</td>
        <td>
          Probe<br/><br/>Type configuration for all types of Kubernetes probes.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecresources">resources</a></b></td>
        <td>object</td>
        <td>
          ResourceRequirements<br/><br/>A simplified version of the Kubernetes native ResourceRequirement field, in which only Limits and Requests are present.<br/>For the units used for resources, see https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-units-in-kubernetes<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>restartPolicy</b></td>
        <td>enum</td>
        <td>
          RestartPolicy describes how the container should be restarted.<br/>Only one of the following restart policies may be specified.<br/>If none of the following policies is specified, the default one<br/>is RestartPolicyAlways.<br/>
          <br/>
            <i>Enum</i>: OnFailure, Never<br/>
            <i>Default</i>: `Never`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecstartup">startup</a></b></td>
        <td>object</td>
        <td>
          Probe<br/><br/>Type configuration for all types of Kubernetes probes.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>team</b></td>
        <td>string</td>
        <td>
          Team specifies the team who owns this particular SKIPJob.<br/>Usually sourced from the namespace label.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecaccesspolicy"></a>
#### SKIPJob.spec.accessPolicy

<sup>[Parent](#skipjobspec-1)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#skipjobspecaccesspolicyinbound">inbound</a></b></td>
        <td>object</td>
        <td>
          Inbound specifies the ingress rules. Which apps on the cluster can talk to this app?<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecaccesspolicyoutbound">outbound</a></b></td>
        <td>object</td>
        <td>
          Outbound specifies egress rules. Which apps on the cluster and the<br/>internet is the Application allowed to send requests to?<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecaccesspolicyinbound"></a>
#### SKIPJob.spec.accessPolicy.inbound

<sup>[Parent](#skipjobspecaccesspolicy)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#skipjobspecaccesspolicyinboundrulesindex">rules</a></b></td>
        <td>[]object</td>
        <td>
          The rules list specifies a list of applications. When no namespace is<br/>specified it refers to an app in the current namespace. For apps in<br/>other namespaces namespace is required<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecaccesspolicyinboundrulesindex"></a>
#### SKIPJob.spec.accessPolicy.inbound.rules[index]

<sup>[Parent](#skipjobspecaccesspolicyinbound)</sup>

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
    <tbody>
      <tr>
        <td><b>application</b></td>
        <td>string</td>
        <td>
          The name of the Application you are allowing traffic to/from. If you wish to allow traffic from a SKIPJob, this field should<br/>be suffixed with -skipjob<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          The namespace in which the Application you are allowing traffic to/from resides. If unset, uses namespace of Application.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>namespacesByLabel</b></td>
        <td>map[string]string</td>
        <td>
          Namespace label value-pair in which the Application you are allowing traffic to/from resides. If both namespace and namespacesByLabel are set, namespace takes precedence and namespacesByLabel is omitted.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecaccesspolicyinboundrulesindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above application.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecaccesspolicyinboundrulesindexportsindex"></a>
#### SKIPJob.spec.accessPolicy.inbound.rules[index].ports[index]

<sup>[Parent](#skipjobspecaccesspolicyinboundrulesindex)</sup>

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
    <tbody>
      <tr>
        <td><b>endPort</b></td>
        <td>integer</td>
        <td>
          endPort indicates that the range of ports from port to endPort if set, inclusive,<br/>should be allowed by the policy. This field cannot be defined if the port field<br/>is not defined or if the port field is defined as a named (string) port.<br/>The endPort must be equal or greater than port.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          port represents the port on the given protocol. This can either be a numerical or named<br/>port on a pod. If this field is not provided, this matches all port names and<br/>numbers.<br/>If present, only traffic on the specified protocol AND port will be matched.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          protocol represents the protocol (TCP, UDP, or SCTP) which traffic must match.<br/>If not specified, this field defaults to TCP.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecaccesspolicyoutbound"></a>
#### SKIPJob.spec.accessPolicy.outbound

<sup>[Parent](#skipjobspecaccesspolicy)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#skipjobspecaccesspolicyoutboundexternalindex">external</a></b></td>
        <td>[]object</td>
        <td>
          External specifies which applications on the internet the application<br/>can reach. Only host is required unless it is on another port than HTTPS port 443.<br/>If other ports or protocols are required then `ports` must be specified as well<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecaccesspolicyoutboundrulesindex">rules</a></b></td>
        <td>[]object</td>
        <td>
          Rules apply the same in-cluster rules as InboundPolicy<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecaccesspolicyoutboundexternalindex"></a>
#### SKIPJob.spec.accessPolicy.outbound.external[index]

<sup>[Parent](#skipjobspecaccesspolicyoutbound)</sup>

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
    <tbody>
      <tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          The allowed hostname. Note that this does not include subdomains.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>ip</b></td>
        <td>string</td>
        <td>
          Non-HTTP requests (i.e. using the TCP protocol) need to use IP in addition to hostname<br/>Only required for TCP requests.<br/><br/>Note: Hostname must always be defined even if IP is set statically<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecaccesspolicyoutboundexternalindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above hostname. When not specified HTTP and<br/>HTTPS on port 80 and 443 respectively are put into the allowlist<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecaccesspolicyoutboundexternalindexportsindex"></a>
#### SKIPJob.spec.accessPolicy.outbound.external[index].ports[index]

<sup>[Parent](#skipjobspecaccesspolicyoutboundexternalindex)</sup>

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
    <tbody>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is required and is an arbitrary name. Must be unique within all ExternalRule ports.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          The port number of the external host<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>protocol</b></td>
        <td>enum</td>
        <td>
          The protocol to use for communication with the host. Supported protocols are: HTTP, HTTPS, TCP and TLS.<br/>
          <br/>
            <i>Enum</i>: HTTP, HTTPS, TCP, TLS<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecaccesspolicyoutboundrulesindex"></a>
#### SKIPJob.spec.accessPolicy.outbound.rules[index]

<sup>[Parent](#skipjobspecaccesspolicyoutbound)</sup>

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
    <tbody>
      <tr>
        <td><b>application</b></td>
        <td>string</td>
        <td>
          The name of the Application you are allowing traffic to/from. If you wish to allow traffic from a SKIPJob, this field should<br/>be suffixed with -skipjob<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          The namespace in which the Application you are allowing traffic to/from resides. If unset, uses namespace of Application.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>namespacesByLabel</b></td>
        <td>map[string]string</td>
        <td>
          Namespace label value-pair in which the Application you are allowing traffic to/from resides. If both namespace and namespacesByLabel are set, namespace takes precedence and namespacesByLabel is omitted.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecaccesspolicyoutboundrulesindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The ports to allow for the above application.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecaccesspolicyoutboundrulesindexportsindex"></a>
#### SKIPJob.spec.accessPolicy.outbound.rules[index].ports[index]

<sup>[Parent](#skipjobspecaccesspolicyoutboundrulesindex)</sup>

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
    <tbody>
      <tr>
        <td><b>endPort</b></td>
        <td>integer</td>
        <td>
          endPort indicates that the range of ports from port to endPort if set, inclusive,<br/>should be allowed by the policy. This field cannot be defined if the port field<br/>is not defined or if the port field is defined as a named (string) port.<br/>The endPort must be equal or greater than port.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          port represents the port on the given protocol. This can either be a numerical or named<br/>port on a pod. If this field is not provided, this matches all port names and<br/>numbers.<br/>If present, only traffic on the specified protocol AND port will be matched.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          protocol represents the protocol (TCP, UDP, or SCTP) which traffic must match.<br/>If not specified, this field defaults to TCP.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecadditionalportsindex"></a>
#### SKIPJob.spec.additionalPorts[index]

<sup>[Parent](#skipjobspec-1)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>protocol</b></td>
        <td>enum</td>
        <td>
          Protocol defines network protocols supported for things like container ports.<br/>
          <br/>
            <i>Enum</i>: TCP, UDP, SCTP<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspeccron-1"></a>
#### SKIPJob.spec.cron

<sup>[Parent](#skipjobspec-1)</sup>

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
    <tbody>
      <tr>
        <td><b>schedule</b></td>
        <td>string</td>
        <td>
          A CronJob string for denoting the schedule of this job. See https://crontab.guru/ for help creating CronJob strings.<br/>Kubernetes CronJobs also include the extended &#34;Vixie cron&#34; step values: https://man.freebsd.org/cgi/man.cgi?crontab%285%29.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>allowConcurrency</b></td>
        <td>enum</td>
        <td>
          Denotes how Kubernetes should react to multiple instances of the Job being started at the same time.<br/>Allow will allow concurrent jobs. Forbid will not allow this, and instead skip the newer schedule Job.<br/>Replace will replace the current active Job with the newer scheduled Job.<br/>
          <br/>
            <i>Enum</i>: Allow, Forbid, Replace<br/>
            <i>Default</i>: `Allow`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>startingDeadlineSeconds</b></td>
        <td>integer</td>
        <td>
          Denotes the deadline in seconds for starting a job on its schedule, if for some reason the Job&#39;s controller was not ready upon the scheduled time.<br/>If unset, Jobs missing their deadline will be considered failed jobs and will not start.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          If set to true, this tells Kubernetes to suspend this Job till the field is set to false. If the Job is active while this field is set to true,<br/>all running Pods will be terminated.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeZone</b></td>
        <td>string</td>
        <td>
          The time zone name for the given schedule, see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones. If not specified,<br/>this will default to the time zone of the cluster.<br/><br/>Example: &#34;Europe/Oslo&#34;<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecenvindex"></a>
#### SKIPJob.spec.env[index]

<sup>[Parent](#skipjobspec-1)</sup>

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
    <tbody>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the environment variable.<br/>May consist of any printable ASCII characters except &#39;=&#39;.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Variable references $(VAR_NAME) are expanded<br/>using the previously defined environment variables in the container and<br/>any service environment variables. If a variable cannot be resolved,<br/>the reference in the input string will be unchanged. Double $$ are reduced<br/>to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e.<br/>&#34;$$(VAR_NAME)&#34; will produce the string literal &#34;$(VAR_NAME)&#34;.<br/>Escaped references will never be expanded, regardless of whether the variable<br/>exists or not.<br/>Defaults to &#34;&#34;.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecenvindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Source for the environment variable&#39;s value. Cannot be used if value is not empty.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecenvindexvaluefrom"></a>
#### SKIPJob.spec.env[index].valueFrom

<sup>[Parent](#skipjobspecenvindex)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#skipjobspecenvindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecenvindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels[&#39;&lt;KEY&gt;&#39;]`, `metadata.annotations[&#39;&lt;KEY&gt;&#39;]`,<br/>spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecenvindexvaluefromfilekeyref">fileKeyRef</a></b></td>
        <td>object</td>
        <td>
          FileKeyRef selects a key of the env file.<br/>Requires the EnvFiles feature gate to be enabled.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecenvindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          Selects a resource of the container: only resources limits and requests<br/>(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecenvindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a secret in the pod&#39;s namespace<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecenvindexvaluefromconfigmapkeyref"></a>
#### SKIPJob.spec.env[index].valueFrom.configMapKeyRef

<sup>[Parent](#skipjobspecenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.<br/>This field is effectively required, but due to backwards compatibility is<br/>allowed to be empty. Instances of this type with an empty value here are<br/>almost certainly wrong.<br/>More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: ``<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecenvindexvaluefromfieldref"></a>
#### SKIPJob.spec.env[index].valueFrom.fieldRef

<sup>[Parent](#skipjobspecenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          Path of the field to select in the specified API version.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          Version of the schema the FieldPath is written in terms of, defaults to &#34;v1&#34;.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecenvindexvaluefromfilekeyref"></a>
#### SKIPJob.spec.env[index].valueFrom.fileKeyRef

<sup>[Parent](#skipjobspecenvindexvaluefrom)</sup>

FileKeyRef selects a key of the env file.
Requires the EnvFiles feature gate to be enabled.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key within the env file. An invalid key will prevent the pod from starting.<br/>The keys defined within a source may consist of any printable ASCII characters except &#39;=&#39;.<br/>During Alpha stage of the EnvFiles feature gate, the key size is limited to 128 characters.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path within the volume from which to select the file.<br/>Must be relative and may not contain the &#39;..&#39; path or start with &#39;..&#39;.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          The name of the volume mount containing the env file.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the file or its key must be defined. If the file or key<br/>does not exist, then the env var is not published.<br/>If optional is set to true and the specified key does not exist,<br/>the environment variable will not be set in the Pod&#39;s containers.<br/><br/>If optional is set to false and the specified key does not exist,<br/>an error will be returned during Pod creation.<br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecenvindexvaluefromresourcefieldref"></a>
#### SKIPJob.spec.env[index].valueFrom.resourceFieldRef

<sup>[Parent](#skipjobspecenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Required: resource to select<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          Container name: required for volumes, optional for env vars<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          Specifies the output format of the exposed resources, defaults to &#34;1&#34;<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecenvindexvaluefromsecretkeyref"></a>
#### SKIPJob.spec.env[index].valueFrom.secretKeyRef

<sup>[Parent](#skipjobspecenvindexvaluefrom)</sup>

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
    <tbody>
      <tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.<br/>This field is effectively required, but due to backwards compatibility is<br/>allowed to be empty. Instances of this type with an empty value here are<br/>almost certainly wrong.<br/>More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: ``<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecenvfromindex"></a>
#### SKIPJob.spec.envFrom[index]

<sup>[Parent](#skipjobspec-1)</sup>



<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>configMap</b></td>
        <td>string</td>
        <td>
          Name of Kubernetes ConfigMap in which the deployment should mount environment variables from. Must be in the same namespace as the Application<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          Name of Kubernetes Secret in which the deployment should mount environment variables from. Must be in the same namespace as the Application<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecfilesfromindex"></a>
#### SKIPJob.spec.filesFrom[index]

<sup>[Parent](#skipjobspec-1)</sup>

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
    <tbody>
      <tr>
        <td><b>mountPath</b></td>
        <td>string</td>
        <td>
          The path to mount the file in the Pods directory. Required.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>configMap</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          defaultMode is optional: mode bits used to set permissions on created files by default.<br/>Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.<br/>YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.<br/>Defaults to 0644.<br/>Directories within the path are not affected by this setting.<br/>This might be in conflict with other options that affect the file<br/>mode, like fsGroup, and the result can be other mode bits set.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>emptyDir</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>persistentVolumeClaim</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecgcp"></a>
#### SKIPJob.spec.gcp

<sup>[Parent](#skipjobspec-1)</sup>

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
    <tbody>
      <tr>
        <td><b><a href="#skipjobspecgcpauth">auth</a></b></td>
        <td>object</td>
        <td>
          Configuration for authenticating a Pod with Google Cloud Platform<br/>For authentication with GCP, to use services like Secret Manager and/or Pub/Sub we need<br/>to set the GCP Service Account Pods should identify as. To allow this, we need the IAM role iam.workloadIdentityUser set on a GCP<br/>service account and bind this to the Pod&#39;s Kubernetes SA.<br/>Documentation on how this is done can be found here (Closed Wiki):<br/>https://kartverket.atlassian.net/wiki/spaces/SKIPDOK/pages/422346824/Autentisering+mot+GCP+som+Kubernetes+SA<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobspecgcpcloudsqlproxy">cloudSqlProxy</a></b></td>
        <td>object</td>
        <td>
          CloudSQL is used to deploy a CloudSQL proxy sidecar in the pod.<br/>This is useful for connecting to CloudSQL databases that require Cloud SQL Auth Proxy.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecgcpauth"></a>
#### SKIPJob.spec.gcp.auth

<sup>[Parent](#skipjobspecgcp)</sup>

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
    <tbody>
      <tr>
        <td><b>serviceAccount</b></td>
        <td>string</td>
        <td>
          Name of the service account in which you are trying to authenticate your pod with<br/>Generally takes the form of some-name@some-project-id.iam.gserviceaccount.com<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecgcpcloudsqlproxy"></a>
#### SKIPJob.spec.gcp.cloudSqlProxy

<sup>[Parent](#skipjobspecgcp)</sup>

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
    <tbody>
      <tr>
        <td><b>connectionName</b></td>
        <td>string</td>
        <td>
          Connection name for the CloudSQL instance. Found in the Google Cloud Console under your CloudSQL resource.<br/>The format is &#34;projectName:region:instanceName&#34; E.g. &#34;skip-prod-bda1:europe-north1:my-db&#34;.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>ip</b></td>
        <td>string</td>
        <td>
          The IP address of the CloudSQL instance. This is used to create a serviceentry for the CloudSQL proxy.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>serviceAccount</b></td>
        <td>string</td>
        <td>
          Service account used by cloudsql auth proxy. This service account must have the roles/cloudsql.client role.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>publicIP</b></td>
        <td>boolean</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>version</b></td>
        <td>string</td>
        <td>
          Image version for the CloudSQL proxy sidecar.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecistiosettings-1"></a>
#### SKIPJob.spec.istioSettings

<sup>[Parent](#skipjobspec-1)</sup>

IstioSettings are used to configure istio specific resources such as telemetry. Currently, adjusting sampling
interval for tracing is the only supported option.
By default, tracing is enabled with a random sampling percentage of 10%.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b><a href="#skipjobspecistiosettingstelemetry-1">telemetry</a></b></td>
        <td>object</td>
        <td>
          Telemetry is a placeholder for all relevant telemetry types, and may be extended in the future to configure additional telemetry settings.<br/>
          <br/>
            <i>Default</i>: `map[tracing:[map[randomSamplingPercentage:10]]]`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecistiosettingstelemetry-1"></a>
#### SKIPJob.spec.istioSettings.telemetry

<sup>[Parent](#skipjobspecistiosettings-1)</sup>

Telemetry is a placeholder for all relevant telemetry types, and may be extended in the future to configure additional telemetry settings.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b><a href="#skipjobspecistiosettingstelemetrytracingindex-1">tracing</a></b></td>
        <td>[]object</td>
        <td>
          Tracing is a list of tracing configurations for the telemetry resource. Normally only one tracing configuration is needed.<br/>
          <br/>
            <i>Default</i>: `[map[randomSamplingPercentage:10]]`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecistiosettingstelemetrytracingindex-1"></a>
#### SKIPJob.spec.istioSettings.telemetry.tracing[index]

<sup>[Parent](#skipjobspecistiosettingstelemetry-1)</sup>

Tracing contains relevant settings for tracing in the telemetry configuration

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody>
      <tr>
        <td><b>randomSamplingPercentage</b></td>
        <td>integer</td>
        <td>
          RandomSamplingPercentage is the percentage of requests that should be sampled for tracing, specified by a whole number between 0-100.<br/>Setting RandomSamplingPercentage to 0 will disable tracing.<br/>
          <br/>
            <i>Default</i>: `10`<br/>
            <i>Minimum</i>: 0<br/>
            <i>Maximum</i>: 100<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecjob-1"></a>
#### SKIPJob.spec.job

<sup>[Parent](#skipjobspec-1)</sup>

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
    <tbody>
      <tr>
        <td><b>activeDeadlineSeconds</b></td>
        <td>integer</td>
        <td>
          ActiveDeadlineSeconds denotes a duration in seconds started from when the job is first active. If the deadline is reached during the job&#39;s workload<br/>the job and its Pods are terminated. If the job is suspended using the Suspend field, this timer is stopped and reset when unsuspended.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>backoffLimit</b></td>
        <td>integer</td>
        <td>
          Specifies the number of retry attempts before determining the job as failed. Defaults to 6.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          If set to true, this tells Kubernetes to suspend this Job till the field is set to false. If the Job is active while this field is set to false,<br/>all running Pods will be terminated.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>ttlSecondsAfterFinished</b></td>
        <td>integer</td>
        <td>
          The number of seconds to wait before removing the Job after it has finished. If unset, Job will not be cleaned up.<br/>It is recommended to set this to avoid clutter in your resource tree.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecliveness"></a>
#### SKIPJob.spec.liveness

<sup>[Parent](#skipjobspec-1)</sup>

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
    <tbody>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after<br/>having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `3`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that<br/>are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `0`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `10`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.<br/>Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.<br/>Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecpodsettings"></a>
#### SKIPJob.spec.podSettings

<sup>[Parent](#skipjobspec-1)</sup>

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
    <tbody>
      <tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          Annotations that are set on Pods created by Skiperator. These annotations can for example be used to change the behaviour of sidecars and similar.<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>disablePodSpreadTopologyConstraints</b></td>
        <td>boolean</td>
        <td>
          DisablePodSpreadTopologyConstraints specifies whether to disable the addition of Pod Topology Spread Constraints to<br/>a given pod.<br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          TerminationGracePeriodSeconds determines how long Kubernetes waits after a SIGTERM signal sent to a Pod before terminating the pod. If your application uses longer than<br/>30 seconds to terminate, you should increase TerminationGracePeriodSeconds.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Default</i>: `30`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecprometheus-1"></a>
#### SKIPJob.spec.prometheus

<sup>[Parent](#skipjobspec-1)</sup>

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
    <tbody>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          The port number or name where metrics are exposed (at the Pod level).<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>allowAllMetrics</b></td>
        <td>boolean</td>
        <td>
          Setting AllowAllMetrics to true will ensure all exposed metrics are scraped. Otherwise, a list of predefined<br/>metrics will be dropped by default. See util/constants.go for the default list.<br/>
          <br/>
            <i>Default</i>: `false`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The HTTP path where Prometheus compatible metrics exists<br/>
          <br/>
            <i>Default</i>: `/metrics`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>scrapeInterval</b></td>
        <td>string</td>
        <td>
          ScrapeInterval specifies the interval at which Prometheus should scrape the metrics.<br/>The interval must be at least 15 seconds (if using &#34;Xs&#34;) and divisible by 5.<br/>If minutes (&#34;Xm&#34;) are used, the value must be at least 1m.<br/>
          <br/>
            <i>Default</i>: `60s`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecreadiness"></a>
#### SKIPJob.spec.readiness

<sup>[Parent](#skipjobspec-1)</sup>

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
    <tbody>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after<br/>having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `3`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that<br/>are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `0`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `10`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.<br/>Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.<br/>Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecresources"></a>
#### SKIPJob.spec.resources

<sup>[Parent](#skipjobspec-1)</sup>

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
    <tbody>
      <tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits set the maximum the app is allowed to use. Exceeding this limit will<br/>make kubernetes kill the app and restart it.<br/><br/>Limits can be set on the CPU and memory, but it is not recommended to put a limit on CPU, see: https://home.robusta.dev/blog/stop-using-cpu-limits<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests set the initial allocation that is done for the app and will<br/>thus be available to the app on startup. More is allocated on demand<br/>until the limit is reached.<br/><br/>Requests can be set on the CPU and memory.<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobspecstartup"></a>
#### SKIPJob.spec.startup

<sup>[Parent](#skipjobspec-1)</sup>

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
    <tbody>
      <tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          The path to access on the HTTP server<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          Number of the port to access on the container<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive failures for the probe to be considered failed after<br/>having succeeded. Defaults to 3. Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `3`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>initialDelay</b></td>
        <td>integer</td>
        <td>
          Delay sending the first probe by X seconds. Can be useful for applications that<br/>are slow to start.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `0`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>period</b></td>
        <td>integer</td>
        <td>
          Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `10`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          Minimum consecutive successes for the probe to be considered successful after having failed.<br/>Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
      <tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          Number of seconds after which the probe times out. Defaults to 1 second.<br/>Minimum value is 1<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: `1`<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobstatus-1"></a>
#### SKIPJob.status

<sup>[Parent](#skipjob-1)</sup>

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
    <tbody>
      <tr>
        <td><b>accessPolicies</b></td>
        <td>string</td>
        <td>
          Indicates if access policies are valid<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobstatusconditionsindex-1">conditions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobstatussubresourceskey-1">subresources</a></b></td>
        <td>map[string]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b><a href="#skipjobstatussummary-1">summary</a></b></td>
        <td>object</td>
        <td>
          Status<br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="skipjobstatusconditionsindex-1"></a>
#### SKIPJob.status.conditions[index]

<sup>[Parent](#skipjobstatus-1)</sup>

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
    <tbody>
      <tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.<br/>This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.<br/>This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition&#39;s last transition.<br/>Producers of specific condition types may define expected values and meanings for this field,<br/>and whether the values are considered a guaranteed API.<br/>The value should be a CamelCase string.<br/>This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.<br/>For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date<br/>with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr>
    </tbody>
</table>
<a id="skipjobstatussubresourceskey-1"></a>
#### SKIPJob.status.subresources[key]

<sup>[Parent](#skipjobstatus-1)</sup>

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
    <tbody>
      <tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Resource accepted by Kubernetes. Waiting for Skiperator to become aware of the resource and start processing.`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Pending`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
<a id="skipjobstatussummary-1"></a>
#### SKIPJob.status.summary

<sup>[Parent](#skipjobstatus-1)</sup>

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
    <tbody>
      <tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Resource accepted by Kubernetes. Waiting for Skiperator to become aware of the resource and start processing.`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: `Pending`<br/>
        </td>
        <td>true</td>
      </tr>
      <tr>
        <td><b>timestamp</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr>
    </tbody>
</table>
