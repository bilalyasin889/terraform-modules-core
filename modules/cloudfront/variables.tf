##############################
# cloudfront :: Variables
###############################

variable "name" {
  type        = string
  description = "THe name for the CLoudFront distribution"
}

variable "enabled" {
  type        = bool
  description = "Set to true to enable the CLoudFront distribution"
}

variable "domains" {
  type        = list(string)
  description = "The domain names for the cloudfront distribution"
}

variable "certificate_arn" {
  type        = string
  description = "The ARN of the ACM certificate to attach. Ensure custom certificate is used to allow domain_names."
}

variable "hosted_zone_id" {
  type        = string
  default     = null
  description = "Optional Route 53 Hosted Zone ID. If provided, Terraform will create DNS records automatically for CloudFront aliases."
}

variable "default_root_object" {
  type        = string
  default     = "index.html"
  description = "The default root object to serve for the distribution"
}

variable "cloudfront_functions" {
  description = "Map of CloudFront Function configurations. Key is used as default function name if 'name' not specified"
  type = map(object({
    name                         = optional(string)
    runtime                      = optional(string, "cloudfront-js-2.0")
    comment                      = optional(string)
    publish                      = optional(bool)
    code                         = string
    key_value_store_associations = optional(list(string))
  }))
  default = {}
}

variable "origins" {
  type = map(object({
    type           = string
    domain_name    = string
    origin_path    = optional(string, "")
    custom_headers = optional(map(string), {})

    # Required only for S3
    bucket_arn = optional(string)
    bucket_id  = optional(string)

  }))

  description = "Map of CloudFront origins keyed by origin ID."

  # Validate type
  validation {
    condition = alltrue([
      for o in values(var.origins) : contains(
        ["s3", "elb", "api-gateway", "mediastore", "mediapackage", "custom"],
        lower(o.type)
      )
    ])
    error_message = "Each origin type must be one of: s3, elb, api-gateway, mediastore, mediapackage, custom."
  }

  # Validate domain_name is not empty
  validation {
    condition = alltrue([
      for o in values(var.origins) : (length(o.domain_name) > 0)
    ])
    error_message = "Each origin must have a non-empty domain_name."
  }

  # Validate S3 origins have bucket_id and bucket_arn
  validation {
    condition = alltrue([
      for o in values(var.origins) :
      (lower(o.type) != "s3" || (contains(keys(o), "bucket_arn") && contains(keys(o), "bucket_id")))
    ])
    error_message = "S3 origins must define both bucket_id and bucket_arn."
  }
}

variable "default_cache_behavior" {
  type = object({
    origin_id                = string
    allowed_methods          = optional(list(string), ["GET", "HEAD", "OPTIONS"])
    cached_methods           = optional(list(string), ["GET", "HEAD"])
    viewer_protocol_policy   = optional(string, "redirect-to-https")
    cache_policy_id          = optional(string, "658327ea-f89d-4fab-a63d-7e88639e58f6")
    origin_request_policy_id = optional(string)

    functions = optional(list(object({
      event_type = string
      arn        = optional(string)
      function_key = optional(string)
    })), [])
  })

  description = "Default cache behavior for the CloudFront distribution. Must reference a valid origin_id from var.origins."

  # Validate that the origin_id exists
  validation {
    condition     = contains(keys(var.origins), var.default_cache_behavior.origin_id)
    error_message = "default_cache_behavior.origin_id must reference a valid origin key from var.origins."
  }

  # Validation: event_type must be valid
  validation {
    condition = alltrue([
      for fn in var.default_cache_behavior.functions : contains(["viewer-request", "viewer-response"], fn.event_type)
    ])
    error_message = "Each function in default_cache_behavior.functions must have event_type 'viewer-request' or 'viewer-response'."
  }

  # Validation: only one function per event_type
  validation {
    condition     = length(distinct([for fn in var.default_cache_behavior.functions : fn.event_type])) == length(var.default_cache_behavior.functions)
    error_message = "Each event_type can only be used once in default_cache_behavior.functions."
  }

  # Validation: function arn or key must be provider
  validation {
    condition = alltrue([
      for fn in var.default_cache_behavior.functions : (fn.arn != null || fn.function_key != null)
    ])
    error_message = "Each function must have either an 'arn' or a 'function_key' defined."
  }

  # Validation: function_key must reference a defined CloudFront Function
  validation {
    condition = alltrue([
      for fn in var.default_cache_behavior.functions :
      (fn.function_key == null || contains(keys(var.cloudfront_functions), fn.function_key))
      ])
    error_message = "Each function_key in default_cache_behavior.functions must reference a key defined in var.cloudfront_functions."
  }
}

