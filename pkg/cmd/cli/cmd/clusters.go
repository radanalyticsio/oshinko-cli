package cmd

import (
	"fmt"
	"io"
	"sort"
	//"strconv"

	"k8s.io/kubernetes/pkg/client/restclient"
	//kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	kubecmdconfig "k8s.io/kubernetes/pkg/kubectl/cmd/config"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/radanalyticsio/oshinko-rest/restapi/operations/clusters"
	kapi "k8s.io/kubernetes/pkg/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/sets"

	"github.com/openshift/origin/pkg/client"
	cliconfig "github.com/openshift/origin/pkg/cmd/cli/config"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	//"github.com/openshift/origin/pkg/project/api"

	"github.com/spf13/cobra"
)

type ClusterOptions struct {
	Config       clientcmdapi.Config
	ClientConfig *restclient.Config
	Client       *client.Client
	KClient      *kclient.Client
	Out          io.Writer
	PathOptions  *kubecmdconfig.PathOptions

	DisplayShort bool
}

// SortByProjectName is sort
type SortByClusterName []*clusters.ClustersItems0

//func (p SortByClusterName) Len() int {
//	return len(p)
//}
//func (p SortByClusterName) Swap(i, j int) {
//	p[i], p[j] = p[j], p[i]
//}
//func (p SortByClusterName) Less(i, j int) bool {
//	return p[i].Name < p[j].Name
//}
func (p SortByClusterName) Len() int {
	return len(p)
}
func (p SortByClusterName) Swap(i, j int) {
	*p[i], *p[j] = *p[j], *p[i]
}
func (p SortByClusterName) Less(i, j int) bool {
	return *(p[i].Name) < *(p[j].Name)
}

const (
	clustersLong = `
Display information about the spark clusters on the server.`
	clustersExample = `  # Display the spark cluster %[1]s`
)

const nameSpaceMsg = "Cannot determine target openshift namespace"
const clientMsg = "Unable to create an openshift client"

const typeLabel = "oshinko-type"
const clusterLabel = "oshinko-cluster"

const workerType = "worker"
const masterType = "master"
const webuiType = "webui"

const masterPortName = "spark-master"
const webPortName = "spark-webui"

