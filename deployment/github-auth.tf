data "google_secret_manager_secret_version" "github_auth" {
  project = "bootstrap-349108"
  secret  = "github-auth"
}

resource "kubernetes_secret_v1" "github_auth" {
  metadata {
    namespace = "skiperator-system"
    name      = "skiperator"
  }

  data = {
    token = data.google_secret_manager_secret_version.github_auth.secret_data
  }
}
