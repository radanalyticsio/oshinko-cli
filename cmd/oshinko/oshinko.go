package main

import (
	"os"
	"path"
	"path/filepath"
	"runtime"

	"k8s.io/apiserver/pkg/util/logs"

	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/cli"
	// install all APIs
	_ "k8s.io/kubernetes/pkg/apis/core/install"
	_ "k8s.io/kubernetes/pkg/apis/autoscaling/install"
	_ "k8s.io/kubernetes/pkg/apis/batch/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"
	"fmt"
)

func main() {
	if path.Base(os.Args[0]) == "oshinko-cli" {
		fmt.Println("\n*** The 'oshinko-cli' command has been deprecated. Use 'oshinko' instead. ***\n")
	}
	logs.InitLogs()
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	basename := filepath.Base(os.Args[0])
	command := cli.CommandFor(basename)
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}