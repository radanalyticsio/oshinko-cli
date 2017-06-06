package auth

import (
	//"errors"
	"crypto/x509"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	kapi "k8s.io/kubernetes/pkg/api"
	kapierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	kclientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/sets"

	"github.com/openshift/origin/pkg/client"
	cliconfig "github.com/openshift/origin/pkg/cmd/cli/config"
	"github.com/openshift/origin/pkg/cmd/flagtypes"
	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"

	"bytes"
	"crypto/tls"
	"github.com/openshift/origin/pkg/cmd/util/term"
	"github.com/openshift/origin/pkg/user/api"
	kterm "k8s.io/kubernetes/pkg/util/term"
)

const defaultClusterURL = "https://localhost:8443"

//=====================================
type AuthOptions struct {
	Server      string
	CAFile      string
	InsecureTLS bool
	APIVersion  unversioned.GroupVersion

	// flags and printing helpers
	Username string
	Password string
	Project  string

	// infra
	StartingKubeConfig *kclientcmdapi.Config
	DefaultNamespace   string
	Config             *restclient.Config
	Reader             io.Reader
	Out                io.Writer
	Client             *client.Client
	KClient            kclient.Interface

	// cert data to be used when authenticating
	CertFile    string
	KeyFile     string
	Token       string
	PathOptions *kclientcmd.PathOptions
	ClientFn    func() (*client.Client, kclient.Interface, error)
}

func (o *AuthOptions) tokenProvided() bool {
	return len(o.Token) > 0
}

func (o AuthOptions) whoAmI() (*api.User, error) {
	client, err := client.New(o.Config)
	if err != nil {
		return nil, err
	}
	return whoAmI(client)
}

func (o *AuthOptions) serverProvided() bool {
	return (len(o.Server) > 0)
}

func (o *AuthOptions) Complete(f *osclientcmd.Factory, cmd *cobra.Command, args []string) error {
	kubeconfig, err := f.OpenShiftClientConfig.RawConfig()
	o.StartingKubeConfig = &kubeconfig
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// build a valid object to use if we failed on a non-existent file
		o.StartingKubeConfig = kclientcmdapi.NewConfig()
	}

	addr := flagtypes.Addr{Value: "localhost:8443", DefaultScheme: "https", DefaultPort: 8443, AllowPrefix: true}.Default()

	if serverFlag := kcmdutil.GetFlagString(cmd, "server"); len(serverFlag) > 0 {
		if err := addr.Set(serverFlag); err != nil {
			return err
		}
		o.Server = addr.String()

	} else if len(o.Server) == 0 {
		if defaultContext, defaultContextExists := o.StartingKubeConfig.Contexts[o.StartingKubeConfig.CurrentContext]; defaultContextExists {
			if cluster, exists := o.StartingKubeConfig.Clusters[defaultContext.Cluster]; exists {
				o.Server = cluster.Server
			}
		}
	}

	o.CertFile = kcmdutil.GetFlagString(cmd, "client-certificate")
	o.KeyFile = kcmdutil.GetFlagString(cmd, "client-key")
	apiVersionString := kcmdutil.GetFlagString(cmd, "api-version")
	o.APIVersion = unversioned.GroupVersion{}

	namespaceFlag := kcmdutil.GetFlagString(cmd, "namespace")
	if namespaceFlag != "" {
		o.Project = namespaceFlag
	}

	// if the API version isn't explicitly passed, use the API version from the default context (same rules as the server above)
	if len(apiVersionString) == 0 {
		if defaultContext, defaultContextExists := o.StartingKubeConfig.Contexts[o.StartingKubeConfig.CurrentContext]; defaultContextExists {
			if cluster, exists := o.StartingKubeConfig.Clusters[defaultContext.Cluster]; exists {
				apiVersionString = cluster.APIVersion
			}
		}
	}

	o.APIVersion, err = unversioned.ParseGroupVersion(apiVersionString)
	if err != nil {
		return err
	}

	o.CAFile = kcmdutil.GetFlagString(cmd, "certificate-authority")
	o.InsecureTLS = kcmdutil.GetFlagBool(cmd, "insecure-skip-tls-verify")
	o.Token = kcmdutil.GetFlagString(cmd, "token")
	o.DefaultNamespace, _, _ = f.OpenShiftClientConfig.Namespace()
	o.PathOptions = cliconfig.NewPathOptions(cmd)

	// for looking at kubeconfig
	o.Config, err = f.OpenShiftClientConfig.ClientConfig()
	if err != nil {
		var errstrings []string
		if strings.Contains(err.Error(), "Missing or incomplete configuration info") {
			errstrings = append(errstrings, "oshinko-cli cannot find KUBECONFIG file.Please login or provide --token value.")
		} else {
			errstrings = append(errstrings, err.Error())
		}
		return fmt.Errorf(strings.Join(errstrings, "\n"))
	}
	o.ClientFn = func() (*client.Client, kclient.Interface, error) {
		return f.Clients()
	}

	return nil
}

