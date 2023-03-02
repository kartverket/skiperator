resource "kubernetes_config_map" "namespace-exclusions-map" {
  metadata {
    name      = "namespace-exclusions"
    namespace = "skiperator-system"
  }
  data = {
    # Networking
    istio-system   = "true"
    istio-gateways = "true"
    asm-system     = "true"
    cert-manager   = "true"

    # Kubernetes Systems
    kube-node-lease = "true"
    kube-public     = "true"
    kube-system     = "true"
    default         = "true"

    # Anthos Systems
    anthos-identity-service      = "true"
    config-management-system     = "true"
    config-management-monitoring = "true"
    gke-connect                  = "true"
    gke-system                   = "true"
    gke-managed-metrics-server   = "true"
    resource-group-system        = "true"
    gmp-public                   = "true"
    gmp-system                   = "true"
    gmp-test                     = "true"

    # SKIP Systems
    binauthz-system             = "true"
    gatekeeper-system           = "true"
    skiperator-system           = "true"
    sysdig-agent                = "true"
    sysdig-admission-controller = "true"
    instana-agent               = "true"
    instana-autotrace-webhook   = "true"
    vault                       = "true"

    # Argo
    argocd            = "true"
    crossplane-system = "true"
    upbound-system    = "true"
    external-secrets  = "true"

    # PoC, to be removed?
    istio-shipwreck = "true"
    kasten-io       = "true"
    fluentd         = "true"
    kubecost        = "true"
  }
}
