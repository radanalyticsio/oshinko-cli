package auth

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"time"
	"strings"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	kclientcmd "k8s.io/client-go/tools/clientcmd"

	kclientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/apimachinery/pkg/util/sets"
	rest "k8s.io/client-go/rest"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	cliconfig "github.com/openshift/origin/pkg/oc/cli/config"
	clientcfg "github.com/openshift/origin/pkg/client/config"
	"github.com/openshift/origin/pkg/cmd/flagtypes"
	osclientcmd "github.com/openshift/origin/pkg/oc/cli/util/clientcmd"
	projectclient "github.com/openshift/client-go/project/clientset/versioned"
	//certutil "k8s.io/client-go/util/cert"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	projectinternalversion "github.com/openshift/origin/pkg/project/generated/internalclientset/typed/project/internalversion"
)

const defaultClusterURL = "https://localhost:8443"

//=====================================
type AuthOptions struct {
	Server      string
	CAFile      string
	InsecureTLS bool
	//APIVersion  unversioned.GroupVersion

	// flags and printing helpers
	Username string
	Password string
	Project  string

	// infra
	StartingKubeConfig *kclientcmdapi.Config
	DefaultNamespace   string
	Config             *rest.Config

	KubeClient			internalclientset.Interface
	ProjectClient 		projectinternalversion.ProjectInterface
	Reader             io.Reader
	Out                io.Writer

	// cert data to be used when authenticating
	CertFile    string
	KeyFile     string
	Token       string
	PathOptions *kclientcmd.PathOptions
	CommandName    string
	RequestTimeout time.Duration
}

func (o *AuthOptions) tokenProvided() bool {
	return len(o.Token) > 0
}

func (o *AuthOptions) serverProvided() bool {
	return (len(o.Server) > 0)
}

func (o *AuthOptions) Complete(f *osclientcmd.Factory, cmd *cobra.Command, args []string, commandName string) error {
	kubeconfig, err := f.OpenShiftClientConfig().RawConfig()
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

	namespaceFlag := kcmdutil.GetFlagString(cmd, "namespace")
	if namespaceFlag != "" {
		o.Project = namespaceFlag
	}

	o.CAFile = kcmdutil.GetFlagString(cmd, "certificate-authority")
	o.InsecureTLS = kcmdutil.GetFlagBool(cmd, "insecure-skip-tls-verify")
	o.Token = kcmdutil.GetFlagString(cmd, "token")
	o.DefaultNamespace, _, _ = f.OpenShiftClientConfig().Namespace()
	o.PathOptions = cliconfig.NewPathOptions(cmd)

	//Look for kubeconfig
	o.Config, err = o.getClientConfig()
	if err != nil {
		var errstrings []string
		if strings.Contains(err.Error(), "could not load client configuration") {
			//we have no kubeconfig
			//do we have token ?
			if !o.tokenProvided(){
				errstrings = append(errstrings, "oshinko-cli cannot find KUBECONFIG file.Please login or provide --token value.")
			} else {
				//get from token
				o.Config, err = o.createClientConfig()
				if err != nil {
					errstrings = append(errstrings, err.Error())
				}
			}
		} else {
			errstrings = append(errstrings, err.Error())
		}
		if len(errstrings)!=0 {
			return fmt.Errorf(strings.Join(errstrings, "\n"))
		}
	}
	o.KubeClient, err = f.ClientSet()
	if err != nil {
		return err
	}
	projectClient, err := f.OpenshiftInternalProjectClient()
	if err != nil {
		return err
	}
	o.ProjectClient = projectClient.Project()


	return nil
}

func (o *AuthOptions) createClientConfig() (*rest.Config, error) {
	tlsClientConfig := rest.TLSClientConfig{}
	if len(o.CertFile) == 0 {
		return nil, fmt.Errorf("Certificate File needed")
	}
	tlsClientConfig.CertFile = o.CertFile
	if len(o.KeyFile) == 0 {
		return nil, fmt.Errorf("Key File needed")
	}
	tlsClientConfig.KeyFile = o.KeyFile
	if len(o.CAFile) == 0 {
		return nil, fmt.Errorf("CA File needed")
	}
	tlsClientConfig.CAFile = o.CAFile
	return &rest.Config{
		Host:            o.Server,
		BearerToken:     string(o.Token),
		TLSClientConfig: tlsClientConfig,
	}, nil
}

func (o *AuthOptions) getClientConfig() (*rest.Config, error) {
	clusterConfig, err := rest.InClusterConfig()
	if err == nil {
		return clusterConfig, nil
	}

	credentials, err := kclientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return nil, fmt.Errorf("could not load credentials from config>: %v", err)
	}

	clusterConfig, err = kclientcmd.NewDefaultClientConfig(*credentials, &kclientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not load client configuration: %v", err)
	}
	return clusterConfig, nil
}


