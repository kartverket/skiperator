terraform {
  backend "kubernetes" {
    secret_suffix = "skiperator"
    namespace     = "skiperator-system"
    config_path   = "~/.kube/config"
  }
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

resource "kubernetes_manifest" "custom_resource_definition" {
  manifest = yamldecode(file("${path.module}/skiperator.kartverket.no_applications.yaml"))
}

resource "kubernetes_manifest" "cluster_role" {
  manifest = yamldecode(file("${path.module}/role.yaml"))
}