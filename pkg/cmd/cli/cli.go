package cli

import (
	"fmt"
	"github.com/openshift/origin/pkg/cmd/cli/cmd"
	"github.com/openshift/origin/pkg/cmd/flagtypes"
	"github.com/openshift/origin/pkg/cmd/templates"
	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/cobra"
	"io"
	"os"
	"runtime"
	"strings"

	oshinkocmd "github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/cmd"
)

const productName = `Oshinko`

const cliLong = productName + ` Client

The Oshinko client.
`

const cliExplain = `
To see the full list of commands supported, run '%[1]s help'.
`

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
		//TODO ; do this later
		//BashCompletionFunction: bashCompletionFunc,
	}

	f := clientcmd.New(cmds.PersistentFlags())

	getCmd := oshinkocmd.NewCmdGet(fullName, f, in, out)
	groups := templates.CommandGroups{
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				getCmd,
				oshinkocmd.NewCmdDelete(fullName, f, in, out),
				oshinkocmd.NewCmdCreate(fullName, f, in, out),
				oshinkocmd.NewCmdScale(fullName, f, in, out),
			},
		},
	}
	groups.Add(cmds)

	filters := []string{
		"options",
	}

	templates.ActsAsRootCommand(cmds, filters, groups...).
		ExposeFlags(getCmd, "server", "client-certificate",
			"client-key", "certificate-authority", "insecure-skip-tls-verify", "token")

	cmds.AddCommand(cmd.NewCmdOptions(out))
	return cmds
}

func moved(fullName, to string, parent, cmd *cobra.Command) string {
	cmd.Long = fmt.Sprintf("DEPRECATED: This command has been moved to \"%s %s\"", fullName, to)
	cmd.Short = fmt.Sprintf("DEPRECATED: %s", to)
	parent.AddCommand(cmd)
	return cmd.Name()
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
		templates.ActsAsRootCommand(cmd, []string{"options"})
	}
	flagtypes.GLog(cmd.PersistentFlags())

	return cmd
}
