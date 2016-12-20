package cmd

import (
	"fmt"
	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/auth"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

type CmdOptions struct {
	Name            string `json:"name,omitempty"`
	Image           string `json:"image"`
	WorkerCount     int    `json:"workerCount"`
	MasterCount     int    `json:"masterCount,omitempty"`
	MasterConfig    string `json:"sparkMasterConfig,omitempty"`
	MasterConfigDir string `json:"masterConfigDir,omitempty"`
	WorkerConfig    string `json:"workerConfig,omitempty"`
	WorkerConfigDir string `json:"workerConfigDir,omitempty"`
	Verbose         bool
	Output          string

	auth.AuthOptions
}

func (o *CmdOptions) Complete(f *osclientcmd.Factory, cmd *cobra.Command, args []string) error {
	o.Image = defaultImage
	o.WorkerCount = 1
	o.MasterCount = 1
	o.MasterConfigDir = defaultsparkconfdir
	o.WorkerConfigDir = defaultsparkconfdir
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
	if cmd.Flags().Lookup("workers") != nil {
		o.WorkerCount = kcmdutil.GetFlagInt(cmd, "workers")
	}
	if cmd.Flags().Lookup("masters") != nil {
		o.MasterCount = kcmdutil.GetFlagInt(cmd, "masters")
	}
	if cmd.Flags().Lookup("image") != nil {
		o.Image = kcmdutil.GetFlagString(cmd, "image")
	}
	if cmd.Flags().Lookup("masterconfigdir") != nil {
		o.MasterConfigDir = kcmdutil.GetFlagString(cmd, "masterconfigdir")
	}
	if cmd.Flags().Lookup("workerconfigdir") != nil {
		o.WorkerConfigDir = kcmdutil.GetFlagString(cmd, "workerconfigdir")
	}
	if cmd.Flags().Lookup("masterconfig") != nil && kcmdutil.GetFlagString(cmd, "masterconfig") != "" {
		o.MasterConfig = kcmdutil.GetFlagString(cmd, "masterconfig")

		err = checkForConfigMap(o.MasterConfig, o.KClient.ConfigMaps(o.Project))
		if err != nil {
			return err
		}
	}
	if cmd.Flags().Lookup("workerconfig") != nil && kcmdutil.GetFlagString(cmd, "workerconfig") != "" {
		o.WorkerConfig = kcmdutil.GetFlagString(cmd, "workerconfig")
		err = checkForConfigMap(o.WorkerConfig, o.KClient.ConfigMaps(o.Project))
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *CmdOptions) GatherInfo() error {
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
