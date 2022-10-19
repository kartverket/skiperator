resource "kubernetes_config_map" "gcp-identity-config" {
  metadata {
    name      = "gcp-identity-config"
    namespace = "skiperator-system"
  }
  data = {
    workloadIdentityPool = var.workloadIdentityPool
    identityProvider     = var.identityProvider
  }
}