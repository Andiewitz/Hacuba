locals {
  name              = "${var.project_name}-${var.environment}"
  image_bucket_name = "${local.name}-listings-${data.aws_caller_identity.current.account_id}"
}

resource "aws_s3_bucket" "listing_images" {
  bucket = local.image_bucket_name
}

resource "aws_s3_bucket_public_access_block" "listing_images" {
  bucket                  = aws_s3_bucket.listing_images.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "listing_images" {
  bucket = aws_s3_bucket.listing_images.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_cors_configuration" "listing_images" {
  bucket = aws_s3_bucket.listing_images.id

  cors_rule {
    allowed_headers = ["content-type", "x-amz-tagging"]
    allowed_methods = ["PUT"]
    allowed_origins = var.client_origins
    expose_headers  = ["ETag"]
    max_age_seconds = 300
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "listing_images" {
  bucket = aws_s3_bucket.listing_images.id

  rule {
    id     = "expire-unregistered-uploads"
    status = "Enabled"

    filter {
      tag {
        key   = "state"
        value = "unregistered"
      }
    }

    expiration {
      days = var.unregistered_upload_expiry_days
    }
  }

  rule {
    id     = "abort-incomplete-uploads"
    status = "Enabled"

    filter {
      prefix = ""
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }
}

resource "aws_cloudfront_origin_access_control" "listing_images" {
  name                              = "${local.name}-listing-images"
  description                       = "CloudFront-only read access to private Hacuba listing images"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

resource "aws_cloudfront_distribution" "listing_images" {
  enabled         = true
  is_ipv6_enabled = true
  price_class     = var.cloudfront_price_class

  origin {
    domain_name              = aws_s3_bucket.listing_images.bucket_regional_domain_name
    origin_id                = "listing-images-s3"
    origin_access_control_id = aws_cloudfront_origin_access_control.listing_images.id
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "listing-images-s3"
    viewer_protocol_policy = "redirect-to-https"
    compress               = true
    min_ttl                = 0
    default_ttl            = 86400
    max_ttl                = 31536000

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}

data "aws_iam_policy_document" "listing_images_bucket" {
  statement {
    sid       = "AllowCloudFrontReadOnly"
    effect    = "Allow"
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.listing_images.arn}/listings/*"]

    principals {
      type        = "Service"
      identifiers = ["cloudfront.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "AWS:SourceArn"
      values   = [aws_cloudfront_distribution.listing_images.arn]
    }
  }
}

resource "aws_s3_bucket_policy" "listing_images" {
  bucket = aws_s3_bucket.listing_images.id
  policy = data.aws_iam_policy_document.listing_images_bucket.json
}

data "aws_iam_policy_document" "listings_image_service" {
  statement {
    sid = "ManageListingImages"
    actions = [
      "s3:GetObject",
      "s3:GetObjectTagging",
      "s3:PutObject",
      "s3:PutObjectTagging",
    ]
    resources = ["${aws_s3_bucket.listing_images.arn}/listings/*"]
  }
}

resource "aws_iam_policy" "listings_image_service" {
  name        = "${local.name}-listings-image-service"
  description = "Minimal S3 access needed by the Listings workload to issue and register image uploads"
  policy      = data.aws_iam_policy_document.listings_image_service.json
}

data "aws_iam_policy_document" "auth_runtime_secrets" {
  statement {
    sid       = "ReadAuthRuntimeSecrets"
    actions   = ["secretsmanager:GetSecretValue"]
    resources = [aws_secretsmanager_secret.auth_database.arn, aws_secretsmanager_secret.jwt.arn]
  }
}

data "aws_iam_policy_document" "listings_runtime_secrets" {
  statement {
    sid       = "ReadListingsRuntimeSecrets"
    actions   = ["secretsmanager:GetSecretValue"]
    resources = [aws_secretsmanager_secret.listings_database.arn, aws_secretsmanager_secret.jwt.arn]
  }
}

data "aws_iam_policy_document" "support_runtime_secrets" {
  statement {
    sid       = "ReadSupportRuntimeSecrets"
    actions   = ["secretsmanager:GetSecretValue"]
    resources = [aws_secretsmanager_secret.support_database.arn, aws_secretsmanager_secret.jwt.arn]
  }
}

resource "aws_iam_policy" "auth_runtime_secrets" {
  name        = "${local.name}-auth-runtime-secrets"
  description = "Read only the Auth database DSN and shared JWT secret"
  policy      = data.aws_iam_policy_document.auth_runtime_secrets.json
}

resource "aws_iam_policy" "listings_runtime_secrets" {
  name        = "${local.name}-listings-runtime-secrets"
  description = "Read only the Listings database DSN and shared JWT secret"
  policy      = data.aws_iam_policy_document.listings_runtime_secrets.json
}

resource "aws_iam_policy" "support_runtime_secrets" {
  name        = "${local.name}-support-runtime-secrets"
  description = "Read only the Support database DSN and shared JWT secret"
  policy      = data.aws_iam_policy_document.support_runtime_secrets.json
}

resource "aws_db_subnet_group" "service_databases" {
  name       = "${local.name}-service-databases"
  subnet_ids = var.private_subnet_ids
}

resource "aws_security_group" "service_databases" {
  name        = "${local.name}-service-databases"
  description = "Only Hacuba service workloads may connect to Postgres"
  vpc_id      = var.vpc_id

  dynamic "ingress" {
    for_each = toset(var.app_security_group_ids)

    content {
      description     = "Postgres from a Hacuba application workload"
      from_port       = 5432
      to_port         = 5432
      protocol        = "tcp"
      security_groups = [ingress.value]
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "random_password" "auth_database" {
  length           = 32
  special          = true
  override_special = "!#$%&*+-=?^_"
}

resource "random_password" "listings_database" {
  length           = 32
  special          = true
  override_special = "!#$%&*+-=?^_"
}

resource "random_password" "support_database" {
  length           = 32
  special          = true
  override_special = "!#$%&*+-=?^_"
}

resource "random_password" "jwt_secret" {
  length  = 64
  special = false
}

resource "aws_db_instance" "auth" {
  identifier                 = "${local.name}-auth"
  engine                     = "postgres"
  instance_class             = var.rds_instance_class
  allocated_storage          = var.rds_allocated_storage
  max_allocated_storage      = var.rds_allocated_storage * 3
  storage_type               = "gp3"
  storage_encrypted          = true
  db_name                    = "auth"
  username                   = "hacuba_auth"
  password                   = random_password.auth_database.result
  port                       = 5432
  db_subnet_group_name       = aws_db_subnet_group.service_databases.name
  vpc_security_group_ids     = [aws_security_group.service_databases.id]
  publicly_accessible        = false
  multi_az                   = var.rds_multi_az
  backup_retention_period    = var.rds_backup_retention_days
  deletion_protection        = var.rds_deletion_protection
  skip_final_snapshot        = var.rds_skip_final_snapshot
  final_snapshot_identifier  = var.rds_skip_final_snapshot ? null : "${local.name}-auth-final"
  auto_minor_version_upgrade = true
  copy_tags_to_snapshot      = true
  apply_immediately          = false
}

resource "aws_db_instance" "listings" {
  identifier                 = "${local.name}-listings"
  engine                     = "postgres"
  instance_class             = var.rds_instance_class
  allocated_storage          = var.rds_allocated_storage
  max_allocated_storage      = var.rds_allocated_storage * 3
  storage_type               = "gp3"
  storage_encrypted          = true
  db_name                    = "listings"
  username                   = "hacuba_listings"
  password                   = random_password.listings_database.result
  port                       = 5432
  db_subnet_group_name       = aws_db_subnet_group.service_databases.name
  vpc_security_group_ids     = [aws_security_group.service_databases.id]
  publicly_accessible        = false
  multi_az                   = var.rds_multi_az
  backup_retention_period    = var.rds_backup_retention_days
  deletion_protection        = var.rds_deletion_protection
  skip_final_snapshot        = var.rds_skip_final_snapshot
  final_snapshot_identifier  = var.rds_skip_final_snapshot ? null : "${local.name}-listings-final"
  auto_minor_version_upgrade = true
  copy_tags_to_snapshot      = true
  apply_immediately          = false
}

resource "aws_db_instance" "support" {
  identifier                 = "${local.name}-support"
  engine                     = "postgres"
  instance_class             = var.rds_instance_class
  allocated_storage          = var.rds_allocated_storage
  max_allocated_storage      = var.rds_allocated_storage * 3
  storage_type               = "gp3"
  storage_encrypted          = true
  db_name                    = "support"
  username                   = "hacuba_support"
  password                   = random_password.support_database.result
  port                       = 5432
  db_subnet_group_name       = aws_db_subnet_group.service_databases.name
  vpc_security_group_ids     = [aws_security_group.service_databases.id]
  publicly_accessible        = false
  multi_az                   = var.rds_multi_az
  backup_retention_period    = var.rds_backup_retention_days
  deletion_protection        = var.rds_deletion_protection
  skip_final_snapshot        = var.rds_skip_final_snapshot
  final_snapshot_identifier  = var.rds_skip_final_snapshot ? null : "${local.name}-support-final"
  auto_minor_version_upgrade = true
  copy_tags_to_snapshot      = true
  apply_immediately          = false
}

resource "aws_secretsmanager_secret" "auth_database" {
  name                    = "${local.name}/auth/database"
  recovery_window_in_days = 7
}

resource "aws_secretsmanager_secret_version" "auth_database" {
  secret_id = aws_secretsmanager_secret.auth_database.id
  secret_string = jsonencode({
    DATABASE_URL = "postgresql://${aws_db_instance.auth.username}:${urlencode(random_password.auth_database.result)}@${aws_db_instance.auth.address}:${aws_db_instance.auth.port}/${aws_db_instance.auth.db_name}?sslmode=require"
  })
}

resource "aws_secretsmanager_secret" "listings_database" {
  name                    = "${local.name}/listings/database"
  recovery_window_in_days = 7
}

resource "aws_secretsmanager_secret_version" "listings_database" {
  secret_id = aws_secretsmanager_secret.listings_database.id
  secret_string = jsonencode({
    DATABASE_URL = "postgresql://${aws_db_instance.listings.username}:${urlencode(random_password.listings_database.result)}@${aws_db_instance.listings.address}:${aws_db_instance.listings.port}/${aws_db_instance.listings.db_name}?sslmode=require"
  })
}

resource "aws_secretsmanager_secret" "support_database" {
  name                    = "${local.name}/support/database"
  recovery_window_in_days = 7
}

resource "aws_secretsmanager_secret_version" "support_database" {
  secret_id = aws_secretsmanager_secret.support_database.id
  secret_string = jsonencode({
    DATABASE_URL = "postgresql://${aws_db_instance.support.username}:${urlencode(random_password.support_database.result)}@${aws_db_instance.support.address}:${aws_db_instance.support.port}/${aws_db_instance.support.db_name}?sslmode=require"
  })
}

resource "aws_secretsmanager_secret" "jwt" {
  name                    = "${local.name}/services/jwt"
  recovery_window_in_days = 7
}

resource "aws_secretsmanager_secret_version" "jwt" {
  secret_id     = aws_secretsmanager_secret.jwt.id
  secret_string = jsonencode({ JWT_SECRET = random_password.jwt_secret.result })
}
