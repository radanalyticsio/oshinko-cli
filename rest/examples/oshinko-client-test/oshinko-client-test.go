package main

import (
	"fmt"

	httptransport "github.com/go-openapi/runtime/client"
	strfmt "github.com/go-openapi/strfmt"

	"github.com/radanalyticsio/oshinko-cli/rest/client"
)

// A simple application to demonstrate the API client for oshinko-rest-server
// this will interact with the API server at 127.0.0.1:8080 and retreive its
// name and version from the server info response.
func main() {
	transport := httptransport.New("127.0.0.1:8080", "/", []string{"http"})
	formats := strfmt.Default
	cli := client.New(transport, formats)

	resp, err := cli.Server.GetServerInfo(nil)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("name:", *resp.Payload.Application.Name)
		fmt.Println("version:", *resp.Payload.Application.Version)
	}
}