func (o *AuthOptions) GatherAuthInfo() (string, error) {
	var msg string
	directClientConfig := o.Config


	// make a copy and use it to avoid mutating the original
	t := *directClientConfig
	clientConfig := &t

	// if a token were explicitly provided, try to use it
	if o.tokenProvided() {
		clientConfig.BearerToken = o.Token
			me, err := whoAmI(clientConfig)
			if err == nil {
				o.Username = me.Name
				//fmt.Println(me.Name)
				clientConfig.CertData = []byte{}
				clientConfig.KeyData = []byte{}
				clientConfig.CertFile = o.CertFile
				clientConfig.KeyFile = o.KeyFile

				o.Config = clientConfig

				msg += fmt.Sprintf("Logged into %q as %q using the token provided.\n\n", o.Config.Host, o.Username)
				return msg, nil
			}

			if !kapierrors.IsUnauthorized(err) {
				return "", err
			}

			return "", fmt.Errorf("The token provided is invalid or expired.\n\n")
	} else {
		//Only use config for contexts
		config := o.StartingKubeConfig
		currentContext := config.Contexts[config.CurrentContext]
		var currentProject string
		if currentContext != nil {
			currentProject = currentContext.Namespace
		}

		var err error


		me, err := whoAmI(o.Config)
		if err != nil {
			return "", err
		}
		o.Username = me.Name
		msg += fmt.Sprintf("Logged into %q as %q using the token provided.\n\n", o.Config.Host, o.Username)

		switch err := confirmProjectAccess(currentProject, o.ProjectClient, o.KubeClient); {
		case osclientcmd.IsForbidden(err):
			return msg, fmt.Errorf("you do not have rights to view project %q.", currentProject)
		case kapierrors.IsNotFound(err):
			return msg, fmt.Errorf("the project %q specified in your config does not exist.", currentProject)
		case err != nil:
			return msg, err
		}

		defaultContextName := clientcfg.GetContextNickname(currentContext.Namespace, currentContext.Cluster, currentContext.AuthInfo)

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


/*
#	Who Am I?
#	Which Project Am I in ?
#	Do I have permissions ?
#	This method needs a valid ClientConfig
 */
func (o *AuthOptions) GatherProjectInfo() (string, error) {
	var msg string
	if o.Project != "" {
		return fmt.Sprintf("Using project %q.\n", o.Project), nil
	}
	me, err := whoAmI(o.Config)
	if err != nil {
		return "", err
	}

	if o.Username != me.Name {
		return "", fmt.Errorf("current user, %v, does not match expected user %v", me.Name, o.Username)
	}

	projectClient, err := projectclient.NewForConfig(o.Config)

	if err != nil {
		return "", err
	}

	projectsList, err := projectClient.ProjectV1().Projects().List(metav1.ListOptions{})
	// if we're running on kube (or likely kube), just set it to "default"
	if kerrors.IsNotFound(err) || kerrors.IsForbidden(err) {
		msg += fmt.Sprintf( "Using \"default\".  You can switch projects with:\n\n '%s project <projectname>'\n", o.CommandName)
		o.Project = "default"
		return msg, nil
	}
	if err != nil {
		return "", err
	}

	projectsItems := projectsList.Items
	projects := sets.String{}
	for _, project := range projectsItems {
		//fmt.Println(project.Name)
		projects.Insert(project.Name)
	}

	if len(o.DefaultNamespace) > 0 && !projects.Has(o.DefaultNamespace) {
		// Attempt a direct get of our current project in case it hasn't appeared in the list yet
		if currentProject, err := projectClient.ProjectV1().Projects().Get(o.DefaultNamespace, metav1.GetOptions{}); err == nil {
			// If we get it successfully, add it to the list
			projectsItems = append(projectsItems, *currentProject)
			projects.Insert(currentProject.Name)
		}
	}

	switch len(projectsItems) {
	case 0:
		msg += fmt.Sprintf(`You don't have any projects. You can try to create a new project, by running

    $ oc new-project <projectname>

`)
		o.Project = ""
		return "", fmt.Errorf("There are no projects for this user.Please create a Project")

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
			if namespace != metav1.NamespaceDefault && projects.Has(metav1.NamespaceDefault) {
				namespace = metav1.NamespaceDefault
			} else {
				namespace = projects.List()[0]
			}
		}

		current, err := projectClient.ProjectV1().Projects().Get(namespace, metav1.GetOptions{})
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
