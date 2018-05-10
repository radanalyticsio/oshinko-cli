package cmd

import (
	"fmt"
	"io"

	osclientcmd "github.com/openshift/origin/pkg/oc/cli/util/clientcmd"
	//kclientcmd "k8s.io/client-go/tools/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/core/clusters"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/auth"
	"github.com/spf13/cobra"

	"os"
	"sort"

	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
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
		c, err := clusters.FindSingleCluster(o.Name, o.Project, o.Config)
		if err != nil {
			return err
		}
		clist = []clusters.SparkCluster{c}
	} else {
		clist, err = clusters.FindClusters(o.Project, o.Config, o.App)
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
		for c, cluster := range tmpClusters {
			if o.Name == "" || cluster.Name == o.Name {
				if o.Output == "" {
					msg += fmt.Sprintf(linebreak+asterisk+"%-14s\t %d\t %-30s\t %-32s\t %-32s\t %s\t  %s", cluster.Name,
						cluster.WorkerCount, cluster.MasterURL, cluster.MasterWebURL, cluster.MasterWebRoute, cluster.Status, cluster.Ephemeral)
				} else if o.NoPods {
					tmpClusters[c].Pods = nil
				}
			}
		}
		if o.Output != "" {
			PrintOutput(o.Output, tmpClusters)
		}
	}
	fmt.Println(msg)
	return nil
}

func NewCmdGet(fullName string, f *osclientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdGet(f, in, out, false)
	return cmd
}

func NewCmdGetExtended(fullName string, f *osclientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdGet(f, in, out, true)
	return cmd
}

func CmdGet(f *osclientcmd.Factory, reader io.Reader, out io.Writer, extended bool) *cobra.Command {
	var cmdString string
	authOptions := &auth.AuthOptions{
		Reader: reader,
		Out:    out,
	}
	options := &CmdOptions{
		AuthOptions:    *authOptions,
		Verbose:        false,
		NoNameRequired: true,
	}

	if extended {
		cmdString = "get_eph"
	} else {
		cmdString = "get"
	}

	cmds := &cobra.Command{
		Use:    cmdString,
		Short:  "Get running spark clusters",
		Hidden: extended,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(err)
			}
			/*
				#	Config should work from this point below
			*/
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
	cmds.Flags().BoolVarP(&options.Verbose, "verbose", "v", options.Verbose, "Turn on verbose output\n\n")
	cmds.Flags().BoolP("nopods", "", false, "Do not include pod list for cluster in yaml or json output")
	if extended {
		cmds.Flags().String("app", "", "Get the clusters associated with the app. The value may be the name of a pod or deployment (but not a deploymentconfig). Ignored if a name is specified.")
	}
	return cmds
}
