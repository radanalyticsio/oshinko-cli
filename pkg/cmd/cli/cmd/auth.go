package cmd

import (
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
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/cmd/util/tokencmd"
	"github.com/openshift/origin/pkg/user/api"
	//"github.com/docker/docker/cliconfig"
	"sort"
)

// RunProjects lists all projects a user belongs to
func (o *AuthOptions) RunClusters(currentProject string) error {
	_ = "breakpoint"

	kubeclient, err := kclient.New(o.Config)
	if err != nil {
		return err
	}
	var msg string
	clusters, err := getClusters(kubeclient, currentProject)
	if err == nil {
		clusterCount := len(clusters)
		if clusterCount <= 0 {
			msg += "There are no clusters in any projects. You can create a cluster with the 'new-cluster' command."
		} else if clusterCount > 0 {
			asterisk := ""
			count := 0
			sort.Sort(SortByClusterName(clusters))
			//fmt.Println(clusterCount)
			for _, cluster := range clusters {
				count = count + 1
				displayName := *(cluster.Name)
				workCount := *(cluster.WorkerCount)
				//fmt.Println(displayName)
				linebreak := "\n"

				msg += fmt.Sprintf(linebreak+asterisk+"%s \t  %d", displayName, workCount)
			}
		}

		fmt.Println(msg)
		return nil
	}

	return err
}

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

	} else if len(args) == 1 {
		if err := addr.Set(args[0]); err != nil {
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
		case clientcmd.IsCertificateAuthorityUnknown(result.Error()):
			// check to see if we already have a cluster stanza that tells us to use --insecure for this particular server.  If we don't, then prompt
			//clientConfigToTest := *clientConfig
			//clientConfigToTest.Insecure = true
			//matchingClusters := getMatchingClusters(clientConfigToTest, *o.StartingKubeConfig)
			//
			//if len(matchingClusters) > 0 {
			//	clientConfig.Insecure = true
			//
			//} else if term.IsTerminal(o.Reader) {
			//	fmt.Fprintln(o.Out, "The server uses a certificate signed by an unknown authority.")
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

func (o *AuthOptions) gatherAuthInfo() error {
	directClientConfig, err := o.getClientConfig()
	if err != nil {
		return err
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

				fmt.Fprintf(o.Out, "Logged into %q as %q using the token provided.\n\n", o.Config.Host, o.Username)
				return nil
			}

			if !kapierrors.IsUnauthorized(err) {
				return err
			}

			return fmt.Errorf("The token provided is invalid or expired.\n\n")
		}
	}

	// if a username was provided try to make use of it, but if a password were provided we force a token
	// request which will return a proper response code for that given password
	//if o.usernameProvided() && !o.passwordProvided() {
	//	// search all valid contexts with matching server stanzas to see if we have a matching user stanza
	//	kubeconfig := *o.StartingKubeConfig
	//	matchingClusters := getMatchingClusters(*clientConfig, kubeconfig)
	//
	//	for key, context := range o.StartingKubeConfig.Contexts {
	//		if matchingClusters.Has(context.Cluster) {
	//			clientcmdConfig := kclientcmd.NewDefaultClientConfig(kubeconfig, &kclientcmd.ConfigOverrides{CurrentContext: key})
	//			if kubeconfigClientConfig, err := clientcmdConfig.ClientConfig(); err == nil {
	//				if osClient, err := client.New(kubeconfigClientConfig); err == nil {
	//					if me, err := whoAmI(osClient); err == nil && (o.Username == me.Name) {
	//						clientConfig.BearerToken = kubeconfigClientConfig.BearerToken
	//						clientConfig.CertFile = kubeconfigClientConfig.CertFile
	//						clientConfig.CertData = kubeconfigClientConfig.CertData
	//						clientConfig.KeyFile = kubeconfigClientConfig.KeyFile
	//						clientConfig.KeyData = kubeconfigClientConfig.KeyData
	//
	//						o.Config = clientConfig
	//
	//						if key == o.StartingKubeConfig.CurrentContext {
	//							fmt.Fprintf(o.Out, "Logged into %q as %q using existing credentials.\n\n", o.Config.Host, o.Username)
	//						}
	//
	//						return nil
	//					}
	//				}
	//			}
	//		}
	//	}
	//}

	// if kubeconfig doesn't already have a matching user stanza...
	clientConfig.BearerToken = ""
	clientConfig.CertData = []byte{}
	clientConfig.KeyData = []byte{}
	clientConfig.CertFile = o.CertFile
	clientConfig.KeyFile = o.KeyFile
	token, err := tokencmd.RequestToken(o.Config, o.Reader, o.Username, o.Password)
	if err != nil {
		return err
	}
	clientConfig.BearerToken = token

	osClient, err := client.New(clientConfig)
	if err != nil {
		return err
	}

	me, err := whoAmI(osClient)
	if err != nil {
		return err
	}
	o.Username = me.Name
	o.Config = clientConfig
	fmt.Fprint(o.Out, "Login successful.\n\n")

	return nil
}

