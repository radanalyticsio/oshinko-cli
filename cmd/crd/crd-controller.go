package main

import (
	"flag"
	"os"
	"os/signal"
	"github.com/golang/glog"
	rad "github.com/radanalyticsio/oshinko-cli/pkg/client/clientset/versioned"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	informers "github.com/radanalyticsio/oshinko-cli/pkg/client/informers/externalversions"
	"github.com/radanalyticsio/oshinko-cli/pkg/crd/controller"
	"github.com/radanalyticsio/oshinko-cli/pkg/signals"
	"github.com/radanalyticsio/oshinko-cli/version"
	"time"
	//"fmt"
)

var (
	configPath string
	masterURL  string
	namespace  string
)

func init() {
	const (
		MY_NAMESPACE = "MY_NAMESPACE"
	)
	// flags
	flag.StringVar(&configPath, "kubeconfig", os.Getenv("KUBECONFIG"), "(Optional) Overrides $KUBECONFIG")
	flag.StringVar(&masterURL, "server", "", "(Optional) URL address of a remote api server.  Do not set for local clusters.")
	flag.StringVar(&namespace, "ns", "", "(Optional) URL address of a remote api server.  Do not set for local clusters.")
	flag.Parse()
	// env variables
	namespace = os.Getenv(MY_NAMESPACE)
	if namespace == "" {
		glog.Infof("usage: oshinko-crd --kubeconfig=$HOME.kube/config --logtostderr")
		glog.Fatalf("init: crd's namespace was not passed in env variable %q", MY_NAMESPACE)
	}

	glog.Infof("%s %s Default spark image: %s\n", version.GetAppName(), version.GetVersion(), version.GetSparkImage())
	glog.Infof("init complete: Oshinko crd version %q starting...\n", version.GetVersion())
}

func main() {

	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, configPath)
	if err != nil {
		glog.Fatalf("Error getting kube config: %v\n", err)
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	radClient, err := rad.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %v", err)
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	sparkInformerFactory := informers.NewSharedInformerFactory(radClient, time.Second*30)

	controller := controller.NewController(kubeClient, radClient, kubeInformerFactory, sparkInformerFactory)

	go kubeInformerFactory.Start(stopCh)
	go sparkInformerFactory.Start(stopCh)

	if err = controller.Run(1, stopCh); err != nil {
		glog.Fatalf("Error running crd: %s", err.Error())
	}

	/*
		working code
	*/
	//list, err := radClient.RadanalyticsV1().SparkClusters(namespace).List(metav1.ListOptions{})
	//if err != nil {
	//	glog.Fatalf("Error listing all databases: %v", err)
	//}
	//
	//for _, s := range list.Items {
	//	fmt.Printf("database %s with user %q\n", s.Name, s.Spec.Name)
	//}
	/*
		working code
	*/
}

// Shutdown gracefully on system signals
func handleSignals() <-chan struct{} {
	sigCh := make(chan os.Signal)
	stopCh := make(chan struct{})
	go func() {
		signal.Notify(sigCh)
		<-sigCh
		close(stopCh)
		os.Exit(1)
	}()
	return stopCh
}