variable "ordered_cache_behaviors" {
  type = list(object({
    path_pattern             = string
    target_origin_id         = string
    allowed_methods          = optional(list(string), ["GET", "HEAD", "OPTIONS"])
    cached_methods           = optional(list(string), ["GET", "HEAD"])
    viewer_protocol_policy   = optional(string, "redirect-to-https")
    cache_policy_id          = optional(string, "658327ea-f89d-4fab-a63d-7e88639e58f6")
    origin_request_policy_id = optional(string)

    functions = optional(list(object({
      event_type = string
      arn        = optional(string)
      function_key = optional(string)
    })), [])
  }))
  default = []

  description = "List of ordered cache behaviors with explicit target_origin_id for each."

  # Validate that each target_origin_id exists
  validation {
    condition = alltrue([
      for cb in var.ordered_cache_behaviors : contains(keys(var.origins), cb.target_origin_id)
    ])
    error_message = "Each ordered_cache_behavior.target_origin_id must reference a valid origin key from var.origins."
  }

  # Validation: each function must have valid event_type
  validation {
    condition = alltrue([
      for beh in var.ordered_cache_behaviors : alltrue([
        for fn in lookup(beh, "functions", []) : contains(["viewer-request", "viewer-response"], fn.event_type)
      ])
    ])
    error_message = "Each function in ordered_cache_behaviors.functions must have event_type 'viewer-request' or 'viewer-response'."
  }

  # Validation: only one function per event_type per behavior
  validation {
    condition = alltrue([
      for beh in var.ordered_cache_behaviors :
      (length(distinct([for fn in lookup(beh, "functions", []) : fn.event_type])) ==
      length(lookup(beh, "functions", [])))
    ])
    error_message = "Each event_type can only be used once per cache behavior in ordered_cache_behaviors.functions."
  }

  # Validation: function arn or key must be provider
  validation {
    condition = alltrue([
      for beh in var.ordered_cache_behaviors : alltrue([
        for fn in lookup(beh, "functions", []) : (fn.arn != null || fn.function_key != null)
      ])
    ])
    error_message = "Each function in ordered_cache_behaviors must have either an 'arn' or a 'function_key' defined."
  }

  # Validation: function_key must reference a defined CloudFront Function
  validation {
    condition = alltrue([
      for beh in var.ordered_cache_behaviors : alltrue([
        for fn in lookup(beh, "functions", []) :
        (fn.function_key == null || contains(keys(var.cloudfront_functions), fn.function_key))
        ])
    ])
    error_message = "Each function_key in ordered_cache_behaviors.functions must reference a key defined in var.cloudfront_functions."
  }
}

variable "custom_error_responses" {
  type = list(object({
    error_caching_min_ttl = number
    error_code            = number
    response_code         = number
    response_page_path    = string
  }))
  default     = []
  description = "List of CloudFront custom error responses to apply."

  # Validate error_code
  validation {
    condition = alltrue([
      for r in var.custom_error_responses :
      contains([400, 403, 404, 405, 414, 416, 500, 501, 502, 503, 504], r.error_code)
    ])
    error_message = "error_code must be one of: 400, 403, 404, 405, 414, 416, 500, 501, 502, 503, 504."
  }

  # Validate response_code
  validation {
    condition = alltrue([
      for r in var.custom_error_responses :
      (r.response_code >= 100 && r.response_code <= 599)
    ])
    error_message = "response_code must be a valid HTTP status code (100-599)."
  }

  # Validate TTL
  validation {
    condition = alltrue([
      for r in var.custom_error_responses :
      (r.error_caching_min_ttl >= 0)
    ])
    error_message = "error_caching_min_ttl must be 0 or greater."
  }

  # Validate response_page_path
  validation {
    condition = alltrue([
      for r in var.custom_error_responses :
      can(regex("^/.+", r.response_page_path))
    ])
    error_message = "response_page_path must start with '/'."
  }
}

variable "price_class" {
  type        = string
  default     = null
  description = "The price class to set"

  validation {
    condition = (var.price_class == null ||
    contains(["PriceClass_All", "PriceClass_200", "PriceClass_100"], var.price_class))
    error_message = "price_class must be one of: PriceClass_All, PriceClass_200, PriceClass_100"
  }
}

variable "restrictions" {
  type = object({
    type      = string
    locations = list(string)
  })

  default     = null
  description = "Geo location restrictions to apply."

  # Validate type value
  validation {
    condition = (var.restrictions == null ||
    contains(["none", "whitelist", "blacklist"], lower(var.restrictions.type)))
    error_message = "restrictions.type must be one of: none, whitelist, blacklist."
  }

  # Validate ISO 3166-1 alpha-2 codes
  validation {
    condition = (var.restrictions == null ||
      alltrue([
        for c in var.restrictions.locations :
        can(regex("^[A-Z]{2}$", upper(c)))
    ]))
    error_message = "restrictions.locations must contain only ISO 3166-1 alpha-2 country codes (e.g., GB, US, FR)."
  }
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Tags to apply to CloudFront resources"
}
