resource "kubernetes_service_v1" "service" {
  metadata {
    namespace = "skiperator-system"
    name      = "skiperator"
  }
  spec {
    selector = kubernetes_deployment_v1.deployment.spec[0].selector[0].match_labels
    port {
      port        = 8080
      target_port = "metrics"
    }
  }
}