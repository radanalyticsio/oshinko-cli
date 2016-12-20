package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/intstr"
	"strconv"
)

// NameFromCommandArgs is a utility function for commands that assume the first argument is a resource name
func NameFromCommandArgs(cmd *cobra.Command, args []string) (string, error) {
	if len(args) == 0 {
		return "", getErrorForNoName(cmd)
	}
	return args[0], nil
}

func getErrorForNoName(cmd *cobra.Command) error {
	if cmd.Name() == "get" {
		return nil
	} else {
		return cmdutil.UsageError(cmd, "NAME is required")
	}
}

func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
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

func ErrorString(msg string, err error) error {
	return fmt.Errorf(msg+", %s", err)
}

func PrintOutput(format string, clusters []SparkCluster) (string, error) {
	var msg string
	tmpCluster := clusters
	if format == "yaml" {
		y, err := yaml.Marshal(tmpCluster)
		if err != nil {
			return "", err
		}
		msg += string(y)
		fmt.Printf(msg)
	} else if format == "json" {
		y, err := json.Marshal(tmpCluster)

		if err != nil {
			return "", err
		}
		pmsg, err := prettyprint(y)
		if err != nil {
			return "", err
		}
		msg += string(pmsg)
		fmt.Printf(msg)
	}
	return msg, nil
}
