package cli

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	//"github.com/golang/glog"
	"github.com/spf13/cobra"
	//"github.com/spf13/pflag"

	//kubecmd "k8s.io/kubernetes/pkg/kubectl/cmd"

	oshinkocmd "github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli/cmd"

	"github.com/openshift/origin/pkg/cmd/cli/cmd"
	//"github.com/openshift/origin/pkg/cmd/cli/cmd/cluster"
	"github.com/openshift/origin/pkg/cmd/cli/cmd/set"
	"github.com/openshift/origin/pkg/cmd/flagtypes"
	"github.com/openshift/origin/pkg/cmd/templates"
	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	//"github.com/openshift/origin/pkg/version"
)

const productName = `Oshinko`

const cliLong = productName + ` Client

The Oshinko client.
`

const cliExplain = `
To create a new application, login to your server and then run :
 TODO
To see the full list of commands supported, run '%[1]s help'.
`

func NewCommandCLI(name, fullName string, in io.Reader, out, errout io.Writer) *cobra.Command {
	// Main command
	cmds := &cobra.Command{
		Use:   name,
		Short: "Command line tools for managing applications",
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

	//loginCmd := cmd.NewCmdLogin(fullName, f, in, out)
	loginCmd := oshinkocmd.NewCmdLogin(fullName, f, in, out)
	groups := templates.CommandGroups{
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				//oshinkocmd.NewCmdTypes(fullName, f, out),
				loginCmd,
			},
		},
	}
	groups.Add(cmds)

	filters := []string{
		"options",
		// These commands are deprecated and should not appear in help
		moved(fullName, "set env", cmds, set.NewCmdEnv(fullName, f, in, out)),
		moved(fullName, "set volume", cmds, set.NewCmdVolume(fullName, f, out, errout)),
		moved(fullName, "logs", cmds, cmd.NewCmdBuildLogs(fullName, f, out)),
	}

	changeSharedFlagDefaults(cmds)
	templates.ActsAsRootCommand(cmds, filters, groups...).
		ExposeFlags(loginCmd, "token")

	return cmds
}

func moved(fullName, to string, parent, cmd *cobra.Command) string {
	cmd.Long = fmt.Sprintf("DEPRECATED: This command has been moved to \"%s %s\"", fullName, to)
	cmd.Short = fmt.Sprintf("DEPRECATED: %s", to)
	parent.AddCommand(cmd)
	return cmd.Name()
}

// changeSharedFlagDefaults changes values of shared flags that we disagree with.  This can't be done in godep code because
// that would change behavior in our `kubectl` symlink. Defend each change.
// 1. show-all - the most interesting pods are terminated/failed pods.  We don't want to exclude them from printing
func changeSharedFlagDefaults(rootCmd *cobra.Command) {
	cmds := []*cobra.Command{rootCmd}

	for i := 0; i < len(cmds); i++ {
		currCmd := cmds[i]
		cmds = append(cmds, currCmd.Commands()...)

		if showAllFlag := currCmd.Flags().Lookup("show-all"); showAllFlag != nil {
			showAllFlag.DefValue = "true"
			showAllFlag.Value.Set("true")
			showAllFlag.Changed = false
			showAllFlag.Usage = "When printing, show all resources (false means hide terminated pods.)"
		}

		// we want to disable the --validate flag by default when we're running kube commands from oc.  We want to make sure
		// that we're only getting the upstream --validate flags, so check both the flag and the usage
		if validateFlag := currCmd.Flags().Lookup("validate"); (validateFlag != nil) && (validateFlag.Usage == "If true, use a schema to validate the input before sending it") {
			validateFlag.DefValue = "false"
			validateFlag.Value.Set("false")
			validateFlag.Changed = false
		}
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
	//switch basename {
	//case "kubectl":
	//	cmd = NewCmdKubectl(basename, out)
	//default:
	//	cmd = NewCommandCLI(basename, basename, in, out, errout)
	//}

	if cmd.UsageFunc() == nil {
		templates.ActsAsRootCommand(cmd, []string{"options"})
	}
	flagtypes.GLog(cmd.PersistentFlags())

	return cmd
}
