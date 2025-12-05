##############################
# Module: s3_bucket
# Description: Creates a securely configured Amazon S3 bucket with optional features such as
#              versioning, policies and CORS.
##############################

locals {
  # Allow public access only if public read is enabled and domains are provided
  allow_public_access = var.public_read.enabled && length(var.public_read.allowed_domains) > 0
}

# -----------------------------
# Main S3 bucket
# -----------------------------
resource "aws_s3_bucket" "this" {
  bucket = var.bucket_name
  tags   = var.tags

  force_destroy = var.allow_force_destroy
}

resource "aws_s3_bucket_versioning" "versioning" {
  bucket = aws_s3_bucket.this.id
  versioning_configuration {
    status = var.versioning_enabled ? "Enabled" : "Disabled"
  }
}

resource "aws_s3_bucket_public_access_block" "public_access_block" {
  count  = local.allow_public_access ? 1 : 0

  bucket = aws_s3_bucket.this.id

  block_public_acls       = true
  ignore_public_acls      = true
  block_public_policy     = !(local.allow_public_access)
  restrict_public_buckets = !(local.allow_public_access)
}

# -----------------------------
# Conditionally create GET policy if enabled and domains are provided
# -----------------------------
resource "aws_s3_bucket_policy" "get_policy" {
  count  = local.allow_public_access ? 1 : 0
  bucket = aws_s3_bucket.this.id

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "PolicyForGetFromAllowedDomains"
    Statement = [
      {
        Sid       = "AllowGetFromAllowedDomains"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.this.arn}/*"
        Condition = {
          StringLike = {
            "aws:Referer" = var.public_read.allowed_domains
          }
        }
      }
    ]
  })

  depends_on = [aws_s3_bucket_public_access_block.public_access_block]
}

# -----------------------------
# Conditionally create PUT CORS if enabled and domains are provided
# -----------------------------
resource "aws_s3_bucket_cors_configuration" "put_cors" {
  count  = var.put_cors.enabled && length(var.put_cors.allowed_domains) > 0 ? 1 : 0
  bucket = aws_s3_bucket.this.id

  cors_rule {
    allowed_methods = ["PUT"]
    allowed_origins = var.put_cors.allowed_domains
    allowed_headers = var.put_cors.allowed_headers
    max_age_seconds = var.put_cors.max_age_seconds
  }
}
