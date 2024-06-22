# Google Cloud Storage Bucket Resource
# https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/storage_bucket.html

resource "google_storage_bucket" "bucket" {
  name          = "my-tf-test-bucket"
  location      = "US"
  force_destroy = true
}
