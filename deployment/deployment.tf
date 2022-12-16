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
        annotations = {
          prometheus.io/scrape = true
          prometheus.io/port   = "8080"
        }
      }
      spec {
        service_account_name = "skiperator"
        container {
          name  = "skiperator"
          image = var.image
          args  = ["-l", "-d", "-t=$(IMAGE_PULL_TOKEN)"]
          env {
            name = "IMAGE_PULL_TOKEN"
            value_from {
              secret_key_ref {
                name = kubernetes_secret_v1.github_auth.metadata[0].name
                key  = "token"
              }
            }
          }
          security_context {
            read_only_root_filesystem  = true
            allow_privilege_escalation = false
            run_as_user                = "65532"
            run_as_group               = "65532"
            seccomp_profile { type = "RuntimeDefault" }
          }
          resources {
            requests = {
              cpu    = "10m"
              memory = "32Mi"
            }
            limits = {
              memory = "256Mi"
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
              path = "/healthz"
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
