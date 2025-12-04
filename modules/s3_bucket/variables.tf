##############################
# s3_bucket :: Variables
##############################

variable "bucket_name" {
  type        = string
  description = "The name of the S3 bucket"
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Tags to assign to the bucket"
}

variable "versioning_enabled" {
  type        = bool
  default     = false
  description = "Enable S3 bucket versioning"
}

variable "public_read" {
  type = object({
    enabled         = bool
    allowed_domains = list(string)
  })
  default = { enabled = false, allowed_domains = [] }
  description = "Configuration for GET access bucket policy. Controls whether GET access is allowed from specific domains."
}

variable "put_cors" {
  type = object({
    enabled         = bool
    allowed_domains = list(string)
    max_age_seconds = optional(number, 300)
    allowed_headers = optional(list(string), ["*"])
  })
  default = { enabled = false, allowed_domains = [], max_age_seconds = 300, allowed_headers = ["*"] }
  description = "Configuration for PUT CORS rules. Controls browser access for PUT requests from allowed domains."
}

variable "allow_force_destroy" {
  type        = bool
  default     = false
  description = "Force bucket destruction by deleting all objects and versions."
}
