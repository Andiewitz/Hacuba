output "listing_images_bucket_name" {
  description = "Private bucket used by Listings for original image objects."
  value       = aws_s3_bucket.listing_images.bucket
}

output "listing_images_cloudfront_domain" {
  description = "Set LISTINGS_IMAGE_BASE_URL to this HTTPS domain in the client deployment."
  value       = "https://${aws_cloudfront_distribution.listing_images.domain_name}"
}

output "listings_image_service_policy_arn" {
  description = "Attach this policy to the Listings workload IAM role."
  value       = aws_iam_policy.listings_image_service.arn
}

output "auth_runtime_secrets_policy_arn" {
  description = "Attach this policy to the Auth workload IAM role."
  value       = aws_iam_policy.auth_runtime_secrets.arn
}

output "listings_runtime_secrets_policy_arn" {
  description = "Attach this policy to the Listings workload IAM role."
  value       = aws_iam_policy.listings_runtime_secrets.arn
}

output "support_runtime_secrets_policy_arn" {
  description = "Attach this policy to the Support workload IAM role."
  value       = aws_iam_policy.support_runtime_secrets.arn
}

output "auth_database_secret_arn" {
  description = "Secret containing the Auth service DATABASE_URL."
  value       = aws_secretsmanager_secret.auth_database.arn
}

output "listings_database_secret_arn" {
  description = "Secret containing the Listings service DATABASE_URL."
  value       = aws_secretsmanager_secret.listings_database.arn
}

output "support_database_secret_arn" {
  description = "Secret containing the Support service DATABASE_URL."
  value       = aws_secretsmanager_secret.support_database.arn
}

output "jwt_secret_arn" {
  description = "Shared JWT_SECRET for Auth and Listings."
  value       = aws_secretsmanager_secret.jwt.arn
}
