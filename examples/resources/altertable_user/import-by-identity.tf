# altertable_user is imported with an identity block (Terraform 1.12+); the legacy string
# form is not supported. Import an accepted member by their raw user UUID — the provider
# resolves it to the stable Terraform id:
import {
  to = altertable_user.teammate
  identity = {
    user_id = "7a6d2c5b-9e0f-4e7f-8a1d-3f9c2d186b4a"
  }
}
