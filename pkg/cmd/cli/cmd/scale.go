package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/pkg/auth"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/common"
	"github.com/radanalyticsio/oshinko-core/clusters"
	"github.com/radanalyticsio/oshinko-core/clusterconfigs"
)

func NewCmdScale(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdScale(f, in, out)
	return cmd
}

func CmdScale(f *clientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	authOptions := &auth.AuthOptions{
		Reader: reader,
		Out:    out,
	}
	options := &CmdOptions{
		CommonOptions: common.CommonOptions{AuthOptions: *authOptions},
	}

	cmd := &cobra.Command{
		Use:   ScaleCmdUsage,
		Short: ScaleCmdShort,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}
			if err := options.RunScale(); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}
	cmd.Flags().Int("masters", -1, "Numbers of workers in spark cluster")
	cmd.Flags().Int("workers", -1, "Numbers of workers in spark cluster")
	cmd.MarkFlagRequired("workers")
	return cmd
}

func (o *CmdOptions) RunScale() error {
	config := clusterconfigs.ClusterConfig{}
	config.WorkerCount = o.WorkerCount
	_, err := clusters.UpdateCluster(o.Name, o.Project, &config, o.Client, o.KClient)
	if err != nil {
		return err
	}
	fmt.Fprintf(o.Out, "cluster \"%s\" scaled \n", o.Name)
	return nil
}
