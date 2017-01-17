package common

import (
	"fmt"
	"github.com/openshift/origin/pkg/cmd/flagtypes"
	"github.com/openshift/origin/pkg/cmd/templates"
	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/cobra"
	"io"
	"os"
	"runtime"
	"strings"
)

func Moved(fullName, to string, parent, cmd *cobra.Command) string {
	cmd.Long = fmt.Sprintf("DEPRECATED: This command has been moved to \"%s %s\"", fullName, to)
	cmd.Short = fmt.Sprintf("DEPRECATED: %s", to)
	parent.AddCommand(cmd)
	return cmd.Name()
}

func NewCmdOptions(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use: "options",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	templates.UseOptionsTemplates(cmd)

	return cmd
}

// changeSharedFlagDefaults changes values of shared flags that we disagree with.  This can't be done in godep code because
// that would change behavior in our `kubectl` symlink. Defend each change.
// 1. show-all - the most interesting pods are terminated/failed pods.  We don't want to exclude them from printing
func ChangeSharedFlagDefaults(rootCmd *cobra.Command) {
	cmds := []*cobra.Command{rootCmd}

	for i := 0; i < len(cmds); i++ {
		currCmd := cmds[i]
		cmds = append(cmds, currCmd.Commands()...)
	}
}

type NewCommand func(name, fullName string, in io.Reader, out, errout io.Writer) *cobra.Command

// CommandFor returns the appropriate command for this base name,
// or the OpenShift CLI command.
func CommandFor(basename string, proc NewCommand) *cobra.Command {
	var cmd *cobra.Command

	in, out, errout := os.Stdin, os.Stdout, os.Stderr

	// Make case-insensitive and strip executable suffix if present
	if runtime.GOOS == "windows" {
		basename = strings.ToLower(basename)
		basename = strings.TrimSuffix(basename, ".exe")
	}

	cmd = proc(basename, basename, in, out, errout)

	if cmd.UsageFunc() == nil {
		templates.ActsAsRootCommand(cmd, []string{"options"})
	}
	flagtypes.GLog(cmd.PersistentFlags())

	return cmd
}

type MakeCommandGroups func(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) templates.CommandGroups

func NewCommandCommon(name, fullName, descShort, descLong, explain string, in io.Reader, out, errout io.Writer, getgroups MakeCommandGroups) *cobra.Command {

	// Main command
	cmds := &cobra.Command{
		Use:   name,
		Short: descShort,
		Long:  descLong,
		Run: func(c *cobra.Command, args []string) {
			c.SetOutput(out)
			cmdutil.RequireNoArguments(c, args)
			fmt.Fprint(out, descLong)
			fmt.Fprintf(out, explain, fullName)
		},
		//TODO ; do this later
		//BashCompletionFunction: bashCompletionFunc,
	}

	f := clientcmd.New(cmds.PersistentFlags())
	groups := getgroups(fullName, f, in, out)
	groups.Add(cmds)
	ChangeSharedFlagDefaults(cmds)

	filters := []string{
		"options",
	}

	templates.ActsAsRootCommand(cmds, filters, groups...).
		ExposeFlags(groups[0].Commands[0], "server", "client-certificate",
		"client-key", "certificate-authority", "insecure-skip-tls-verify", "token")

	cmds.AddCommand(NewCmdOptions(out))
	return cmds
}