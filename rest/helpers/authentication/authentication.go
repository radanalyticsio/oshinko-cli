package authentication

import (
	"net"

	_ "github.com/openshift/origin/pkg/api/install"
	restclient "k8s.io/client-go/rest"
	kclientcmd "k8s.io/client-go/tools/clientcmd"
	certutil "k8s.io/client-go/util/cert"
	"github.com/radanalyticsio/oshinko-cli/rest/helpers/info"
)

func GetConfig() (*restclient.Config, error) {

	credentials, err := kclientcmd.NewDefaultClientConfigLoadingRules().Load()
	clusterConfig, err := kclientcmd.NewDefaultClientConfig(*credentials, &kclientcmd.ConfigOverrides{}).ClientConfig()
	if err == nil  {
		return clusterConfig, nil
	}

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
	if _, err := certutil.NewPool(CAFile); err != nil {
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

