package cmd

import (
	"fmt"
	"io"

	"github.com/renstrom/dedent"
	"github.com/spf13/cobra"

	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/radanalyticsio/oshinko-core/clusters"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/auth"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"github.com/radanalyticsio/oshinko-core/clusterconfigs"
)

var (
	sparkClusterLong = dedent.Dedent(`
		Create a spark cluster with the specified name.`)

	sparkClusterExample = dedent.Dedent(`
		  # Create a new spark cluster named my-spark-cluster
		  $ oshinko create cluster my-spark-cluster`)
)

func NewCmdCreate(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdCreate(f, in, out)
	return cmd
}

func CmdCreate(f *clientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
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

	cmd.Flags().Int("masters", 0, "Numbers of workers in spark cluster")
	cmd.Flags().Int("workers", 0, "Numbers of workers in spark cluster")
	cmd.Flags().String("masterconfigdir", defaultsparkconfdir, "Config folder for spark master")
	cmd.Flags().String("workerconfigdir", defaultsparkconfdir, "Config folder for spark worker")
	cmd.Flags().String("masterconfig", "", "ConfigMap name for spark master")
	cmd.Flags().String("workerconfig", "", "ConfigMap name for spark worker")
	cmd.Flags().String("storedconfig", "", "ConfigMap name for spark cluster")
	cmd.Flags().String("image", defaultImage, "spark image to be used.Default value is radanalyticsio/openshift-spark.")
	//cmd.MarkFlagRequired("workers")
	return cmd
}

func (o *CmdOptions) RunCreate(out io.Writer, cmd *cobra.Command, args []string) error {
	config := clusterconfigs.ClusterConfig{}
	config.WorkerCount = o.WorkerCount
	config.MasterCount = o.MasterCount
	config.SparkWorkerConfig = o.WorkerConfig
	config.SparkMasterConfig = o.MasterConfig
        config.Name = o.StoredConfig
	_, err := clusters.CreateCluster(o.Name, o.Project, o.Image, &config, o.Client, o.KClient)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "cluster \"%s\" created \n", o.Name)
	return nil
}
