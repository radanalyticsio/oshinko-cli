package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/intstr"
	"fmt"
)

// NameFromCommandArgs is a utility function for commands that assume the first argument is a resource name
func NameFromCommandArgs(cmd *cobra.Command, args []string) (string, error) {
	if(cmd.Name() == "get") {
		if len(args) == 0 {
			return "", nil
		} else {
			return args[0], nil
		}
	}
	if len(args) == 0 {
		return "", cmdutil.UsageError(cmd, "NAME is required")
	}
	return args[0], nil
}

func getIntValue(intString string) (int, error) {
	if len(intString) == 0 {
		return 0, nil
	}
	var newIntStr intstr.IntOrString
	integer, err := strconv.Atoi(intString)
	if err != nil {
		newIntStr = intstr.FromString(intString)
	} else {
		newIntStr = intstr.FromInt(integer)
	}
	return newIntStr.IntValue(), nil
}

func ErrorString(msg string, err error)  error {
	return fmt.Errorf(msg + ", %s", err)
}
