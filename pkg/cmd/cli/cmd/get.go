package cmd

import (
	"fmt"
	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/auth"
	"github.com/radanalyticsio/oshinko-cli/core/clusters"
	"github.com/spf13/cobra"
	"io"
	kapierrors "k8s.io/kubernetes/pkg/api/errors"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"
	"sort"
)

type SortByClusterName []clusters.SparkCluster

func (p SortByClusterName) Len() int {
	return len(p)
}

func (p SortByClusterName) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p SortByClusterName) Less(i, j int) bool {
	return p[i].Name < p[j].Name
}

// RunProjects lists all projects a user belongs to
func (o *CmdOptions) RunClusters() error {

	var msg string
	var clist []clusters.SparkCluster
	var err error

	linebreak := "\n"
	asterisk := ""

	if o.Name != "" {
		c, err := clusters.FindSingleCluster(o.Name, o.Project, o.Client, o.KClient)
		if err != nil {
			return err
		}
		clist = []clusters.SparkCluster{c}
	} else {
		clist, err = clusters.FindClusters(o.Project, o.KClient)
		if err != nil {
			return err
		}
	}

	clusterCount := len(clist)
	tmpClusters := clist
	if clusterCount <= 0 {
		msg += "There are no clusters in any projects. You can create a cluster with the 'create' command."
	} else if clusterCount > 0 {
		sort.Sort(SortByClusterName(tmpClusters))
		for _, cluster := range tmpClusters {
			if o.Name == "" || cluster.Name == o.Name {
				if o.Output == "" {
					msg += fmt.Sprintf(linebreak+asterisk+"%s \t  %d\t  %s\t  %s\t  %s", cluster.Name,
						cluster.WorkerCount, cluster.MasterURL, cluster.MasterWebURL, cluster.Status)
				}
			}
		}
		if o.Output != "" {
			PrintOutput(o.Output, clist)
		}
	}
	fmt.Println(msg)
	return nil
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
	cmds.Flags().BoolVarP(&options.Verbose, "verbose", "v", options.Verbose, "See details for resolving issues.")
	return cmds
}
