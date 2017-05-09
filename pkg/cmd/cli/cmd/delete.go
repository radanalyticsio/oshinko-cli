package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"

	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/core/clusters"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/auth"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func NewCmdDelete(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdDelete(f, in, out, false)
	return cmd
}

func NewCmdDeleteExtended(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	cmd := CmdDelete(f, in, out, true)
	return cmd
}

func CmdDelete(f *clientcmd.Factory, reader io.Reader, out io.Writer, extended bool) *cobra.Command {
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
	if extended {
		cmd.Flags().String("app", "", "The app tied to an ephemeral cluster (name of pod or deployment)")
		cmd.Flags().String("app-status", "", "How the application has ended ('completed' or 'terminated')")
	}
	return cmd
}

func (o *CmdOptions) RunDelete(out io.Writer, cmd *cobra.Command, args []string) error {

	if (o.App == "" || o.AppStatus == "") && o.App + o.AppStatus != "" {
		return fmt.Errorf("Both --app and --appstatus must be set")
	}
	info, err := clusters.DeleteCluster(o.Name, o.Project, o.Client, o.KClient, o.App, o.AppStatus)
	if info != "" {
		fmt.Println(info)
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "cluster \"%s\" deleted \n", o.Name)
	return nil
}
