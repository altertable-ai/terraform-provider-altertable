variable "altertable_management_api_key" {
  description = "Altertable management API key."
  type        = string
  sensitive   = true
}

variable "pg_password" {
  description = "Password for the external Postgres connection catalog."
  type        = string
  sensitive   = true
}
