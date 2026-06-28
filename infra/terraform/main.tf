# Neon provider uses NEON_API_KEY environment variable automatically
provider "neon" {}

# ────────────────────────────────────────────
# Project
# ────────────────────────────────────────────
resource "neon_project" "yossid" {
  name                      = var.neon_project_name
  region_id                 = var.neon_region_id
  org_id                    = var.neon_org_id
  history_retention_seconds = 21600
}

# ────────────────────────────────────────────
# Roles
# ────────────────────────────────────────────

# app_user — runtime only (SELECT / INSERT / UPDATE / DELETE)
resource "neon_role" "app_user" {
  project_id = neon_project.yossid.id
  branch_id  = neon_project.yossid.default_branch_id
  name       = var.app_user_name
}

# migration_user — DDL execution (never passed to Lambda)
resource "neon_role" "migration_user" {
  project_id = neon_project.yossid.id
  branch_id  = neon_project.yossid.default_branch_id
  name       = var.migration_user_name
}

# ────────────────────────────────────────────
# Database
# ────────────────────────────────────────────
resource "neon_database" "yossid" {
  project_id = neon_project.yossid.id
  branch_id  = neon_project.yossid.default_branch_id
  name       = var.neon_database_name
  owner_name = var.migration_user_name

  depends_on = [neon_role.migration_user]
}
