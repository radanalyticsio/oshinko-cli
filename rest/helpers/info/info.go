package info

import (
	"fmt"
	"io/ioutil"
	"os"
	"github.com/radanalyticsio/oshinko-cli/rest/version"
)

const CA_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
const TOKEN_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/token"
const NS_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

func InAPod() bool {
	return os.Getenv("OSHINKO_REST_POD_NAME") != ""
}

func GetNamespace() (string, error) {
	if InAPod() {
		// Try the secrets file first
		// If that fails, fall back to env var
		ns, err := GetServiceAccountNS()
		if ns != nil && err == nil {
			return string(ns), err
		}
	}
	return os.Getenv("OSHINKO_CLUSTER_NAMESPACE"), nil
}

func GetSparkImage() string {
	image := os.Getenv("OSHINKO_CLUSTER_IMAGE")
	if image == "" {
		image = version.GetSparkImage()
	}
	return image
}

func GetKubeConfigPath() string {
	return os.Getenv("OSHINKO_KUBE_CONFIG")
}

func GetKubeProxyAddress() (string, error) {
	proxyHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	if len(proxyHost) == 0 {
		return "", fmt.Errorf("Unable to fetch KUBERNETES_SERVICE_HOST in Pod.")
	}
	return proxyHost, nil
}

func GetKubeProxyPort() (string, error) {
	proxyPort := os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(proxyPort) == 0 {
		return "", fmt.Errorf("Unable to fetch KUBERNETES_SERVICE_PORT in Pod.")
	}
	return proxyPort, nil
}

func GetServiceAccountCAPath() string {
	return CA_PATH
}

func GetServiceAccountTokenPath() string {
	return TOKEN_PATH
}

func GetServiceAccountNSPath() string {
	return NS_PATH
}

func GetServiceAccountToken() ([]byte, error) {
	token, err := ioutil.ReadFile(GetServiceAccountTokenPath())
	if err != nil {
		return nil, err
	}
	return token, err
}

func GetServiceAccountNS() ([]byte, error) {
	namespace, err := ioutil.ReadFile(GetServiceAccountNSPath())
	if err != nil {
		return nil, err
	}
	return namespace, err
}

func GetWebServiceName() string {
	return os.Getenv("OSHINKO_WEB_NAME")
}
