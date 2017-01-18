package common

import (
	"fmt"
	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/pkg/auth"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

)

type CommonOptions struct {
	Verbose         bool
	Output          string
	Name            string `json:"name,omitempty"`
	auth.AuthOptions
}

func (o *CommonOptions) Complete(f *osclientcmd.Factory, cmd *cobra.Command, args []string) error {
	currentCluster, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}
	o.Name = currentCluster
	if err := o.AuthOptions.Complete(f, cmd, args); err != nil {
		kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
	}
	if err := o.GatherInfo(); err != nil {
		return err
	}
	if cmd.Flags().Lookup("verbose") != nil {
		o.Verbose = kcmdutil.GetFlagBool(cmd, "verbose")
	}
	if cmd.Flags().Lookup("output") != nil {
		o.Output = kcmdutil.GetFlagString(cmd, "output")
		if o.Output != "yaml" || o.Output != "json" {
			cmdutil.UsageError(cmd, "INVALID output format only yaml|json allowed")
		}
	}
	return nil
}

func (o *CommonOptions) GatherInfo() error {
	var msg string
	var err error
	if msg, err = o.AuthOptions.GatherAuthInfo(); err != nil {
		return err
	}
	if o.Verbose {
		fmt.Printf(msg)
	}
	if msg, err = o.AuthOptions.GatherProjectInfo(); err != nil {
		return err
	}
	if o.Verbose {
		fmt.Printf(msg)
	}
	return nil
}
