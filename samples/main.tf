# Deploy sample application locally using terraform
# Uses local state and the default kind-kind kubernetes context
provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = "kind-kind"
}

resource "kubernetes_manifest" "nginx_config" {
  manifest = yamldecode(file("nginx_config.yaml"))
}

resource "kubernetes_manifest" "application" {
  manifest = yamldecode(file("application.yaml"))
  field_manager {
    force_conflicts = true
  }
}
