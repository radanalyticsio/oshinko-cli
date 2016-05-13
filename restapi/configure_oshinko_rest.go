package restapi

import (
	"crypto/tls"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"

	"github.com/redhatanalytics/oshinko-rest/handlers"
	"github.com/redhatanalytics/oshinko-rest/restapi/operations"
	"github.com/redhatanalytics/oshinko-rest/restapi/operations/clusters"
	"github.com/redhatanalytics/oshinko-rest/restapi/operations/server"
)

// This file is safe to edit. Once it exists it will not be overwritten

func configureFlags(api *operations.OshinkoRestAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.OshinkoRestAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.ClustersCreateClusterHandler = clusters.CreateClusterHandlerFunc(func(params clusters.CreateClusterParams) middleware.Responder {
		return middleware.NotImplemented("operation clusters.CreateCluster has not yet been implemented")
	})
	api.ClustersDeleteSingleClusterHandler = clusters.DeleteSingleClusterHandlerFunc(func(params clusters.DeleteSingleClusterParams) middleware.Responder {
		return middleware.NotImplemented("operation clusters.DeleteSingleCluster has not yet been implemented")
	})
	api.ClustersFindClustersHandler = clusters.FindClustersHandlerFunc(func() middleware.Responder {
		return middleware.NotImplemented("operation clusters.FindClusters has not yet been implemented")
	})
	api.ClustersFindSingleClusterHandler = clusters.FindSingleClusterHandlerFunc(func(params clusters.FindSingleClusterParams) middleware.Responder {
		return middleware.NotImplemented("operation clusters.FindSingleCluster has not yet been implemented")
	})
	api.ServerGetServerInfoHandler = server.GetServerInfoHandlerFunc(func() middleware.Responder {
		return middleware.NotImplemented("operation server.GetServerInfo has not yet been implemented")
	})
	api.ClustersUpdateSingleClusterHandler = clusters.UpdateSingleClusterHandlerFunc(func(params clusters.UpdateSingleClusterParams) middleware.Responder {
		return middleware.NotImplemented("operation clusters.UpdateSingleCluster has not yet been implemented")
	})

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
