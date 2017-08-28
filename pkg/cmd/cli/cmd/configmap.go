package cmd

import (
	"github.com/spf13/cobra"
	"io"

	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/auth"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func NewCmdConfigMap(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdConfigMap(f, in, out)
	return cmd
}

func CmdConfigMap(f *clientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	authOptions := &auth.AuthOptions{
		Reader: reader,
		Out:    out,
	}
	options := &CmdOptions{
		AuthOptions: *authOptions,
	}

	cmd := &cobra.Command{
		Use:   "configmap <NAME> ",
		Short: "Return a configmap in json",
		Long:  "Lookup a configmap by name and print as json if it exists",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Complete(f, cmd, args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}
			if err := options.RunCmdConfigMap(out, cmd, args); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}
	cmd.Flags().StringP("output", "o", "json", "Output format. One of: json|yaml")
	return cmd
}

func (o *CmdOptions) RunCmdConfigMap(out io.Writer, cmd *cobra.Command, args []string) error {

	cmap, err := o.KClient.ConfigMaps(o.Project).Get(o.Name)
	if err != nil {
		return err
	}
	PrintOutput(o.Output, cmap)
	return nil
}
