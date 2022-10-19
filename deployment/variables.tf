variable "workloadIdentityPool" {
  type        = string
  default     = "empty"
  description = "WorkloadIdentityPool for kubernetes cluster fleet. Looks like this: FLEET_PROJECT_ID.svc.id.goog"
}

variable "identityProvider" {
  type        = string
  default     = "empty"
  description = "The name of the identity provider associated with your Kubernetes cluster"
}