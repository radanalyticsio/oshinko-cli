package restapi

import (
	"crypto/tls"
	"flag"
	"net/http"

	"github.com/rs/cors"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"

	"github.com/radanalyticsio/oshinko-cli/rest/handlers"
	oe "github.com/radanalyticsio/oshinko-cli/rest/helpers/errors"
	"github.com/radanalyticsio/oshinko-cli/rest/helpers/flags"
	"github.com/radanalyticsio/oshinko-cli/rest/helpers/logging"
	"github.com/radanalyticsio/oshinko-cli/rest/restapi/operations"
	"github.com/radanalyticsio/oshinko-cli/rest/restapi/operations/clusters"
	"github.com/radanalyticsio/oshinko-cli/rest/restapi/operations/server"
	"github.com/radanalyticsio/oshinko-cli/rest/version"
	"k8s.io/apiserver/pkg/util/logs"
)

// This file is safe to edit. Once it exists it will not be overwritten

func configureFlags(api *operations.OshinkoRestAPI) {
	api.CommandLineOptionsGroups = flags.GetLineOptionsGroups()
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

	logs.InitLogs()

	logging.Debug("Setting log level ", flags.LogLevel())
	flag.Set("v", flags.LogLevel())

	if logFile := flags.GetLogFile(); logFile != "" {
		err := logging.SetLoggerFile(logFile)
		if err != nil {
			logging.GetLogger().Println("unable to set log file;", err)
		}
	}

	// Print something if we are in debug mode
	logging.Debug("Debug mode enabled")

	api.Logger = logging.GetLogger().Printf

	logging.GetLogger().Println("Starting", version.GetAppName(), "version", version.GetVersion())

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
	corsHeaders := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "HEAD", "POST", "DELETE", "PUT", "OPTIONS"},
	})
	finalHandler = corsHeaders.Handler(finalHandler)
	return finalHandler
}
