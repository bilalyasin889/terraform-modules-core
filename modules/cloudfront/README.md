# cloudfront Module

## Description

Creates a CloudFront distribution with configurable origins, cache behaviors, CloudFront Functions, SSL support, geo
restrictions, and optional Route 53 DNS records for custom domains.

---

## Inputs

| Name                    | Description                                                   | Type         | Default      | Required |
|-------------------------|---------------------------------------------------------------|--------------|--------------|----------|
| name                    | Name for the CloudFront distribution                          | string       | -            | yes      |
| enabled                 | Enable or disable the CloudFront distribution                 | bool         | -            | yes      |
| domains                 | Custom domain names (aliases) for the distribution            | list(string) | -            | yes      |
| certificate_arn         | ACM certificate ARN for HTTPS and custom domains              | string       | -            | yes      |
| hosted_zone_id          | Route 53 Hosted Zone ID to automatically create alias records | string       | null         | no       |
| default_root_object     | Default object served when accessing the root URL             | string       | "index.html" | no       |
| cloudfront_functions    | Map of CloudFront Function definitions                        | map(object)  | {}           | no       |
| origins                 | Map of CloudFront origins keyed by origin ID                  | map(object)  | -            | yes      |
| default_cache_behavior  | Default cache behavior configuration                          | object       | -            | yes      |
| ordered_cache_behaviors | Ordered list of additional cache behaviors                    | list(object) | []           | no       |
| custom_error_responses  | Custom error response configurations                          | list(object) | []           | no       |
| price_class             | CloudFront price class (PriceClass_All, 200, or 100)          | string       | null         | no       |
| restrictions            | Geo restriction configuration                                 | object       | null         | no       |
| tags                    | Tags applied to CloudFront resources                          | map(string)  | {}           | no       |

---

## Outputs

| Name                                   | Description                                   |
|----------------------------------------|-----------------------------------------------|
| cloudfront_distribution_id             | CloudFront distribution ID                    |
| cloudfront_distribution_arn            | CloudFront distribution ARN                   |
| cloudfront_distribution_domain_name    | CloudFront distribution domain name           |
| cloudfront_distribution_hosted_zone_id | Route 53 hosted zone ID for CloudFront        |
| cloudfront_distribution_status         | Current deployment status of the distribution |

---

## Example Usage

```hcl
module "cloudfront" {
  source = "git::ssh://github.com/bilalyasin889/terraform-modules-core.git//modules/cloudfront?ref=vX.Y.Z"

  name            = "example-cdn"
  enabled         = true
  domains         = ["cdn.example.com"]
  certificate_arn = aws_acm_certificate.this.arn

  origins = {
    app = {
      type        = "custom"
      domain_name = "app.example.com"
    }
  }

  default_cache_behavior = {
    origin_id = "app"
  }

  tags = {
    Environment = "prod"
  }
}
```

---