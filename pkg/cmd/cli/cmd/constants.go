package cmd

const (
	ScaleCmdUsage   = "scale <NAME>"
	ScaleCmdShort   = "Scale spark cluster by name"
	clustersExample = `  # Display the spark cluster %[1]s`
	clustersLong = `

Display information about the spark clusters on the server.`
	createLong = `Create a resource by filename or stdin

JSON and YAML formats are accepted.`

	createExample = `  # Create a cluster using the data in cluster.json.
  %[1]s create -f cluster.json

  # Create a cluster based on the JSON passed into stdin.
  cat cluster.json | %[1]s create -f -`
)
