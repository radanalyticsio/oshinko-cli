package cmd

import (
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

// NameFromCommandArgs is a utility function for commands that assume the first argument is a resource name
func NameFromCommandArgs(cmd *cobra.Command, args []string) (string, error) {
	if len(args) == 0 {
		return "", cmdutil.UsageError(cmd, "NAME is required")
	}
	return args[0], nil
}
