package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/extended"
	"github.com/radanalyticsio/oshinko-cli/pkg/cmd/common"
	// install all APIs
	_ "github.com/openshift/origin/pkg/api/install"
	_ "k8s.io/kubernetes/pkg/api/install"
	_ "k8s.io/kubernetes/pkg/apis/autoscaling/install"
	_ "k8s.io/kubernetes/pkg/apis/batch/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"
)

func main() {
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	basename := filepath.Base(os.Args[0])
	command := common.CommandFor(basename, extended.NewCommandExtended)
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
