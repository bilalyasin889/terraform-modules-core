##############################
# Module: cloudfront
# Description: Creates a CloudFront distribution with configurable origins, caching rules, SSL support, and optional custom domains.
##############################

# -----------------------------
# CloudFront Origin Access Control for S3 Buckets
# -----------------------------
resource "aws_cloudfront_origin_access_control" "s3_oac" {
  for_each = {
    for key, origin in var.origins :
    key => origin
    if lower(origin.type) == "s3"
  }

  name                              = "${each.key}-oac"
  description                       = "Origin Access Control for access private S3 bucket ${each.value.bucket_id}"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# -----------------------------
# S3 Bucket Policies for CloudFront to read objects
# -----------------------------
resource "aws_s3_bucket_policy" "cloudfront_access" {
  for_each = {
    for k, o in var.origins :
    k => o
    if lower(o.type) == "s3"
  }

  bucket = each.value.bucket_id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid      = "AllowCloudFrontServicePrincipal"
        Effect   = "Allow"
        Action   = "s3:GetObject"
        Resource = "${each.value.bucket_arn}/*"
        Principal = {
          Service = "cloudfront.amazonaws.com"
        }
        Condition = {
          StringEquals = {
            "AWS:SourceArn" = aws_cloudfront_distribution.this.arn
          }
        }
      }
    ]
  })
}

# -----------------------------
# Function(s)
# -----------------------------
resource "aws_cloudfront_function" "this" {
  for_each = var.cloudfront_functions

  name                         = try(coalesce(each.value.name, each.key))
  code                         = trimspace(each.value.code)
  comment                      = each.value.comment
  key_value_store_associations = each.value.key_value_store_associations
  publish                      = each.value.publish
  runtime                      = each.value.runtime
}

# -----------------------------
# CloudFront Distribution
# -----------------------------
resource "aws_cloudfront_distribution" "this" {
  enabled             = var.enabled
  is_ipv6_enabled     = true
  comment             = "CloudFront distribution for ${var.name}"
  default_root_object = var.default_root_object
  aliases             = var.domains
  tags                = var.tags
  price_class         = var.price_class

  dynamic "origin" {
    for_each = var.origins

    content {
      origin_id   = origin.key
      domain_name = origin.value.domain_name
      origin_path = origin.value.origin_path

      # Attach OAC only for S3 origins
      origin_access_control_id = (
        contains(keys(aws_cloudfront_origin_access_control.s3_oac), origin.key)
        ? aws_cloudfront_origin_access_control.s3_oac[origin.key].id
        : null
      )

      # Non-S3 only
      dynamic "custom_origin_config" {
        for_each = [origin.value.type] != ["s3"] ? [1] : []
        content {
          http_port              = 80
          https_port             = 443
          origin_protocol_policy = "https-only"
          origin_ssl_protocols   = ["TLSv1.2"]
        }
      }

      dynamic "custom_header" {
        for_each = origin.value.custom_headers
        content {
          name  = custom_header.key
          value = custom_header.value
        }
      }
    }
  }

  default_cache_behavior {
    allowed_methods        = var.default_cache_behavior.allowed_methods
    cached_methods         = var.default_cache_behavior.cached_methods
    target_origin_id       = var.default_cache_behavior.origin_id
    viewer_protocol_policy = var.default_cache_behavior.viewer_protocol_policy

    # Apply functions
    dynamic "function_association" {
      for_each = var.default_cache_behavior.functions
      content {
        event_type   = function_association.value.event_type
        function_arn = try(coalesce(function_association.value.arn, try(aws_cloudfront_function.this[function_association.value.function_key].arn, null)), null)
      }
    }

    cache_policy_id          = var.default_cache_behavior.cache_policy_id
    origin_request_policy_id = var.default_cache_behavior.origin_request_policy_id
  }

  dynamic "ordered_cache_behavior" {
    for_each = var.ordered_cache_behaviors

    content {
      path_pattern     = ordered_cache_behavior.value.path_pattern
      target_origin_id = ordered_cache_behavior.value.target_origin_id
      allowed_methods  = ordered_cache_behavior.value.allowed_methods
      cached_methods   = ordered_cache_behavior.value.cached_methods

      viewer_protocol_policy   = ordered_cache_behavior.value.viewer_protocol_policy

      cache_policy_id          = ordered_cache_behavior.value.cache_policy_id
      origin_request_policy_id = ordered_cache_behavior.value.origin_request_policy_id

      # Apply functions
      dynamic "function_association" {
        for_each = ordered_cache_behavior.value.functions
        content {
          event_type   = function_association.value.event_type
          function_arn = try(coalesce(function_association.value.arn, try(aws_cloudfront_function.this[function_association.value.function_key].arn, null)), null)
        }
      }
    }
  }

  dynamic "custom_error_response" {
    for_each = var.custom_error_responses

    content {
      error_caching_min_ttl = custom_error_response.value.error_caching_min_ttl
      error_code            = custom_error_response.value.error_code
      response_code         = custom_error_response.value.response_code
      response_page_path    = custom_error_response.value.response_page_path
    }
  }

  # Restrict distribution to specified location(s) or "none"
  restrictions {
    geo_restriction {
      restriction_type = var.restrictions != null ? var.restrictions.type : "none"
      locations        = var.restrictions != null ? var.restrictions.locations : []
    }
  }

  # -----------------------------
  # Viewer Certificate - Ensure custom certificate is used to allow aliases
  # -----------------------------
  viewer_certificate {
    cloudfront_default_certificate = false
    acm_certificate_arn            = var.certificate_arn
    ssl_support_method             = "sni-only"
    minimum_protocol_version       = "TLSv1.2_2021"
  }
}

# A record (IPv4)
resource "aws_route53_record" "cf_aliases" {
  for_each = var.hosted_zone_id != null ? { for d in var.domains : d => d } : {}

  zone_id = var.hosted_zone_id
  name    = each.value
  type    = "A"

  alias {
    name                   = aws_cloudfront_distribution.this.domain_name
    zone_id                = aws_cloudfront_distribution.this.hosted_zone_id
    evaluate_target_health = false
  }
}

# AAAA record (IPv6)
resource "aws_route53_record" "cf_aliases_ipv6" {
  for_each = var.hosted_zone_id != null ? { for d in var.domains : d => d } : {}

  zone_id = var.hosted_zone_id
  name    = each.value
  type    = "AAAA"

  alias {
    name                   = aws_cloudfront_distribution.this.domain_name
    zone_id                = aws_cloudfront_distribution.this.hosted_zone_id
    evaluate_target_health = false
  }
}
