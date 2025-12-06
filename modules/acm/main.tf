##############################
# Module: acm
# Description: Provision an ACM certificate with DNS validation via Route53.
##############################

locals {
  manual_validation = var.hosted_zone_id == null
}

# -----------------------------
# ACM Certificate
# -----------------------------
resource "aws_acm_certificate" "this" {
  region                    = var.region
  domain_name               = var.domain_name
  subject_alternative_names = var.subject_alternative_names
  validation_method         = "DNS"
  tags                      = var.tags
}

# -----------------------------
# Route53 DNS validation records
# -----------------------------
resource "aws_route53_record" "validation" {
  for_each = {
    for dvo in aws_acm_certificate.this.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
    if local.manual_validation == false
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 300
  type            = each.value.type
  zone_id         = var.hosted_zone_id
}

# -----------------------------
# Validate ACM certificate
# -----------------------------
resource "aws_acm_certificate_validation" "this" {
  count                   = local.manual_validation == false ? 1 : 0

  region                  = var.region
  certificate_arn         = aws_acm_certificate.this.arn
  validation_record_fqdns = [for record in aws_route53_record.validation : record.fqdn]
}

