variable "image" {
  type = string
}

resource "kubernetes_deployment_v1" "deployment" {
  metadata {
    namespace = "skiperator-system"
    name      = "skiperator"
  }
  spec {
    replicas = 3
    selector {
      match_labels = {
        app = "skiperator"
      }
    }
    template {
      metadata {
        labels = {
          app = "skiperator"
        }
      }
      spec {
        service_account_name = "skiperator"
        container {
          name  = "skiperator"
          image = var.image
          resources {
            limits = {
              cpu    = "0.2"
              memory = "64Mi"
            }
          }
          port {
            name           = "metrics"
            container_port = 8080
          }
          port {
            name           = "probes"
            container_port = 8081
          }
          liveness_probe {
            http_get {
              path = "/healtz"
              port = "probes"
            }
          }
          readiness_probe {
            http_get {
              path = "/readyz"
              port = "probes"
            }
          }
        }
      }
    }
  }
}