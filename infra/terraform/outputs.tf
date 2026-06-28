output "project_id" {
  description = "Neon project ID"
  value       = neon_project.yossid.id
}

output "branch_id" {
  description = "Neon branch ID"
  value       = neon_project.yossid.default_branch_id
}

output "database_name" {
  description = "Database name"
  value       = neon_database.yossid.name
}

# ── Connection strings ───────────────────────────────────────────────────────
# app_user: pooled connection for Lambda runtime (sslmode=verify-full)
output "app_database_url" {
  description = "Pooled connection string for app_user (store in SSM SecureString)"
  value       = "postgresql://${neon_role.app_user.name}:${neon_role.app_user.password}@${neon_project.yossid.connection_uri_pooler}/${var.neon_database_name}?sslmode=verify-full"
  sensitive   = true
}

# migration_user: direct connection for running migrations
output "migration_database_url" {
  description = "Direct connection string for migration_user (DDL only, do NOT pass to Lambda)"
  value       = "postgresql://${neon_role.migration_user.name}:${neon_role.migration_user.password}@${neon_project.yossid.connection_uri}/${var.neon_database_name}?sslmode=verify-full"
  sensitive   = true
}