func (o *AuthOptions) gatherProjectInfo() error {
	me, err := o.whoAmI()
	if err != nil {
		return err
	}

	if o.Username != me.Name {
		return fmt.Errorf("current user, %v, does not match expected user %v", me.Name, o.Username)
	}

	oClient, err := client.New(o.Config)
	if err != nil {
		return err
	}

	projects, err := oClient.Projects().List(kapi.ListOptions{})
	if err != nil {
		return err
	}

	projectsItems := projects.Items

	switch len(projectsItems) {
	case 0:
		fmt.Fprintf(o.Out, `You don't have any projects. You can try to create a new project, by running

    $ oc new-project <projectname>

`)
		o.Project = ""

	case 1:
		o.Project = projectsItems[0].Name
		fmt.Fprintf(o.Out, "Using project %q.\n", o.Project)

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

		current, err := oClient.Projects().Get(namespace)
		if err != nil && !kapierrors.IsNotFound(err) && !clientcmd.IsForbidden(err) {
			return err
		}
		o.Project = current.Name

		fmt.Fprintf(o.Out, "You have access to the following projects and can switch between them with 'oc project <projectname>':\n\n")
		for _, p := range projects.List() {
			if o.Project == p {
				fmt.Fprintf(o.Out, "  * %s (current)\n", p)
			} else {
				fmt.Fprintf(o.Out, "  * %s\n", p)
			}
		}
		fmt.Fprintln(o.Out)
		fmt.Fprintf(o.Out, "Using project %q.\n", o.Project)
	}

	return nil
}

func (o *AuthOptions) GatherInfo() error {
	if err := o.gatherAuthInfo(); err != nil {
		return err
	}
	if err := o.gatherProjectInfo(); err != nil {
		return err
	}
	return nil
}

// RunLogin contains all the necessary functionality for the OpenShift cli login command
func RunLogin(cmd *cobra.Command, options *AuthOptions) error {
	if err := options.GatherInfo(); err != nil {
		return err
	}

	if err := options.RunClusters(options.Project); err != nil {
		return err
	}

	return nil
}

func NewCmdLogin(fullName string, f *osclientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	options := &AuthOptions{
		Reader: reader,
		Out:    out,
	}

	cmds := &cobra.Command{
		Use:   "get ",
		Short: "get cluster",
		//Long:    loginLong,
		//Example: fmt.Sprintf(loginExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(err)
			}

			err := RunLogin(cmd, options)

			if kapierrors.IsUnauthorized(err) {
				fmt.Fprintln(out, "Login failed (401 Unauthorized)")

				if err, isStatusErr := err.(*kapierrors.StatusError); isStatusErr {
					if details := err.Status().Details; details != nil {
						for _, cause := range details.Causes {
							fmt.Fprintln(out, cause.Message)
						}
					}
				}

				os.Exit(1)

			} else {
				kcmdutil.CheckErr(err)
			}
		},
	}
	return cmds
}
