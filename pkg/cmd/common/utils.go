package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	//"github.com/radanalyticsio/oshinko-core/clusters"
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

func PrintOutput(format string, clusters interface {}) error {
	var msg string
	tmpCluster := clusters
	if format == "yaml" {
		y, err := yaml.Marshal(tmpCluster)
		if err != nil {
			return err
		}
		msg += string(y)
		fmt.Printf(msg)
	} else if format == "json" {
		y, err := json.Marshal(tmpCluster)

		if err != nil {
			return err
		}
		pmsg, err := prettyprint(y)
		if err != nil {
			return err
		}
		msg += string(pmsg)
		fmt.Printf(msg)
	}
	return nil
}
