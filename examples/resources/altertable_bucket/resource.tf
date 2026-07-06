# An AWS S3 bucket (no endpoint needed)
resource "altertable_bucket" "landing" {
  environment_id    = altertable_environment.example.id
  name              = "acme-landing-zone"
  region            = "eu-west-1"
  access_key_id     = var.aws_access_key_id
  secret_access_key = var.aws_secret_access_key
}

# A Cloudflare R2 bucket (S3-compatible endpoint)
resource "altertable_bucket" "archive" {
  environment_id    = altertable_environment.example.id
  name              = "acme-archive"
  endpoint          = "https://f7c1a2b3d4e5f6a7b8c9d0e1f2a3b4c5.r2.cloudflarestorage.com"
  access_key_id     = var.r2_access_key_id
  secret_access_key = var.r2_secret_access_key
}
