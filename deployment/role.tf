resource "kubernetes_service_account_v1" "service_account" {
  metadata {
    namespace = "skiperator-system"
    name      = "skiperator"
  }
}

resource "kubernetes_cluster_role_binding_v1" "cluster_role_binding" {
  metadata {
    name = "skiperator"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "skiperator"
  }
  subject {
    kind      = "ServiceAccount"
    namespace = "skiperator-system"
    name      = "skiperator"
  }
}