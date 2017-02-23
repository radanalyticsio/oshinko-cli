// Package errors contains primitives for working with oshinko errors easier.
package errors

import (
    "fmt"

	"github.com/radanalyticsio/oshinko-cli/rest/models"
)

// NewSingleErrorResponse creates a new error reponse object for use as a return
// value from a REST request.
func NewSingleErrorResponse(status int32, title string, details string) *models.ErrorResponse {
	error := models.ErrorModel{Status: &status, Details: &details, Title: &title}
	response := models.ErrorResponse{Errors: []*models.ErrorModel{&error}}
	return &response
}

// SingleErrorToString takes a pointer to an models.ErrorModel and returns a
// string with the expanded information.
func SingleErrorToString(err *models.ErrorModel) string {
    return fmt.Sprintf(
        "Title:\t\t%s\nStatus:\t\t%d\nDetails:\t%s",
        *err.Title, *err.Status, *err.Details)
}
