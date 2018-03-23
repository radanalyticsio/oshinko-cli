package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	//"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/oc/cli/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/core/clusters"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/auth"
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
		AuthOptions: *authOptions,
	}

	cmd := &cobra.Command{
		Use:   ScaleCmdUsage,
		Short: ScaleCmdShort,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageErrorf(cmd, err.Error()))
			}
			if err := options.RunScale(); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}
	cmd.Flags().Int("masters", clusters.SentinelCountValue, fmt.Sprintf("Number of masters in the spark cluster (%d"+
		" means leave masters unchanged)", clusters.SentinelCountValue))
	cmd.Flags().Int("workers", clusters.SentinelCountValue, fmt.Sprintf("Number of workers in the spark cluster (%d"+
		" means leave workers unchanged)", clusters.SentinelCountValue))
	cmd.MarkFlagRequired("workers")
	return cmd
}

func (o *CmdOptions) RunScale() error {
	if o.MasterCount <= clusters.SentinelCountValue && o.WorkerCount <= clusters.SentinelCountValue {
		fmt.Fprintf(o.Out, "neither masters nor workers specified, cluster \"%s\" not scaled \n", o.Name)
	} else {
		err := clusters.ScaleCluster(o.Name, o.Project, o.MasterCount, o.WorkerCount, o.Config)
		if err != nil {
			return err
		}
		fmt.Fprintf(o.Out, "cluster \"%s\" scaled \n", o.Name)
	}
	return nil
}
