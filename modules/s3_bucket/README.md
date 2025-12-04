# s3_bucket Module

## Description

Creates a securely configured Amazon S3 bucket with optional features such as versioning, public access controls, CORS,
and force destroy.

## Inputs

| Name                  | Description                                                                                               | Type                                                                                                                                                   | Default                                                                                     | Required |
|-----------------------|-----------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------|----------|
| `bucket_name`         | The name of the S3 bucket                                                                                 | `string`                                                                                                                                               | -                                                                                           | yes      |
| `tags`                | Tags to assign to the bucket                                                                              | `map(string)`                                                                                                                                          | `{}`                                                                                        | no       |
| `versioning_enabled`  | Enable S3 bucket versioning                                                                               | `bool`                                                                                                                                                 | `false`                                                                                     | no       |
| `public_read`         | Configuration for GET access bucket policy. Controls whether GET access is allowed from specific domains. | `object({ enabled = bool, allowed_domains = list(string) })`                                                                                           | `{ enabled = false, allowed_domains = [] }`                                                 | no       |
| `put_cors`            | Configuration for PUT CORS rules. Controls browser access for PUT requests from allowed domains.          | `object({ enabled = bool, allowed_domains = list(string), max_age_seconds = optional(number, 300), allowed_headers = optional(list(string), ["*"]) })` | `{ enabled = false, allowed_domains = [], max_age_seconds = 300, allowed_headers = ["*"] }` | no       |
| `allow_force_destroy` | Upon bucket destruction, delete all objects and versions to avoid errors.                                 | `bool`                                                                                                                                                 | `false`                                                                                     | no       |

---

## Outputs

| Name                 | Description                           |
|----------------------|---------------------------------------|
| `bucket_name`        | Name of the S3 bucket                 |
| `bucket_id`          | ID of the S3 bucket                   |
| `bucket_arn`         | ARN of the S3 bucket                  |
| `bucket_domain_name` | Regional domain name of the S3 bucket |

---

## Example Usage

## Public Read & PUT CORS Enabled
```hcl
module "s3_bucket" {
  source = "git::ssh://github.com/bilalyasin889/terraform-modules-core.git//modules/s3_bucket?ref=vX.Y.Z"

  bucket_name = "terraform-terratest-bucket-1"
  tags        = { CreatedBy = "terraform" }

  public_read = {
    enabled         = true
    allowed_domains = ["https://example.com"]
  }

  put_cors = {
    enabled         = true
    allowed_domains = ["https://example.com"]
  }
}
```

## Versioning & Force Destroy Enabled
```hcl
module "s3_bucket" {
  source = "git::ssh://github.com/bilalyasin889/terraform-modules-core.git//modules/s3_bucket?ref=vX.Y.Z"

  bucket_name         = "terraform-terratest-bucket-2"
  tags                = { CreatedBy = "terraform" }
  versioning_enabled  = true
  allow_force_destroy = true
}
```

---