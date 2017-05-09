package cmd

import (
	"fmt"
	"io"

	"github.com/renstrom/dedent"
	"github.com/spf13/cobra"

	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/core/clusters"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/auth"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
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
	authOptions := &auth.AuthOptions{
		Reader: reader,
		Out:    out,
	}
	options := &CmdOptions{
		AuthOptions: *authOptions,
	}

	cmd := &cobra.Command{
		Use: "create <NAME> ",
		//--masters <MASTER> --workers <WORKERS> --image <IMAGE> --sparkmasterconfig <DIR>
		Short: "Create new spark clusters",
		Long:  "Create spark cluster.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
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
	cmd.Flags().String("image", "", "spark image to be used. Default image is radanalyticsio/openshift-spark.")
	cmd.Flags().Bool("exposeui", true, "True will expose the Spark WebUI via a route")
	if extended {
		cmd.Flags().String("app", "", "Treat the cluster as ephemeral and tied to an app (name of pod or deployment)")
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
	_, err := clusters.CreateCluster(o.Name, o.Project, defaultImage, &config, o.Client, o.KClient, o.App)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "cluster \"%s\" created \n", o.Name)
	return nil
}
