package cmd

const (
	ScaleCmdUsage   = "scale <NAME>"
	ScaleCmdShort   = "Scale spark cluster by name."
	maxWorkers      = 25
	clustersExample = `  # Display the spark cluster %[1]s`

	masterConfigMsg = "Error processing spark master configuration value"
	workerConfigMsg = "Error processing spark worker configuration value"

	nameSpaceMsg = "Cannot determine target openshift namespace"
	clientMsg    = "Unable to create an openshift client"

	typeLabel    = "oshinko-type"
	clusterLabel = "oshinko-cluster"

	workerType = "worker"
	masterType = "master"
	webuiType  = "webui"

	masterPortName = "spark-master"
	webPortName    = "spark-webui"

	mDepConfigMsg  = "Unable to create master deployment configuration"
	wDepConfigMsg  = "Unable to create worker deployment configuration"
	masterSrvMsg   = "Unable to create spark master service endpoint"
	imageMsg       = "Cannot determine name of spark image"
	respMsg        = "Created cluster but failed to construct a response object"
	defaultImage   = "radanalyticsio/openshift-spark"
	defaultProject = "default"

	masterPort = 7077
	webPort    = 8080

	defaultsparkconfdir = "/etc/oshinko-spark-configs"

	// The suffix to add to the spark master hostname (clustername) for the web service
	webServiceSuffix = "-ui"
	clustersLong     = `
Display information about the spark clusters on the server.`
	createLong = `Create a resource by filename or stdin

JSON and YAML formats are accepted.`

	createExample = `  # Create a cluster using the data in cluster.json.
  %[1]s create -f cluster.json

  # Create a cluster based on the JSON passed into stdin.
  cat cluster.json | %[1]s create -f -`
)
