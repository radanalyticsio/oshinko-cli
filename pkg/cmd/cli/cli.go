package cli

import (
	"fmt"
	"github.com/openshift/origin/pkg/cmd/flagtypes"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"github.com/openshift/origin/pkg/oc/cli/util/clientcmd"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/cmd"
	"github.com/spf13/cobra"
	"io"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"
	"runtime"
	"strings"
	oshinkocmd "github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/cmd"
)

const (
	productName = `Oshinko`
	cliLong     = productName + ` Client

This client helps you deploy, scale and run your spark applications on Openshift
`
	cliExplain = `
To see the full list of commands supported, run '%[1]s help'.
`
)

func NewCommandCLI(name, fullName string, in io.Reader, out, errout io.Writer) *cobra.Command {
	// Main command
	cmds := &cobra.Command{
		Use:   name,
		Short: "Command line tools for managing spark cluster",
		Long:  cliLong,
		Run: func(c *cobra.Command, args []string) {
			c.SetOutput(out)
			cmdutil.RequireNoArguments(c, args)
			fmt.Fprint(out, cliLong)
			fmt.Fprintf(out, cliExplain, fullName)
		},
		//TODO ; add bash completion
		//BashCompletionFunction: bashCompletionFunc,
	}

	f := clientcmd.New(cmds.PersistentFlags())

	first := oshinkocmd.NewCmdGet(fullName, f, in, out)
	groups:= ktemplates.CommandGroups{
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				first,
				oshinkocmd.NewCmdWhoAmI("whoami", fullName, f, out),
				oshinkocmd.NewCmdDelete(fullName, f, in, out),
				oshinkocmd.NewCmdCreate(fullName, f, in, out),
				oshinkocmd.NewCmdScale(fullName, f, in, out),
			},
		},
	}
	groups.Add(cmds)
	changeSharedFlagDefaults(cmds)

	filters := []string{
		"options",
	}

	cmds.AddCommand(cmd.NewCmdVersion(fullName, f, in, out))
	cmds.AddCommand(NewCmdOptions(out))

	// Add hidden commands
	tmp := oshinkocmd.NewCmdConfigMap(fullName, f, in, out)
	cmds.AddCommand(tmp)

	// If we move the expose to one of the hidden commands,
	// then the printing of the help is still messed up for the
	// exposed flags but at least it's not on one of the main commands
	ktemplates.ActsAsRootCommand(cmds, filters, groups...).
		ExposeFlags(tmp, "server", "client-certificate",
		"client-key", "certificate-authority", "insecure-skip-tls-verify", "token")

	cmds.AddCommand(oshinkocmd.NewCmdCreateExtended(fullName, f, in, out))
	cmds.AddCommand(oshinkocmd.NewCmdDeleteExtended(fullName, f, in, out))
	cmds.AddCommand(oshinkocmd.NewCmdGetExtended(fullName, f, in, out))
	return cmds
}

//TODO ensure we can limit the number of options
func NewCmdOptions(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use: "options",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	ktemplates.UseOptionsTemplates(cmd)

	return cmd
}

// changeSharedFlagDefaults changes values of shared flags that we disagree with.  This can't be done in godep code because
// that would change behavior in our `kubectl` symlink. Defend each change.
// 1. show-all - the most interesting pods are terminated/failed pods.  We don't want to exclude them from printing
func changeSharedFlagDefaults(rootCmd *cobra.Command) {
	cmds := []*cobra.Command{rootCmd}

	for i := 0; i < len(cmds); i++ {
		currCmd := cmds[i]
		cmds = append(cmds, currCmd.Commands()...)
	}
}

// CommandFor returns the appropriate command for this base name,
// or the OpenShift CLI command.
func CommandFor(basename string) *cobra.Command {
	var cmd *cobra.Command

	in, out, errout := os.Stdin, os.Stdout, os.Stderr

	// Make case-insensitive and strip executable suffix if present
	if runtime.GOOS == "windows" {
		basename = strings.ToLower(basename)
		basename = strings.TrimSuffix(basename, ".exe")
	}

	cmd = NewCommandCLI(basename, basename, in, out, errout)

	if cmd.UsageFunc() == nil {
		ktemplates.ActsAsRootCommand(cmd, []string{"options"})
	}
	flagtypes.GLog(cmd.PersistentFlags())

	return cmd
}
