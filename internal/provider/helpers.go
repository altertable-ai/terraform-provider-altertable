package provider

import (
	"errors"
	"net/http"

	"github.com/altertable-ai/terraform-provider-altertable/internal/client"
)

// isNotFound reports whether err is an Altertable 404, used to drop deleted resources from state.
func isNotFound(err error) bool {
	var apiErr *client.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}