// NewCmdClusters implements the OpenShift cli rollback command
func NewCmdClusters(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &ClusterOptions{}

	cmd := &cobra.Command{
		Use:     "clusters",
		Short:   "Display existing clusters",
		Long:    clustersLong,
		Example: fmt.Sprintf(clustersExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			options.PathOptions = cliconfig.NewPathOptions(cmd)

			if err := options.Complete(f, args, out); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			if err := options.RunClusters(); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}

	cmd.Flags().BoolVarP(&options.DisplayShort, "short", "q", false, "If true, display only the cluster names")
	return cmd
}

func (o *ClusterOptions) Complete(f *clientcmd.Factory, args []string, out io.Writer) error {
	if len(args) > 0 {
		return fmt.Errorf("no arguments should be passed")
	}

	var err error
	o.Config, err = f.OpenShiftClientConfig.RawConfig()
	if err != nil {
		return err
	}

	o.ClientConfig, err = f.OpenShiftClientConfig.ClientConfig()
	if err != nil {
		return err
	}

	o.Client, o.KClient, err = f.Clients()
	if err != nil {
		return err
	}

	o.Out = out

	return nil
}

func makeSelector(otype string, clustername string) kapi.ListOptions {
	const typeLabel = "oshinko-type"
	const clusterLabel = "oshinko-cluster"
	// Build a selector list based on type and/or cluster name
	ls := labels.NewSelector()
	if otype != "" {
		ot, _ := labels.NewRequirement(typeLabel, labels.EqualsOperator, sets.NewString(otype))
		ls = ls.Add(*ot)
	}
	if clustername != "" {
		cname, _ := labels.NewRequirement(clusterLabel, labels.EqualsOperator, sets.NewString(clustername))
		ls = ls.Add(*cname)
	}
	return kapi.ListOptions{LabelSelector: ls}
}

func tostrptr(val string) *string {
	v := val
	return &v
}
func toint64ptr(val int64) *int64 {
	v := val
	return &v
}

func sparkMasterURL(name string, port *kapi.ServicePort) string {
	return "spark://" + name + ":"
	//+ strconv.Itoa(port.Port)
}

func countWorkers(client kclient.PodInterface, clustername string) (int64, *kapi.PodList, error) {
	// If we are  unable to retrieve a list of worker pods, return -1 for count
	// This is an error case, differnt from a list of length 0. Let the caller
	// decide whether to report the error or the -1 count
	cnt := int64(-1)
	selectorlist := makeSelector(workerType, clustername)
	pods, err := client.List(selectorlist)
	if pods != nil {
		cnt = int64(len(pods.Items))
	}
	return cnt, pods, err
}

func retrieveMasterURL(client kclient.ServiceInterface, clustername string) string {
	selectorlist := makeSelector(masterType, clustername)
	srvs, err := client.List(selectorlist)
	if err == nil && len(srvs.Items) != 0 {
		srv := srvs.Items[0]
		return sparkMasterURL(srv.Name, &srv.Spec.Ports[0])
	}
	return ""
}
func getClusters(kClient *kclient.Client, namespace string) ([]*clusters.ClustersItems0, error) {
	//fmt.Println("-------")
	//fmt.Println(namespace)
	pc := kClient.Pods(namespace)
	//fmt.Println(pc)
	sc := kClient.Services(namespace)
	//fmt.Println(sc)

	payload := clusters.FindClustersOKBodyBody{}
	payload.Clusters = []*clusters.ClustersItems0{}
	// Create a map so that we can track clusters by name while we
	// find out information about them
	clist := map[string]*clusters.ClustersItems0{}

	// Get all of the master pods
	pods, err := pc.List(makeSelector(masterType, ""))
	if err != nil {
		//return reterr(fail(err, mastermsg, 500))
	}

	for i := range pods.Items {

		// Build the cluster record if we don't already have it
		// (theoretically with HA we might have more than 1 master)
		clustername := pods.Items[i].Labels[clusterLabel]
		if citem, ok := clist[clustername]; !ok {
			clist[clustername] = new(clusters.ClustersItems0)
			citem = clist[clustername]
			citem.Name = tostrptr(clustername)
			//fmt.Println(clustername)
			citem.Href = tostrptr("/clusters/" + clustername)

			// Note, we do not report an error here since we are
			// reporting on multiple clusters. Instead cnt will be -1.
			cnt, _, _ := countWorkers(pc, clustername)
			//fmt.Println(cnt)
			// TODO we only want to count running pods (not terminating)
			citem.WorkerCount = toint64ptr(cnt)
			// TODO make something real for status
			citem.Status = tostrptr("Running")
			citem.MasterURL = tostrptr(retrieveMasterURL(sc, clustername))
			payload.Clusters = append(payload.Clusters, citem)
		}
	}
	//projects, err := oClient.Projects().List(kapi.ListOptions{})
	//if err != nil {
	//	return nil, err
	//}
	return payload.Clusters, nil
}

// RunProjects lists all projects a user belongs to
func (o ClusterOptions) RunClusters() error {
	_ = "breakpoint"
	config := o.Config
	clientCfg := o.ClientConfig
	out := o.Out

	currentContext := config.Contexts[config.CurrentContext]
	currentProject := currentContext.Namespace

	var currentProjectExists bool
	var currentProjectErr error

	kclient := o.KClient
	oclient := o.Client

	if len(currentProject) > 0 {
		if _, currentProjectErr := oclient.Projects().Get(currentProject); currentProjectErr == nil {
			currentProjectExists = true
		}
	}

	defaultContextName := cliconfig.GetContextNickname(currentContext.Namespace, currentContext.Cluster, currentContext.AuthInfo)

	var msg string
	clusters, err := getClusters(kclient, currentProject)
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
				//if len(displayName) == 0 {
				//	displayName = cluster.Annotations["displayName"]
				//}

				//if currentProjectExists && !o.DisplayShort {
				//	asterisk = "    "
				//	if currentProject == *(cluster.Name) {
				//		asterisk = "  * "
				//	}
				//}
				//if len(displayName) > 0 && displayName != cluster.Name && !o.DisplayShort {
				//	msg += fmt.Sprintf("\n"+asterisk+"%s - %s", cluster.Name, displayName)
				//} else {
				//	if o.DisplayShort && count == 1 {
				//		linebreak = ""
				//	}
				//	msg += fmt.Sprintf(linebreak+asterisk+"%s", cluster.Name)
				//}
				msg += fmt.Sprintf(linebreak+asterisk+"%s \t  %d", displayName, workCount)
			}
		}
		//switch len(clusters) {
		//case 0:
		//
		//case 1:
		//	if o.DisplayShort {
		//		//msg += fmt.Sprintf("%s", &clusters[0])
		//	} else {
		//		//msg += fmt.Sprintf("You have one project on this server: %q.", api.DisplayNameAndNameForProject(&clusters[0]))
		//	}
		//default:
		//	asterisk := ""
		//	count := 0
		//	if !o.DisplayShort {
		//		msg += "You have access to the following projects and can switch between them with 'oc project <projectname>':\n"
		//	}
		//
		//	sort.Sort(SortByClusterName(clusters))
		//	for _, cluster := range clusters {
		//		count = count + 1
		//		displayName := *(cluster.Name)
		//		linebreak := "\n"
		//		//if len(displayName) == 0 {
		//		//	displayName = cluster.Annotations["displayName"]
		//		//}
		//
		//		//if currentProjectExists && !o.DisplayShort {
		//		//	asterisk = "    "
		//		//	if currentProject == *(cluster.Name) {
		//		//		asterisk = "  * "
		//		//	}
		//		//}
		//		//if len(displayName) > 0 && displayName != cluster.Name && !o.DisplayShort {
		//		//	msg += fmt.Sprintf("\n"+asterisk+"%s - %s", cluster.Name, displayName)
		//		//} else {
		//		//	if o.DisplayShort && count == 1 {
		//		//		linebreak = ""
		//		//	}
		//		//	msg += fmt.Sprintf(linebreak+asterisk+"%s", cluster.Name)
		//		//}
		//		msg += fmt.Sprintf(linebreak+asterisk+"%s", displayName)
		//	}
		//}
		fmt.Println(msg)

		if len(clusters) > 0 && !o.DisplayShort {
			if !currentProjectExists {
				if clientcmd.IsForbidden(currentProjectErr) {
					fmt.Printf("you do not have rights to view project %q. Please switch to an existing one.", currentProject)
				}
				return currentProjectErr
			}

			// if they specified a project name and got a generated context, then only show the information they care about.  They won't recognize
			// a context name they didn't choose
			if config.CurrentContext == defaultContextName {
				fmt.Fprintf(out, "\nUsing project %q on server %q.\n", currentProject, clientCfg.Host)
			} else {
				fmt.Fprintf(out, "\nUsing project %q from context named %q on server %q.\n", currentProject, config.CurrentContext, clientCfg.Host)
			}
		}
		return nil
	}

	return err
}
