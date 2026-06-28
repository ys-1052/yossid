variable "neon_project_name" {
  description = "Neon project name"
  type        = string
  default     = "yossid"
}

variable "neon_region_id" {
  description = "Neon region ID (e.g. aws-ap-southeast-1)"
  type        = string
  default     = "aws-ap-southeast-1"
}

variable "neon_database_name" {
  description = "Database name"
  type        = string
  default     = "yossid"
}

variable "app_user_name" {
  description = "Application DB role (runtime use only)"
  type        = string
  default     = "app_user"
}

variable "migration_user_name" {
  description = "Migration DB role (DDL, not passed to Lambda)"
  type        = string
  default     = "migration_user"
}

variable "neon_org_id" {
  description = "Neon Organization ID"
  type        = string
}
