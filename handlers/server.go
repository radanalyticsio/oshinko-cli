package handlers

import (
    middleware "github.com/go-openapi/runtime/middleware"
    
    "github.com/redhatanalytics/oshinko-rest/restapi/operations/server"
)

var appName string
var gitTag string

// ServerResponse respond to the server info request
func ServerResponse() middleware.Responder {
    payload := server.GetServerInfoOKBodyBody{
        Application: &server.GetServerInfoOKBodyApplication{Name: &appName, Version: &gitTag}}
    return server.NewGetServerInfoOK().WithPayload(payload)   
}