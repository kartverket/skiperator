resource "kubernetes_service_account_v1" "service_account" {
  metadata {
    namespace = "skiperator-system"
    name      = kubernetes_deployment_v1.deployment.metadata[0].name
  }
}

resource "kubernetes_cluster_role_binding_v1" "cluster_role_binding" {
  metadata {
    name = kubernetes_deployment_v1.deployment.metadata[0].name
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "skiperator"
  }
  subject {
    kind      = "ServiceAccount"
    namespace = "skiperator-system"
    name      = kubernetes_deployment_v1.deployment.metadata[0].name
  }
}