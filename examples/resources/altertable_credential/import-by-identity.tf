# Terraform 1.12+: import by identity (preferred — typed, no delimiter parsing).
import {
  to = altertable_credential.ci
  identity = {
    principal_type = "service_account"
    principal_id   = "sa_abc"
    environment_id = "env_def"
    id             = "cred_ghi"
  }
}
