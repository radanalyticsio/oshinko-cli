package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"

	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/radanalyticsio/oshinko-core/clusters"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/auth"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func NewCmdDelete(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdDelete(f, in, out)
	return cmd
}

func CmdDelete(f *clientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	authOptions := &auth.AuthOptions{
		Reader: reader,
		Out:    out,
	}
	options := &CmdOptions{
		AuthOptions: *authOptions,
	}

	cmd := &cobra.Command{
		Use:   "delete <NAME>",
		Short: "Delete spark cluster by name.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}
			if err := options.RunDelete(out, cmd, args); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}

	return cmd
}

func (o *CmdOptions) RunDelete(out io.Writer, cmd *cobra.Command, args []string) error {

	info, err := clusters.DeleteCluster(o.Name, o.Project, o.Client, o.KClient)
	if err != nil {
		return err
	}
	if info != "" {
		fmt.Println(info)
	}
	fmt.Fprintf(out, "cluster \"%s\" deleted \n", o.Name)
	return nil
}