func whoAmI(client *client.Client) (*api.User, error) {
	me, err := client.Users().Get("~")
	if err != nil {
		return nil, err
	}

	return me, nil
}

func (o *AuthOptions) getClientConfig() (*restclient.Config, error) {
	if o.Config != nil {
		return o.Config, nil
	}

	clientConfig := &restclient.Config{}

	if len(o.Server) == 0 {
		// we need to have a server to talk to
		if kterm.IsTerminal(o.Reader) {
			for !o.serverProvided() {
				defaultServer := defaultClusterURL
				promptMsg := fmt.Sprintf("Server [%s]: ", defaultServer)
				o.Server = term.PromptForStringWithDefault(o.Reader, o.Out, defaultServer, promptMsg)
			}
		}
	}

	// normalize the provided server to a format expected by config
	serverNormalized, err := cliconfig.NormalizeServerURL(o.Server)
	if err != nil {
		return nil, err
	}
	o.Server = serverNormalized
	clientConfig.Host = o.Server

	// use specified CA or find existing CA
	if len(o.CAFile) > 0 {
		clientConfig.CAFile = o.CAFile
		clientConfig.CAData = nil
	} else if caFile, caData, ok := findExistingClientCA(clientConfig.Host, *o.StartingKubeConfig); ok {
		clientConfig.CAFile = caFile
		clientConfig.CAData = caData
	}
	// try to TCP connect to the server to make sure it's reachable, and discover
	// about the need of certificates or insecure TLS
	if err := dialToServer(*clientConfig); err != nil {
		switch err.(type) {
		// certificate authority unknown, check or prompt if we want an insecure
		// connection or if we already have a cluster stanza that tells us to
		// connect to this particular server insecurely
		case x509.UnknownAuthorityError, x509.HostnameError, x509.CertificateInvalidError:
			if o.InsecureTLS ||
				hasExistingInsecureCluster(*clientConfig, *o.StartingKubeConfig) ||
				promptForInsecureTLS(o.Reader, o.Out, err) {
				clientConfig.Insecure = true
				clientConfig.CAFile = ""
				clientConfig.CAData = nil
			} else {
				return nil, osclientcmd.GetPrettyErrorForServer(err, o.Server)
			}
		// TLS record header errors, like oversized record which usually means
		// the server only supports "http"
		case tls.RecordHeaderError:
			return nil, osclientcmd.GetPrettyErrorForServer(err, o.Server)
		default:
			// suggest the port used in the cluster URL by default, in case we're not already using it
			host, port, parsed, err1 := getHostPort(o.Server)
			_, defaultClusterPort, _, err2 := getHostPort(defaultClusterURL)
			if err1 == nil && err2 == nil && port != defaultClusterPort {
				parsed.Host = net.JoinHostPort(host, defaultClusterPort)
				return nil, fmt.Errorf("%s\nYou may want to try using the default cluster port: %s", err.Error(), parsed.String())
			}
			return nil, err
		}
	}

	// check for matching api version
	if !o.APIVersion.Empty() {
		clientConfig.GroupVersion = &o.APIVersion
	}
	o.Config = clientConfig
	return o.Config, nil
}

// getHostPort returns the host and port parts of the given URL string. It's
// expected that the provided URL is already normalized (always has host and port).
func getHostPort(hostURL string) (string, string, *url.URL, error) {
	parsedURL, err := url.Parse(hostURL)
	if err != nil {
		return "", "", nil, err
	}
	host, port, err := net.SplitHostPort(parsedURL.Host)
	return host, port, parsedURL, err
}

