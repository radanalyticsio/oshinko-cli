package controller

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/radanalyticsio/oshinko-cli/pkg/apis/radanalytics.io/v1"
	rad "github.com/radanalyticsio/oshinko-cli/pkg/client/clientset/versioned"
	sparkscheme "github.com/radanalyticsio/oshinko-cli/pkg/client/clientset/versioned/scheme"
	informers "github.com/radanalyticsio/oshinko-cli/pkg/client/informers/externalversions"
	listers "github.com/radanalyticsio/oshinko-cli/pkg/client/listers/radanalytics.io/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"time"
)

const (
	controllerAgentName   = "radanalyticsio-spark-crd"
	SuccessSynced         = "Synced"
	MessageResourceSynced = "SparkCluster synced successfully"
)

type Controller struct {
	kubeclientset     kubernetes.Interface
	radclientset      rad.Interface
	queue             workqueue.RateLimitingInterface
	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	sparksLister      listers.SparkClusterLister
	sparksSynced      cache.InformerSynced
	recorder          record.EventRecorder
}

func NewController(
	kclient kubernetes.Interface,
	radclient rad.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	sampleInformerFactory informers.SharedInformerFactory,
) *Controller {

	deploymentInformer := kubeInformerFactory.Apps().V1().Deployments()
	sparkInformer := sampleInformerFactory.Radanalytics().V1().SparkClusters()

	sparkscheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kclient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kclient,
		radclientset:      radclient,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		sparksLister:      sparkInformer.Lister(),
		sparksSynced:      sparkInformer.Informer().HasSynced,
		queue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "RadSpark"),
		recorder:          recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	sparkInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c := obj.(*v1.SparkCluster)
			glog.Info("SparkCluster Added: ", c.Name)
			controller.enqueueSparkCluster(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			c := new.(*v1.SparkCluster)
			glog.Info("SparkCluster Update: ", c.Name)
			controller.enqueueSparkCluster(new)
		},
		DeleteFunc: func(obj interface{}) {
			c := obj.(*v1.SparkCluster)
			glog.Info("SparkCluster Delete: ", c.Name)
			//oshinkocli.DeleteAll(config, cluster)
		},
	})
	return controller
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting SparkCluster crd")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.sparksSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch  workers to process
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers...")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.queue.Get()
	glog.Infof("processNextWorkItem '%s'", obj)

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.queue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.queue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.queue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}
func (c *Controller) enqueueSparkCluster(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.queue.AddRateLimited(key)
}

func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the SparkClusters resource with this namespace/name
	sparkc, err := c.sparksLister.SparkClusters(namespace).Get(name)
	if err != nil {
		// The SparkClusters resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("sparkc '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}
	c.recorder.Event(sparkc, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}
