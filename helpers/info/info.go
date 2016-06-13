package info

import (
	"os"
)

func GetNamespace() (string, error) {
	// TODO if we're running in a pod is there a cool way to get the current namespace?
	return os.Getenv("OSHINKO_CLUSTER_NAMESPACE"), nil
}

func GetSparkImage() (string, error) {
	// TODO is there a good well-known location for a spark image?
	return os.Getenv("OSHINKO_CLUSTER_IMAGE"), nil
}
