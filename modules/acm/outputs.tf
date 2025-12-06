##############################
# acm :: Outputs
##############################

output "certificate_arn" {
  description = "The ARN of the ACM certificate"
  value       = aws_acm_certificate.this.arn
}

output "dns_validation" {
  description = "The DNS records to create for validating the ACM certificate, if manually validating."
  value       = var.hosted_zone_id != null ? null : {
    for dvo in aws_acm_certificate.this.domain_validation_options :
    dvo.domain_name => {
      name  = dvo.resource_record_name
      type  = dvo.resource_record_type
      value = dvo.resource_record_value
    }
  }
}