package authentication

import (
	"errors"
	_ "github.com/openshift/origin/pkg/api/install"
	"github.com/openshift/origin/pkg/client"
	serverapi "github.com/openshift/origin/pkg/cmd/server/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"os"
)

func GetKubeClient() (*kclient.Client, error) {

	// If the configfile option has been
	// set, use it to get a client otherwise use the
	// serviceaccount for authentication
	configfile := os.Getenv("OSHINKO_KUBE_CONFIG")
	if configfile != "" {
		client, _, err := serverapi.GetKubeClient(configfile)
		return client, err
	} else {
		// TODO whatever we have to do here to get a client using the serviceaccount
		return nil, errors.New("OSHINKO_KUBE_CONFIG env var is required at this time")
	}
}

func GetOpenShiftClient() (*client.Client, error) {

	// If the configfile option has been
	// set, use it to get a client otherwise use the
	// serviceaccount for authentication
	configfile := os.Getenv("OSHINKO_KUBE_CONFIG")
	if configfile != "" {
		client, _, err := serverapi.GetOpenShiftClient(configfile)
		return client, err
	} else {
		// TODO whatever we have to do here to get a client using the serviceaccount
		return nil, errors.New("OSHINKO_KUBE_CONFIG env var is required at this time")
	}
}
