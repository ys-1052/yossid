# Neon Postgres Provisioning (Terraform)

This directory manages the Neon serverless PostgreSQL database project, branch, and database roles using Terraform.

## Prerequisites

1. Install Terraform (`v1.5.7` or compatible).
2. Set your Neon API Key in your shell environment:
   ```bash
   export NEON_API_KEY="your-neon-api-key"
   ```

## Usage

```bash
# 1. Initialize
terraform init

# 2. Plan configuration changes
terraform plan

# 3. Apply changes to Neon
terraform apply
```

## Outputs

After a successful apply, Terraform outputs two sensitive connection strings:
- **`migration_database_url`**: Direct connection string with DDL rights. Used for running database migrations locally.
- **`app_database_url`**: Pooled connection string with limited CRUD rights. Used for the Lambda application runtime.
