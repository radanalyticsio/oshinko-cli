package handlers

import (
	middleware "github.com/go-openapi/runtime/middleware"

	"github.com/redhatanalytics/oshinko-rest/restapi/operations/server"
	"github.com/redhatanalytics/oshinko-rest/version"
)

// ServerResponse respond to the server info request
func ServerResponse() middleware.Responder {
	vers := version.GetVersion()
	name := version.GetAppName()
	payload := server.GetServerInfoOKBodyBody{
		Application: &server.GetServerInfoOKBodyApplication{
			Name: &name, Version: &vers}}
	return server.NewGetServerInfoOK().WithPayload(payload)
}
