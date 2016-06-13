package handlers

import (
	"github.com/redhatanalytics/oshinko-rest/models"
)

func makeSingleErrorResponse(status int32, title string, details string) *models.ErrorResponse {
	error := models.ErrorModel{Status: &status, Details: &details, Title: &title}
	response := models.ErrorResponse{Errors: []*models.ErrorModel{&error}}
	return &response
}
