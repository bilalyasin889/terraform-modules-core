##############################
# acm :: Variables
##############################

variable "domain_name" {
  type        = string
  description = "Primary domain name for the certificate"
}

variable "hosted_zone_id" {
  type        = string
  default     = null
  description = "Route53 Hosted Zone ID for DNS validation. Leave empty for manual validation."
}

variable "subject_alternative_names" {
  type        = list(string)
  default     = []
  description = "Additional SANs for the certificate"
}

variable "region" {
  type        = string
  default     = "eu-west-2"
  description = "AWS region to create the ACM certificate in."
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Tags to apply to ACM resources"
}
