// Package errors contains primitives for working with oshinko errors easier.
package errors

import (
	"github.com/redhatanalytics/oshinko-rest/models"
)

// NewSingleErrorResponse creates a new error reponse object for use as a return
// value from a REST request.
func NewSingleErrorResponse(status int32, title string, details string) *models.ErrorResponse {
	error := models.ErrorModel{Status: &status, Details: &details, Title: &title}
	response := models.ErrorResponse{Errors: []*models.ErrorModel{&error}}
	return &response
}
