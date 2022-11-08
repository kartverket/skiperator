resource "kubernetes_config_map" "instana-networkpolicy-config" {
  metadata {
    name      = "instana-networkpolicy-config"
    namespace = "skiperator-system"
  }
  data = {
    cidrBlock = var.instanaCidrBlock
  }
}