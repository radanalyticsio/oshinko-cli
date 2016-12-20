package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"

	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/auth"
	kapierrors "k8s.io/kubernetes/pkg/api/errors"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"
	"sort"
)

func getClusters(o *CmdOptions) ([]SparkCluster, error) {

	pc := o.KClient.Pods(o.Project)
	//sc := o.KClient.Services(o.Project)

	clusters := []SparkCluster{}
	// Create a map so that we can track clusters by name while we
	// find out information about them
	clist := map[string]SparkCluster{}

	// Get all of the master pods
	pods, err := pc.List(makeSelector(masterType, ""))
	if err != nil {
		return nil, err
	}

	for i := range pods.Items {

		// Build the cluster record if we don't already have it
		// (theoretically with HA we might have more than 1 master)
		clustername := pods.Items[i].Labels[clusterLabel]
		if cluster, ok := clist[clustername]; !ok {
			//For each master
			clist[clustername] = SparkCluster{Namespace: o.Project,
				Name: clustername}
			cluster = clist[clustername]
			cluster.Href = "/clusters/" + clustername

			// Note, we do not report an error here since we are
			// reporting on multiple clusters. Instead cnt will be -1.
			cnt, _ := cluster.countWorkers(o.KClient)
			//fmt.Println(cnt)
			// TODO we only want to count running pods (not terminating)
			cluster.WorkerCount = cnt
			// TODO make something real for status
			cluster.Status = "Running"
			cluster.MasterURL = cluster.retrieveServiceURL(o.KClient, masterType)
			cluster.MasterWebURL = cluster.retrieveServiceURL(o.KClient, webuiType)
			clusters = append(clusters, cluster)
		}
	}

	return clusters, nil
}

// RunProjects lists all projects a user belongs to
func (o *CmdOptions) RunClusters() error {
	_ = "breakpoint"

	var msg string
	clusterExist := false
	linebreak := "\n"
	asterisk := ""
	clusters, err := getClusters(o)
	var namedCluster SparkCluster
	if err == nil {
		clusterCount := len(clusters)
		tmpClusters := clusters
		if clusterCount <= 0 {
			msg += "There are no clusters in any projects. You can create a cluster with the 'create' command."
		} else if clusterCount > 0 {
			sort.Sort(SortByClusterName(tmpClusters))
			for _, cluster := range tmpClusters {
				if o.Name == "" || cluster.Name == o.Name {
					clusterExist = true
					namedCluster = cluster
					if o.Output == "" {
						msg += fmt.Sprintf(linebreak+asterisk+"%s \t  %d\t  %s\t  %s\t  %s", cluster.Name,
							cluster.WorkerCount, cluster.MasterURL, cluster.MasterWebURL, cluster.Status)
					}
				}
			}
			if o.Output != "" {
				if o.Name == "" {
					PrintOutput(o.Output, clusters)
				} else if clusterExist {
					PrintOutput(o.Output, []SparkCluster{namedCluster})
				}
			}
		}
		if !clusterExist {
			msg += fmt.Sprintf(linebreak+asterisk+"There are no clusters with name %s", o.Name)
		}
		fmt.Println(msg)
		return nil
	}

	return err
}

func NewCmdGet(fullName string, f *osclientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	authOptions := &auth.AuthOptions{
		Reader: reader,
		Out:    out,
	}
	options := &CmdOptions{
		AuthOptions: *authOptions,
		Verbose:     false,
	}

	cmds := &cobra.Command{
		Use:   "get",
		Short: "Get running spark clusters",
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(err)
			}

			err := options.RunClusters()

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
	cmds.Flags().StringP("output", "o", "", "Output format. One of: json|yaml")
	return cmds
}
