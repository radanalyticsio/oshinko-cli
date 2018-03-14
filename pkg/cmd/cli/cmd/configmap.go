package cmd

import (
	"github.com/spf13/cobra"
	"io"

	//"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/oc/cli/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/auth"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"fmt"
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
				kcmdutil.CheckErr(kcmdutil.UsageErrorf(cmd, err.Error()))
			}
			if err := options.RunCmdConfigMap(f, out, cmd, args); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}
	cmd.Flags().StringP("output", "o", "", "Output format if set. One of: json|yaml")
	cmd.Flags().String("directory", "", "Directory in which to write files representing key / value pairs.")
	cmd.Flags().BoolVarP(&options.Verbose, "verbose", "v", options.Verbose, "Turn on verbose output\n\n")
	return cmd
}

func (o *CmdOptions) RunCmdConfigMap(f *clientcmd.Factory, out io.Writer, cmd *cobra.Command, args []string) error {
	kubeClient, err := f.ClientSet()
	cmap, err := kubeClient.Core().ConfigMaps(o.Project).Get(o.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if cmap != nil && o.Directory != "" {
		_, err := os.Stat(o.Directory)
		if err != nil {
			return err
		}
		for k, v := range cmap.Data {
			file, err := os.Create(filepath.Join(o.Directory, k))
			if err == nil {
				if o.Verbose {
					fmt.Printf("Writing %s\n", filepath.Join(o.Directory, k))
				}
				file.WriteString(v)
			} else {
				return err
			}
		}
	}
	if o.Output != "" || (o.Output == "" && o.Directory == "") {
		if o.Output == "" {
			o.Output = "json"
		}
		PrintOutput(o.Output, cmap)
	}
	return nil
}
