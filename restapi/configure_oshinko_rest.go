package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/rs/cors"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/swag"

	"github.com/radanalyticsio/oshinko-rest/handlers"
	oe "github.com/radanalyticsio/oshinko-rest/helpers/errors"
	"github.com/radanalyticsio/oshinko-rest/helpers/logging"
	"github.com/radanalyticsio/oshinko-rest/restapi/operations"
	"github.com/radanalyticsio/oshinko-rest/restapi/operations/clusters"
	"github.com/radanalyticsio/oshinko-rest/restapi/operations/server"
)

// This file is safe to edit. Once it exists it will not be overwritten

type oshinkoOptions struct {
	LogFile string `long:"log-file" description:"the file to write logs into, defaults to stdout"`
}

func configureFlags(api *operations.OshinkoRestAPI) {
	api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{
		{
			ShortDescription: "Oshinko REST server options",
			Options:          &oshinkoOptions{},
		},
	}
}

func configureAPI(api *operations.OshinkoRestAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.ClustersCreateClusterHandler = clusters.CreateClusterHandlerFunc(handlers.CreateClusterResponse)
	api.ClustersDeleteSingleClusterHandler = clusters.DeleteSingleClusterHandlerFunc(handlers.DeleteClusterResponse)
	api.ClustersFindClustersHandler = clusters.FindClustersHandlerFunc(handlers.FindClustersResponse)
	api.ClustersFindSingleClusterHandler = clusters.FindSingleClusterHandlerFunc(handlers.FindSingleClusterResponse)
	api.ServerGetServerInfoHandler = server.GetServerInfoHandlerFunc(handlers.ServerResponse)
	api.ClustersUpdateSingleClusterHandler = clusters.UpdateSingleClusterHandlerFunc(handlers.UpdateSingleClusterResponse)

	api.ServerShutdown = func() {}

	for _, optsGroup := range api.CommandLineOptionsGroups {
		opts, ok := optsGroup.Options.(*oshinkoOptions)
		if ok == true {
			if opts.LogFile != "" {
				err := logging.SetLoggerFile(optsGroup.Options.(*oshinkoOptions).LogFile)
				if err != nil {
					logging.GetLogger().Println("unable to set log file;", err)
				}
			}
		}
	}

	api.Logger = logging.GetLogger().Printf

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
func setupGlobalMiddleware(handler http.Handler) (finalHandler http.Handler) {
	finalHandler = handler
	finalHandler = oe.AddErrorHandler(finalHandler)
	finalHandler = logging.AddLoggingHandler(finalHandler)
	corsHeaders := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "HEAD", "POST", "DELETE", "PUT", "OPTIONS"},
	})
	finalHandler = corsHeaders.Handler(finalHandler)
	return finalHandler
}
