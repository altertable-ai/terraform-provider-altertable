# Terraform 1.12+: import by identity (preferred — typed, no delimiter parsing).
import {
  to = altertable_catalog.warehouse
  identity = {
    environment_id = "env_abc123"
    id             = "db_def456"
  }
}
