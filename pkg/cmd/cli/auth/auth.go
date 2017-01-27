package auth

import (
	//"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"

	kapi "k8s.io/kubernetes/pkg/api"
	kapierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kclientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	kcmdconfig "k8s.io/kubernetes/pkg/kubectl/cmd/config"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/sets"

	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/cli/config"
	"github.com/openshift/origin/pkg/cmd/flagtypes"
	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/user/api"

)

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
	KClient            *kclient.Client

	// cert data to be used when authenticating
	CertFile string
	KeyFile  string

	Token string

	PathOptions *kcmdconfig.PathOptions
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

	o.PathOptions = config.NewPathOptions(cmd)

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
		//if term.IsTerminal(o.Reader) {
		//	for !o.serverProvided() {
		//		defaultServer := defaultClusterURL
		//		promptMsg := fmt.Sprintf("Server [%s]: ", defaultServer)
		//
		//		o.Server = cmdutil.PromptForStringWithDefault(o.Reader, o.Out, defaultServer, promptMsg)
		//	}
		//}
	}

	// normalize the provided server to a format expected by config
	serverNormalized, err := config.NormalizeServerURL(o.Server)
	if err != nil {
		return nil, err
	}
	o.Server = serverNormalized
	clientConfig.Host = o.Server

	if len(o.CAFile) > 0 {
		clientConfig.CAFile = o.CAFile

	} else {
		// check all cluster stanzas to see if we already have one with this URL that contains a client cert
		for _, cluster := range o.StartingKubeConfig.Clusters {
			if cluster.Server == clientConfig.Host {
				if len(cluster.CertificateAuthority) > 0 {
					clientConfig.CAFile = cluster.CertificateAuthority
					break
				}

				if len(cluster.CertificateAuthorityData) > 0 {
					clientConfig.CAData = cluster.CertificateAuthorityData
					break
				}
			}
		}
	}

	// ping to check if server is reachable
	osClient, err := client.New(clientConfig)
	if err != nil {
		return nil, err
	}

	result := osClient.Get().AbsPath("/").Do()
	if result.Error() != nil {
		switch {
		case o.InsecureTLS:
			clientConfig.Insecure = true
			// insecure, clear CA info
			clientConfig.CAFile = ""
			clientConfig.CAData = nil

		// certificate issue, prompt user for insecure connection
		case osclientcmd.IsCertificateAuthorityUnknown(result.Error()):
			// check to see if we already have a cluster stanza that tells us to use --insecure for this particular server.  If we don't, then prompt
			//clientConfigToTest := *clientConfig
			//clientConfigToTest.Insecure = true
			//matchingClusters := getMatchingClusters(clientConfigToTest, *o.StartingKubeConfig)
			//
			//if len(matchingClusters) > 0 {
			//	clientConfig.Insecure = true
			//
			//} else if term.IsTerminal(o.Reader) {
			fmt.Fprintln(o.Out, "The server uses a certificate signed by an unknown authority.")
			//	fmt.Fprintln(o.Out, "You can bypass the certificate check, but any data you send to the server could be intercepted by others.")
			//
			//	clientConfig.Insecure = cmdutil.PromptForBool(os.Stdin, o.Out, "Use insecure connections? (y/n): ")
			//	if !clientConfig.Insecure {
			//		return nil, fmt.Errorf(clientcmd.GetPrettyMessageFor(result.Error()))
			//	}
			//	// insecure, clear CA info
			//	clientConfig.CAFile = ""
			//	clientConfig.CAData = nil
			//	fmt.Fprintln(o.Out)
			//}

		default:
			return nil, result.Error()
		}
	}

	// check for matching api version
	if !o.APIVersion.IsEmpty() {
		clientConfig.GroupVersion = &o.APIVersion
	}

	o.Config = clientConfig
	//o.KClient = kclient.New(clientConfig)

	return o.Config, nil
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

				msg+=fmt.Sprintf("Logged into %q as %q using the token provided.\n\n", o.Config.Host, o.Username)
				return msg, nil
			}

			if !kapierrors.IsUnauthorized(err) {
				return "", err
			}

			return "", fmt.Errorf("The token provided is invalid or expired.\n\n")
		}
	} else {
		return "", fmt.Errorf("The token is not provided.\n\n")
	}

	msg+=fmt.Sprintf("Login successful.\n\n")

	return msg, nil
}

func (o *AuthOptions) GatherProjectInfo() (string,error) {
	var msg string
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
		msg+=fmt.Sprintf(`You don't have any projects. You can try to create a new project, by running

    $ oc new-project <projectname>

`)
		o.Project = ""

	case 1:
		o.Project = projectsItems[0].Name
		msg+=fmt.Sprintf("Using project %q.\n", o.Project)

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

		msg+=fmt.Sprintf( "You have access to the following projects and can switch between them with 'oc project <projectname>':\n\n")
		for _, p := range projects.List() {
			if o.Project == p {
				msg+=fmt.Sprintf("  * %s (current)\n", p)
			} else {
				msg+=fmt.Sprintf("  * %s\n", p)
			}
		}
		msg+=fmt.Sprintf("\n")
		msg+=fmt.Sprintf("Using project %q.\n", o.Project)
	}

	return msg, nil
}
