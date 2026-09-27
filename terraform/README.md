# Hacuba production infrastructure

This module creates the AWS data and image layer. It deliberately does not
create a public database endpoint or grant the browser AWS credentials.

- Two independent encrypted RDS PostgreSQL instances (`auth` and `listings`)
  sit in existing private subnets. Their security group accepts port 5432 only
  from the supplied application security groups.
- Each database DSN and the shared JWT secret are stored in Secrets Manager.
- Listing originals live in a private S3 bucket. The Listings workload receives
  a narrowly scoped policy for `listings/*`; browsers upload only through
  short-lived presigned PUT URLs.
- CloudFront is the only S3 reader. Use its HTTPS output as the client
  `LISTINGS_IMAGE_BASE_URL`.
- Direct uploads begin with `state=unregistered`. The lifecycle rule removes
  objects that were never registered with the Listings API after one day.

## Apply

Install Terraform 1.6+ and configure AWS credentials for the target account.
The VPC, at least two private subnets in separate Availability Zones, and the
security group used by the application workload must exist before this module
is applied. Terraform state contains generated database credentials, so create
an encrypted remote state bucket and lock table through the platform bootstrap
process, then copy and fill `backend.hcl.example` before the first apply.

```powershell
Copy-Item terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with the real VPC, subnets, workload security group,
# and deployed HTTPS client origin.
Copy-Item backend.hcl.example backend.hcl
# Edit backend.hcl with the remote state bucket and lock table.
terraform init -backend-config=backend.hcl
terraform plan -out hacuba.tfplan
terraform apply hacuba.tfplan
```

Attach `listings_image_service_policy_arn` and
`listings_runtime_secrets_policy_arn` to the IAM role used by the Listings
service. Attach `auth_runtime_secrets_policy_arn` to the Auth workload role.
Inject the JSON values as `DATABASE_URL` and `JWT_SECRET`. Listings also needs
`S3_REGION` and `S3_BUCKET` from the Terraform outputs. Set the client
deployment's `LISTINGS_IMAGE_BASE_URL` to `listing_images_cloudfront_domain`.

The Support workload uses its own private RDS instance. Attach
`support_runtime_secrets_policy_arn` to that workload and inject both its
database secret and the shared `jwt_secret_arn`; then apply
`services/support/migrations/001_init.sql` before accepting reports.

Run the SQL migrations from a private CI runner or an approved bastion that can
reach the RDS security group, before starting either service:

```powershell
Get-ChildItem ..\services\auth\migrations\*.sql | ForEach-Object { psql $env:AUTH_DATABASE_URL -v ON_ERROR_STOP=1 -f $_ }
Get-ChildItem ..\services\listings\migrations\*.sql | ForEach-Object { psql $env:LISTINGS_DATABASE_URL -v ON_ERROR_STOP=1 -f $_ }
```

The `sslmode=require` DSNs enforce encrypted database connections. Configure
the application runtime with the current Amazon RDS CA bundle and use
`sslmode=verify-full` if its PostgreSQL client setup supports CA verification.
