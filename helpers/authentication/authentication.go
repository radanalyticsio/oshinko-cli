package authentication

import (
	"net"

	_ "github.com/openshift/origin/pkg/api/install"
	"github.com/openshift/origin/pkg/client"
	oclient "github.com/openshift/origin/pkg/client"
	serverapi "github.com/openshift/origin/pkg/cmd/server/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/crypto"

	"github.com/redhatanalytics/oshinko-rest/helpers/info"
)

func SAConfig() (*restclient.Config, error) {
	//fetch proxy IP + port
	host, err := info.GetKubeProxyAddress()
	if err != nil {
		return nil, err
	}
	port, err := info.GetKubeProxyPort()
	if err != nil {
		return nil, err
	}
	token, err := info.GetServiceAccountToken()
	if err != nil {
		return nil, err
	}
	tlsClientConfig := restclient.TLSClientConfig{}
	CAFile := info.GetServiceAccountCAPath()
	_, err = crypto.CertPoolFromFile(CAFile)
	if err != nil {
		return nil, err
	} else {
		tlsClientConfig.CAFile = CAFile
	}

	return &restclient.Config{
		Host:            "https://" + net.JoinHostPort(host, port),
		BearerToken:     string(token),
		TLSClientConfig: tlsClientConfig,
	}, nil
}

func GetKubeClient() (*kclient.Client, error) {

	// If the configfile option has been
	// set, use it to get a client otherwise use the
	// serviceaccount for authentication
	configfile := info.GetKubeConfigPath()
	if configfile != "" {
		client, _, err := serverapi.GetKubeClient(configfile)
		return client, err
	} else {
		saConfig, err := SAConfig()
		if err != nil {
			return nil, err
		}
		client, err := kclient.New(saConfig)
		if err != nil {
			return nil, err
		}
		return client, err
	}
}

func GetOpenShiftClient() (*client.Client, error) {

	// If the configfile option has been
	// set, use it to get a client otherwise use the
	// serviceaccount for authentication
	configfile := info.GetKubeConfigPath()
	if configfile != "" {
		client, _, err := serverapi.GetOpenShiftClient(configfile)
		return client, err
	} else {
		saConfig, err := SAConfig()
		if err != nil {
			return nil, err
		}
		client, err := oclient.New(saConfig)
		if err != nil {
			return nil, err
		}
		return client, err
	}
}
