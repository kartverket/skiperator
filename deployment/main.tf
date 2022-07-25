terraform {
  backend "gcs" {}
}

provider "kubernetes" {}

resource "kubernetes_namespace_v1" "namespace" {
  metadata {
    name = "skiperator-system"
  }
}

resource "kubernetes_manifest" "custom_resource_definition" {
  manifest = yamldecode(file("${path.module}/skiperator.kartverket.no_applications.yaml"))
}

resource "kubernetes_manifest" "cluster_role" {
  manifest = yamldecode(file("${path.module}/role.yaml"))
}