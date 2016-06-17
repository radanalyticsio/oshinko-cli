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

	if info.InAPod() {
		saConfig, err := SAConfig()
		if err != nil {
			return nil, err
		}
		client, err := kclient.New(saConfig)
		if err != nil {
			return nil, err
		}
		return client, err
	} else {
		client, _, err := serverapi.GetKubeClient(info.GetKubeConfigPath())
		return client, err
	}
}

func GetOpenShiftClient() (*client.Client, error) {

	if info.InAPod() {
		saConfig, err := SAConfig()
		if err != nil {
			return nil, err
		}
		client, err := oclient.New(saConfig)
		if err != nil {
			return nil, err
		}
		return client, err
	} else {
		client, _, err := serverapi.GetOpenShiftClient(info.GetKubeConfigPath())
		return client, err
	}
}
