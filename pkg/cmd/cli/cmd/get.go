package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"

	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/radanalyticsio/oshinko-rest/restapi/operations/clusters"
	kapierrors "k8s.io/kubernetes/pkg/api/errors"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"
	"sort"
)

func getClusters(kClient *kclient.Client, namespace string) ([]*clusters.ClustersItems0, error) {

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
		return nil, err
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

	return payload.Clusters, nil
}

// RunProjects lists all projects a user belongs to
func (o *AuthOptions) RunClusters(currentProject string) error {
	_ = "breakpoint"

	kubeclient := o.KClient

	var msg string
	clusters, err := getClusters(kubeclient, currentProject)
	if err == nil {
		clusterCount := len(clusters)
		if clusterCount <= 0 {
			msg += "There are no clusters in any projects. You can create a cluster with the 'create' command."
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

// RunLogin contains all the necessary functionality for the OpenShift cli login command
func RunGetCmd(cmd *cobra.Command, options *AuthOptions) error {
	if err := options.GatherInfo(); err != nil {
		return err
	}

	if err := options.RunClusters(options.Project); err != nil {
		return err
	}

	return nil
}

func NewCmdGet(fullName string, f *osclientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	options := &AuthOptions{
		Reader: reader,
		Out:    out,
	}

	cmds := &cobra.Command{
		Use:   "get <NAME>",
		Short: "Get running spark clusters",
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(err)
			}

			err := RunGetCmd(cmd, options)

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