func promptForInsecureTLS(reader io.Reader, out io.Writer, reason error) bool {
	var insecureTLSRequestReason string
	if reason != nil {
		switch reason.(type) {
		case x509.UnknownAuthorityError:
			insecureTLSRequestReason = "The server uses a certificate signed by an unknown authority."
		case x509.HostnameError:
			insecureTLSRequestReason = fmt.Sprintf("The server is using a certificate that does not match its hostname: %s", reason.Error())
		case x509.CertificateInvalidError:
			insecureTLSRequestReason = fmt.Sprintf("The server is using an invalid certificate: %s", reason.Error())
		}
	}
	var input bool
	if kterm.IsTerminal(reader) {
		if len(insecureTLSRequestReason) > 0 {
			fmt.Fprintln(out, insecureTLSRequestReason)
		}
		fmt.Fprintln(out, "You can bypass the certificate check, but any data you send to the server could be intercepted by others.")
		input = term.PromptForBool(os.Stdin, out, "Use insecure connections? (y/n): ")
		fmt.Fprintln(out)
	}
	return input
}

// dialToServer takes the Server URL from the given clientConfig and dials to
// make sure the server is reachable. Note the config received is not mutated.
func dialToServer(clientConfig restclient.Config) error {
	// take a RoundTripper based on the config we already have (TLS, proxies, etc)
	rt, err := restclient.TransportFor(&clientConfig)
	if err != nil {
		return err
	}

	parsedURL, err := url.Parse(clientConfig.Host)
	if err != nil {
		return err
	}

	// Do a HEAD request to serverPathToDial to make sure the server is alive.
	// We don't care about the response, any err != nil is valid for the sake of reachability.
	serverURLToDial := (&url.URL{Scheme: parsedURL.Scheme, Host: parsedURL.Host, Path: "/"}).String()
	req, err := http.NewRequest("HEAD", serverURLToDial, nil)
	if err != nil {
		return err
	}

	res, err := rt.RoundTrip(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	return nil
}

// findExistingClientCA returns *either* the existing client CA file name as a string,
// *or* data in a []byte for a given host, and true if it exists in the given config
func findExistingClientCA(host string, kubeconfig kclientcmdapi.Config) (string, []byte, bool) {
	for _, cluster := range kubeconfig.Clusters {
		if cluster.Server == host {
			if len(cluster.CertificateAuthority) > 0 {
				return cluster.CertificateAuthority, nil, true
			}
			if len(cluster.CertificateAuthorityData) > 0 {
				return "", cluster.CertificateAuthorityData, true
			}
		}
	}
	return "", nil, false
}

func hasExistingInsecureCluster(clientConfigToTest restclient.Config, kubeconfig kclientcmdapi.Config) bool {
	clientConfigToTest.Insecure = true
	matchingClusters := getMatchingClusters(clientConfigToTest, kubeconfig)
	return len(matchingClusters) > 0
}

// getMatchingClusters examines the kubeconfig for all clusters that point to the same server
func getMatchingClusters(clientConfig restclient.Config, kubeconfig kclientcmdapi.Config) sets.String {
	ret := sets.String{}

	for key, cluster := range kubeconfig.Clusters {
		if (cluster.Server == clientConfig.Host) && (cluster.InsecureSkipTLSVerify == clientConfig.Insecure) && (cluster.CertificateAuthority == clientConfig.CAFile) && (bytes.Compare(cluster.CertificateAuthorityData, clientConfig.CAData) == 0) {
			ret.Insert(key)
		}
	}

	return ret
}

func (o *AuthOptions) GatherAuthInfo() (string, error) {
	var msg string
	directClientConfig, err := o.getClientConfig()
	if err != nil {
		return "", err
	}

	// make a copy and use it to avoid mutating the original
	t := *directClientConfig
	clientConfig := &t

	// if a token were explicitly provided, try to use it
	if o.tokenProvided() {
		clientConfig.BearerToken = o.Token
		if osClient, err := client.New(clientConfig); err == nil {
			me, err := whoAmI(osClient)
			if err == nil {
				o.Username = me.Name
				o.Config = clientConfig

				clientConfig.CertData = []byte{}
				clientConfig.KeyData = []byte{}
				clientConfig.CertFile = o.CertFile
				clientConfig.KeyFile = o.KeyFile

				osClient, err := client.New(clientConfig)
				if err != nil {
					return "", err
				}
				o.Client = osClient

				kubeclient, err := kclient.New(o.Config)
				if err != nil {
					return "", err
				}
				o.KClient = kubeclient

				me, err := whoAmI(osClient)
				if err != nil {
					return "", err
				}
				o.Username = me.Name
				o.Config = clientConfig

				msg += fmt.Sprintf("Logged into %q as %q using the token provided.\n\n", o.Config.Host, o.Username)
				return msg, nil
			}

			if !kapierrors.IsUnauthorized(err) {
				return "", err
			}

			return "", fmt.Errorf("The token provided is invalid or expired.\n\n")
		}
	} else {
		config := o.StartingKubeConfig
		currentContext := config.Contexts[config.CurrentContext]
		var currentProject string
		if currentContext != nil {
			currentProject = currentContext.Namespace
		}

		var err error
		o.Client, o.KClient, err = o.ClientFn()
		if err != nil {
			return msg, err
		}
		me, err := whoAmI(o.Client)
		if err != nil {
			return "", err
		}
		o.Username = me.Name
		msg += fmt.Sprintf("Logged into %q as %q using the token provided.\n\n", o.Config.Host, o.Username)

		switch err := confirmProjectAccess(currentProject, o.Client, o.KClient); {
		case osclientcmd.IsForbidden(err):
			return msg, fmt.Errorf("you do not have rights to view project %q.", currentProject)
		case kapierrors.IsNotFound(err):
			return msg, fmt.Errorf("the project %q specified in your config does not exist.", currentProject)
		case err != nil:
			return msg, err
		}

		defaultContextName := cliconfig.GetContextNickname(currentContext.Namespace, currentContext.Cluster, currentContext.AuthInfo)

		// if they specified a project name and got a generated context, then only show the information they care about.  They won't recognize
		// a context name they didn't choose
		if config.CurrentContext == defaultContextName {
			msg += fmt.Sprintf("Using project %q on server %q.\n", currentProject, o.Config.Host)

		} else {
			msg += fmt.Sprintf("Using project %q from context named %q on server %q.\n", currentProject, config.CurrentContext, o.Config.Host)
		}
	}

	msg += fmt.Sprintf("Login successful.\n\n")
	return msg, nil
}

func confirmProjectAccess(currentProject string, oClient *client.Client, kClient kclient.Interface) error {
	_, projectErr := oClient.Projects().Get(currentProject)
	if !kapierrors.IsNotFound(projectErr) {
		return projectErr
	}

	// at this point we know the error is a not found, but we'll test namespaces just in case we're running on kube
	if _, err := kClient.Namespaces().Get(currentProject); err == nil {
		return nil
	}

	// otherwise return the openshift error default
	return projectErr
}

func (o *AuthOptions) GatherProjectInfo() (string, error) {
	var msg string
	if o.Project != "" {
		return fmt.Sprintf("Using project %q.\n", o.Project), nil
	}
	me, err := o.whoAmI()
	if err != nil {
		return "", err
	}

	if o.Username != me.Name {
		return "", fmt.Errorf("current user, %v, does not match expected user %v", me.Name, o.Username)
	}

	projects, err := o.Client.Projects().List(kapi.ListOptions{})
	if err != nil {
		return "", err
	}

	projectsItems := projects.Items

	switch len(projectsItems) {
	case 0:
		msg += fmt.Sprintf(`You don't have any projects. You can try to create a new project, by running

    $ oc new-project <projectname>

`)
		o.Project = ""

	case 1:
		o.Project = projectsItems[0].Name
		msg += fmt.Sprintf("Using project %q.\n", o.Project)

	default:
		projects := sets.String{}
		for _, project := range projectsItems {
			projects.Insert(project.Name)
		}

		namespace := o.DefaultNamespace
		if !projects.Has(namespace) {
			if namespace != kapi.NamespaceDefault && projects.Has(kapi.NamespaceDefault) {
				namespace = kapi.NamespaceDefault
			} else {
				namespace = projects.List()[0]
			}
		}

		current, err := o.Client.Projects().Get(namespace)
		if err != nil && !kapierrors.IsNotFound(err) && !osclientcmd.IsForbidden(err) {
			return "", err
		}
		o.Project = current.Name

		msg += fmt.Sprintf("You have access to the following projects and can switch between them with 'oc project <projectname>':\n\n")
		for _, p := range projects.List() {
			if o.Project == p {
				msg += fmt.Sprintf("  * %s (current)\n", p)
			} else {
				msg += fmt.Sprintf("  * %s\n", p)
			}
		}
		msg += fmt.Sprintf("\n")
		msg += fmt.Sprintf("Using project %q.\n", o.Project)
	}

	return msg, nil
}
