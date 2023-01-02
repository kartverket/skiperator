data "google_secret_manager_secret_version" "github_auth" {
  project = "994831889648"
  secret  = "github-auth"
}

resource "kubernetes_secret_v1" "github_auth" {
  metadata {
    namespace = "skiperator-system"
    name      = "github-auth"
  }

  data = {
    token = data.google_secret_manager_secret_version.github_auth.secret_data
  }

  wait_for_service_account_token = false
}
