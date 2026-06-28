terraform {
  required_version = ">= 1.5.0, < 2.0.0"

  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = "= 0.13.0"
    }
  }
}
