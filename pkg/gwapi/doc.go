// Package gwapi contains controller-side support for Kubernetes Gateway API
// routing.
//
// This package does not generate Gateway API resources. Resource generation
// lives in pkg/resourcegenerator/gatewayapi. The code here answers questions
// the Application and Routing reconcilers must answer before and after
// generation:
//   - Is it safe to create the requested Gateway API resources?
//   - Are the generated resources accepted and programmed by Istio?
//   - Should legacy Istio Gateway and VirtualService resources still be kept?
//   - Which status conditions and migration events should be reported?
//
// Legacy Istio routing uses a Skiperator-owned Istio Gateway and VirtualService.
// Standard routing uses Kubernetes Gateway API resources instead. Skiperator
// assumes shared Gateway objects already exist in the istio-gateways namespace.
// Each Application or Routing object then gets:
//   - a ListenerSet for each hostname it owns, attached to the shared Gateway
//   - one or more HTTPRoutes for redirect and backend routing rules
//   - cert-manager Certificates and TLS Secrets in the application namespace
//
// The Gateway API controller, Istio in this case, is responsible for accepting
// those resources and programming Envoy. Skiperator waits for that status before
// pruning legacy Istio resources. This is the zero-downtime migration rule:
// legacy routing remains active until standard routing is ready, unless the
// object is greenfield and has no legacy resources to preserve.
//
// Hostname ownership differs between Application and Routing. Application owns
// a whole hostname through ListenerSet listeners. Routing can share a hostname
// with other teams, so ownership is per hostname and path prefix. Accepted
// resources win; pending or conflicting resources are surfaced as status errors
// instead of relying on last-writer-wins behavior.
package gwapi
