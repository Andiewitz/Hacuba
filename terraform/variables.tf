variable "aws_region" {
  description = "AWS region for the Hacuba environment."
  type        = string
  default     = "ap-southeast-1"
}

variable "project_name" {
  description = "Short, lowercase name used in AWS resource names."
  type        = string
  default     = "hacuba"
}

variable "environment" {
  description = "Deployment environment name, for example staging or production."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9-]{2,20}$", var.environment))
    error_message = "environment must be 2-20 lowercase letters, digits, or hyphens."
  }
}

variable "vpc_id" {
  description = "Existing VPC ID that contains the private application and database subnets."
  type        = string
}

variable "private_subnet_ids" {
  description = "At least two private subnet IDs in different Availability Zones for RDS."
  type        = list(string)

  validation {
    condition     = length(var.private_subnet_ids) >= 2
    error_message = "private_subnet_ids must contain at least two private subnets."
  }
}

variable "app_security_group_ids" {
  description = "Security groups assigned to the Auth and Listings workloads; only these may reach RDS."
  type        = list(string)

  validation {
    condition     = length(var.app_security_group_ids) > 0
    error_message = "app_security_group_ids must contain at least one workload security group."
  }
}

variable "client_origins" {
  description = "Exact HTTPS origins allowed to PUT directly to S3 through presigned URLs."
  type        = list(string)

  validation {
    condition     = length(var.client_origins) > 0
    error_message = "client_origins must contain the deployed client origin or origins."
  }
}

variable "rds_instance_class" {
  description = "RDS instance size for both service databases."
  type        = string
  default     = "db.t4g.micro"
}

variable "rds_allocated_storage" {
  description = "Initial gp3 storage in GiB for each database."
  type        = number
  default     = 20
}

variable "rds_multi_az" {
  description = "Enable Multi-AZ failover for both databases."
  type        = bool
  default     = true
}

variable "rds_deletion_protection" {
  description = "Prevent accidental deletion of the production databases."
  type        = bool
  default     = true
}

variable "rds_skip_final_snapshot" {
  description = "Allow deletion without a final snapshot. Keep false for every persistent environment."
  type        = bool
  default     = false
}

variable "rds_backup_retention_days" {
  description = "Number of automated backup days retained by RDS."
  type        = number
  default     = 7
}

variable "unregistered_upload_expiry_days" {
  description = "Days before S3 removes a direct upload that was never registered as a listing image."
  type        = number
  default     = 1
}

variable "cloudfront_price_class" {
  description = "CloudFront edge-location price class."
  type        = string
  default     = "PriceClass_200"
}
