provider "google-beta" {
  project = var.project
  region  = var.region
}

resource "google_project_service" "enabled_services" {
  project            = var.project
  service            = each.key
  for_each           = toset(["run.googleapis.com"])
  disable_on_destroy = false

}

resource "google_service_account" "demo_cloud_run_service_account" {
  project      = var.project
  account_id   = var.service_account_id
  display_name = "cloud run bin authz demo service account"
}

#Creates a Cloud Run service and deploys an attested image
resource "google_cloud_run_service" "demo_cloud_run_service" {
  name     = var.service
  provider = google-beta
  location = var.region
  project  = var.project
  metadata {
    annotations = {
      "run.googleapis.com/ingress" = "all"
      "run.googleapis.com/binary-authorization" : "default"
    }
  }
  template {
    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale"  = "1000"
        "autoscaling.knative.dev/min-scale" = "3"
      }
    }
    spec {
      service_account_name = google_service_account.demo_cloud_run_service_account.email
      containers {
        image = "gcr.io/${var.project}/product-listing-api:binauth"
      }
    }
  }
  traffic {
    percent         = 100
    latest_revision = true
  }
  depends_on = [
    google_project_service.enabled_services
  ]
}

#Assign invoker role to the account/user who's token will be used for authentication and authorization by Cloud Run
resource "google_cloud_run_service_iam_member" "developer_invoker_member" {
  location = google_cloud_run_service.demo_cloud_run_service.location
  project  = google_cloud_run_service.demo_cloud_run_service.project
  service  = google_cloud_run_service.demo_cloud_run_service.name
  role     = "roles/run.invoker"
  member   = var.developer
}