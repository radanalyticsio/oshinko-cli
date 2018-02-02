package cmd

import (
	"fmt"
	"io"

	"github.com/renstrom/dedent"
	"github.com/spf13/cobra"

	//"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/oc/cli/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/core/clusters"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/auth"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"github.com/radanalyticsio/oshinko-cli/version"
)

var (
	sparkClusterLong = dedent.Dedent(`
		Create a spark cluster with the specified name.`)

	sparkClusterExample = dedent.Dedent(`
		  # Create a new spark cluster named my-spark-cluster
		  $ oshinko create cluster my-spark-cluster`)
)

func NewCmdCreate(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdCreate(f, in, out, false)
	return cmd
}

func NewCmdCreateExtended(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdCreate(f, in, out, true)
	return cmd
}

func CmdCreate(f *clientcmd.Factory, reader io.Reader, out io.Writer, extended bool) *cobra.Command {
	var cmdString string
	authOptions := &auth.AuthOptions{
		Reader: reader,
		Out:    out,
	}
	options := &CmdOptions{
		AuthOptions: *authOptions,
	}


	if extended {
		cmdString = "create_eph"

	} else {
		cmdString = "create"
	}

	cmd := &cobra.Command{
		Use: cmdString + " <NAME> ",
		//--masters <MASTER> --workers <WORKERS> --image <IMAGE> --sparkmasterconfig <DIR>
		Short: "Create new spark cluster",
		Hidden: extended,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageErrorf(cmd, err.Error()))
			}
			if err := options.RunCreate(out, cmd, args); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}

	cmd.Flags().Int("masters", clusters.SentinelCountValue, fmt.Sprintf("Number of masters in the spark cluster (%d"+
		" means take masters from storedconfig and/or defaults)", clusters.SentinelCountValue))
	cmd.Flags().Int("workers", clusters.SentinelCountValue, fmt.Sprintf("Number of workers in the spark cluster (%d"+
		" means take workers from storedconfig and/or defaults)", clusters.SentinelCountValue))
	cmd.Flags().String("masterconfig", "", "ConfigMap name for spark master")
	cmd.Flags().String("workerconfig", "", "ConfigMap name for spark worker")
	cmd.Flags().String("storedconfig", "", "ConfigMap name for spark cluster")
	cmd.Flags().String("image", "", "Spark image to be used. Default image is " + version.GetSparkImage() + ".")
	cmd.Flags().String("exposeui", "", "True or False, expose the Spark WebUI via a route (default True)")
	cmd.Flags().String("metrics", "", "Enable spark metrics (default false). Set the value to 'prometheus' for prometheus metrics and 'true' or 'jolokia' for jolokia metrics (deprecated).")
	if extended {
		cmd.Flags().BoolP("ephemeral", "e", false, "Treat the cluster as ephemeral. The 'app' flag must also be set.")
		cmd.Flags().String("app", "", "Associate the cluster with an app.  Value may be the name of a pod or deployment (but not a deploymentconfig)")
	}
	return cmd
}

func (o *CmdOptions) RunCreate(out io.Writer, cmd *cobra.Command, args []string) error {
	config := clusters.ClusterConfig{}
	config.WorkerCount = o.WorkerCount
	config.MasterCount = o.MasterCount
	config.SparkWorkerConfig = o.WorkerConfig
	config.SparkMasterConfig = o.MasterConfig
	config.SparkImage = o.Image
	config.Name = o.StoredConfig
	config.ExposeWebUI = o.ExposeWebUI
	config.Metrics = o.Metrics
	result, err := clusters.CreateCluster(o.Name, o.Project, version.GetSparkImage(), &config, o.Config, o.App, o.Ephemeral)
	if err != nil {
		return err
	}
	if result.Ephemeral == "<shared>" {
		fmt.Fprintf(out, "shared cluster \"%s\" created \n", o.Name)
	} else {
		fmt.Fprintf(out, "ephemeral cluster \"%s\" created \n", o.Name)
	}
	return nil
}
