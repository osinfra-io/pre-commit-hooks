# Google Cloud Storage Bucket Resource
# https://www.terraform.io/docs/providers/google/r/storage_bucket.html

resource "google_storage_bucket" "this" {
  name          = "test-bucket"
  location      = var.location
  force_destroy = true
}
