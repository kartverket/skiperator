resource "kubernetes_service_v1" "service" {
  metadata {
    namespace = kubernetes_namespace_v1.namespace.metadata[0].name
    name      = kubernetes_deployment_v1.deployment.metadata[0].name
  }
  spec {
    selector = kubernetes_deployment_v1.deployment.spec[0].selector[0].match_labels
    port {
      port        = 8080
      target_port = "metrics"
    }
  }
}