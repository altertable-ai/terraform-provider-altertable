package provider

import "strings"

// parseEnvScopedImportID splits the back-compat "environment_id:id" import string
// shared by every environment-scoped resource (catalog, bucket).
func parseEnvScopedImportID(importID string) (env, id string, ok bool) {
	parts := strings.SplitN(importID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}
