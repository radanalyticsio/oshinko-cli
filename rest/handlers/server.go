package handlers

import (
	middleware "github.com/go-openapi/runtime/middleware"
	routeclient "github.com/openshift/client-go/route/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	osa "github.com/radanalyticsio/oshinko-cli/rest/helpers/authentication"
	"github.com/radanalyticsio/oshinko-cli/rest/helpers/info"
	"github.com/radanalyticsio/oshinko-cli/rest/restapi/operations/server"
	"github.com/radanalyticsio/oshinko-cli/rest/version"
)

// ServerResponse respond to the server info request
func ServerResponse(params server.GetServerInfoParams) middleware.Responder {
	vers := version.GetVersion()
	name := version.GetAppName()
	webname := info.GetWebServiceName()
	weburl := GetWebServiceURL()
	clusterimage := info.GetSparkImage()
	payload := server.GetServerInfoOKBodyBody{
		Application: &server.GetServerInfoOKBodyApplication{
			Name: &name, Version: &vers,
			WebServiceName: &webname, WebURL: &weburl,
			DefaultClusterImage: &clusterimage}}
	return server.NewGetServerInfoOK().WithPayload(payload)
}

// Look up routes for current namespace and find the one used by oshinko-web
// Will return empty string if no route can be found
func GetWebServiceURL() string {
	weburl := ""
	restConfig, err := osa.GetConfig()
	if err != nil {
		return ""
	}
	namespace, _ := info.GetNamespace()
	routecl, _ := routeclient.NewForConfig(restConfig)
	route, err := routecl.RouteV1().Routes(namespace).Get(info.GetWebServiceName(), metav1.GetOptions{})

	if err != nil || len(route.Status.Ingress) == 0 {
		return ""
	}
	weburl = route.Status.Ingress[0].Host
	return weburl
}
