package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"

	//"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/oc/cli/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/version"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func NewCmdVersion(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdVersion(f, in, out)
	return cmd
}

func CmdVersion(f *clientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {

	options := &CmdOptions{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Return the version of the CLI.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.RunVersion(out, cmd, args); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}

	return cmd
}

func (o *CmdOptions) RunVersion(out io.Writer, cmd *cobra.Command, args []string) error {

	fmt.Fprintf(out, "%s %s\nDefault spark image: %s\n", version.GetAppName(), version.GetVersion(), version.GetSparkImage())
	return nil
}
