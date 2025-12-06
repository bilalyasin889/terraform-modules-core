# acm Module

## Description

Provision an ACM certificate with DNS validation via Route53.

---

## Inputs

| Name                        | Description                                                                   | Type           | Default       | Required |
|-----------------------------|-------------------------------------------------------------------------------|----------------|---------------|----------|
| `domain_name`               | Primary domain name for the certificate                                       | `string`       | -             | yes      |
| `hosted_zone_id`            | Route53 Hosted Zone ID for DNS validation. Leave empty for manual validation. | `string`       | `null`        | no       |
| `subject_alternative_names` | Additional SANs for the certificate                                           | `list(string)` | `[]`          | no       |
| `region`                    | AWS region to create the ACM certificate in                                   | `string`       | `"eu-west-2"` | no       |
| `tags`                      | Tags to apply to ACM resources                                                | `map(string)`  | `{}`          | no       |

## Outputs

| Name              | Description                                                                           |
|-------------------|---------------------------------------------------------------------------------------|
| `certificate_arn` | The ARN of the ACM certificate                                                        |
| `dns_validation`  | The DNS records to create for validating the ACM certificate, if manually validating. |

## Example Usage

```hcl
module "acm_cert_public" {
  source = "git::ssh://github.com/bilalyasin889/terraform-modules-core.git//modules/acm?ref=vX.Y.Z"

  domain_name    = "example.com"
  subject_alternative_names = ["www.example.com"]
  region         = "eu-west-2"
  hosted_zone_id = "Z123456ABCDEFG"
  tags = {
    Environment = "prod"
    ManagedBy   = "Terraform"
  }
}
```
