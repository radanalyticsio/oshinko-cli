package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	coreclusters "github.com/radanalyticsio/oshinko-cli/core/clusters"
	osa "github.com/radanalyticsio/oshinko-cli/rest/helpers/authentication"
	oe "github.com/radanalyticsio/oshinko-cli/rest/helpers/errors"
	"github.com/radanalyticsio/oshinko-cli/rest/helpers/info"
	"github.com/radanalyticsio/oshinko-cli/rest/models"
	apiclusters "github.com/radanalyticsio/oshinko-cli/rest/restapi/operations/clusters"
)

const nameSpaceMsg = "cannot determine target openshift namespace"
const clientMsg = "unable to create an openshift client"

var codes map[int]int32 = map[int]int32{
	coreclusters.NoCodeAvailable: 500,
	coreclusters.ClusterConfigCode: 409,
	coreclusters.ClientOperationCode: 500,
	coreclusters.ClusterIncompleteCode: 409,
	coreclusters.NoSuchClusterCode: 404,
	coreclusters.ComponentExistsCode: 409,
}

func generalErr(err error, title, msg string, code int32) *models.ErrorResponse {
	if err != nil {
		if msg != "" {
			msg += ", reason: "
		}
		msg += err.Error()
	}
	return oe.NewSingleErrorResponse(code, title, msg)
}

func tostrptr(val string) *string {
	v := val
	return &v
}

func getErrorCode(err error) int32 {

	code := coreclusters.ErrorCode(err)
	if httpcode, ok := codes[code]; ok {
		return httpcode
	}
	return 500

}

func int64ptr(val int) *int64 {
	if val <= coreclusters.SentinelCountValue {
		return nil
	}
	ret := int64(val)
	return &ret
}

func boolptr(val bool) *bool {
	return &val
}

func singleClusterResponse(sc coreclusters.SparkCluster) *models.SingleCluster {

	addpod := func(p coreclusters.SparkPod) *models.ClusterModelPodsItems0 {
		pod := new(models.ClusterModelPodsItems0)
		pod.IP = tostrptr(p.IP)
		pod.Status = tostrptr(p.Status)
		pod.Type = tostrptr(p.Type)
		return pod
	}

	// Build the response
	cluster := &models.SingleCluster{&models.ClusterModel{}}
	cluster.Cluster.Name = tostrptr(sc.Name)
	cluster.Cluster.MasterURL = tostrptr(sc.MasterURL)
	cluster.Cluster.MasterWebURL = tostrptr(sc.MasterWebURL)
	cluster.Cluster.MasterWebRoute = sc.MasterWebRoute
	cluster.Cluster.Status = tostrptr(sc.Status)

	cluster.Cluster.Pods = []*models.ClusterModelPodsItems0{}
	for i := range sc.Pods {
		cluster.Cluster.Pods = append(cluster.Cluster.Pods, addpod(sc.Pods[i]))
	}

	cluster.Cluster.Config = &models.NewClusterConfig{
		SparkMasterConfig: sc.Config.SparkMasterConfig,
		SparkWorkerConfig: sc.Config.SparkWorkerConfig,
		MasterCount: int64ptr(sc.Config.MasterCount),
		WorkerCount: int64ptr(sc.Config.WorkerCount),
		Name: sc.Config.Name,
		ExposeWebUI: sc.Config.ExposeWebUI,
		Metrics: sc.Config.Metrics,
		SparkImage: sc.Config.SparkImage,
	}
	return cluster
}

func getModelCount(val *int64) int {
	if val == nil {
		return coreclusters.SentinelCountValue
	}
	return int(*val)
}

func getBoolVal(val *bool) bool {
	if val == nil {
		return true
	}
	return bool(*val)
}

func assignConfig(config *models.NewClusterConfig) *coreclusters.ClusterConfig {
	if config == nil {
		return nil
	}
	result := &coreclusters.ClusterConfig{
		Name: config.Name,
		MasterCount: getModelCount(config.MasterCount),
		WorkerCount: getModelCount(config.WorkerCount),
		SparkMasterConfig: config.SparkMasterConfig,
		SparkWorkerConfig: config.SparkWorkerConfig,
		SparkImage: config.SparkImage,
		ExposeWebUI: config.ExposeWebUI,
		Metrics: config.Metrics,
	}
	return result
}

// CreateClusterResponse create a cluster and return the representation
func CreateClusterResponse(params apiclusters.CreateClusterParams) middleware.Responder {

	// Do this so that we only have to specify the error code when we build ErrorResponse
	reterr := func(err *models.ErrorResponse) *apiclusters.CreateClusterDefault {
		return apiclusters.NewCreateClusterDefault(int(*err.Errors[0].Status)).WithPayload(err)
	}

	// Convenience wrapper for create failure
	fail := func(err error, msg string, code int32) *models.ErrorResponse {
		return generalErr(err, "cannot create cluster", msg, code)
	}

	const imageMsg = "cannot determine name of spark image"

	clustername := *params.Cluster.Name

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, nameSpaceMsg, 500))
	}

	// Even if the image comes back "" at this point, let oshinko-core
	// generate an error. It is possible that the cluster config specifies
	// an image even if no default is set in the environment
	image := info.GetSparkImage()

	restConfig, err := osa.GetConfig()
	if err != nil {
		return reterr(fail(err, clientMsg, 500))
	}

	config := assignConfig(params.Cluster.Config)
	sc, err := coreclusters.CreateCluster(clustername, namespace, image, config, restConfig, "", false)
	if err != nil {
		return reterr(fail(err, "", getErrorCode(err)))
	}
	return apiclusters.NewCreateClusterCreated().WithLocation(sc.Href).WithPayload(singleClusterResponse(sc))
}

// DeleteClusterResponse delete a cluster
func DeleteClusterResponse(params apiclusters.DeleteSingleClusterParams) middleware.Responder {

	// Do this so that we only have to specify the error code when we build ErrorResponse
	reterr := func(err *models.ErrorResponse) *apiclusters.DeleteSingleClusterDefault {
		return apiclusters.NewDeleteSingleClusterDefault(int(*err.Errors[0].Status)).WithPayload(err)
	}

	// Convenience wrapper for delete failure
	fail := func(err error, msg string, code int32) *models.ErrorResponse {
		return generalErr(err, "cluster deletion failed", msg, code)
	}

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, nameSpaceMsg, 500))
	}

	restConfig, err := osa.GetConfig()
	if err != nil {
		return reterr(fail(err, clientMsg, 500))
	}

	_, err = coreclusters.DeleteCluster(params.Name, namespace, restConfig, "", "")
	if err != nil {
		return reterr(fail(err, "", getErrorCode(err)))
	}
	return apiclusters.NewDeleteSingleClusterNoContent()
}

// FindClustersResponse find a cluster and return its representation
func FindClustersResponse(params apiclusters.FindClustersParams) middleware.Responder {

	// Do this so that we only have to specify the error code when we build ErrorResponse
	reterr := func(err *models.ErrorResponse) *apiclusters.FindClustersDefault {
		return apiclusters.NewFindClustersDefault(int(*err.Errors[0].Status)).WithPayload(err)
	}

	// Convenience wrapper for list failure
	fail := func(err error, msg string, code int32) *models.ErrorResponse {
		return generalErr(err, "cannot list clusters", msg, code)
	}

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, nameSpaceMsg, 500))
	}

	restConfig, err := osa.GetConfig()
	if err != nil {
		return reterr(fail(err, clientMsg, 500))
	}

	scs, err := coreclusters.FindClusters(namespace, restConfig, "")
	if err != nil {
		return reterr(fail(err, "", getErrorCode(err)))
	}

	// Create the payload that we're going to write into for the response
	payload := apiclusters.FindClustersOKBodyBody{}
	payload.Clusters = []*apiclusters.ClustersItems0{}
	for idx := range(scs) {
		clt := new(apiclusters.ClustersItems0)
		clt.Href = &scs[idx].Href
		clt.MasterURL = &scs[idx].MasterURL
		clt.MasterWebURL = &scs[idx].MasterWebURL
		clt.Name = &scs[idx].Name
		clt.Status = &scs[idx].Status
		wc := int64(scs[idx].WorkerCount)
		clt.WorkerCount = &wc
		payload.Clusters = append(payload.Clusters, clt)
	}

	return apiclusters.NewFindClustersOK().WithPayload(payload)
}

// FindSingleClusterResponse find a cluster and return its representation
func FindSingleClusterResponse(params apiclusters.FindSingleClusterParams) middleware.Responder {

	clustername := params.Name

	// Do this so that we only have to specify the error code when we build ErrorResponse
	reterr := func(err *models.ErrorResponse) *apiclusters.FindSingleClusterDefault {
		return apiclusters.NewFindSingleClusterDefault(int(*err.Errors[0].Status)).WithPayload(err)
	}

	// Convenience wrapper for get failure
	fail := func(err error, msg string, code int32) *models.ErrorResponse {
		return generalErr(err, "cannot get cluster", msg, code)
	}

	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, nameSpaceMsg, 500))
	}

	restConfig, err := osa.GetConfig()
	if err != nil {
		return reterr(fail(err, clientMsg, 500))
	}

	sc, err := coreclusters.FindSingleCluster(clustername, namespace, restConfig)
	if err != nil {
		return reterr(fail(err, "", getErrorCode(err)))
	}

	return apiclusters.NewFindSingleClusterOK().WithPayload(singleClusterResponse(sc))
}

// UpdateSingleClusterResponse update a cluster and return the new representation
func UpdateSingleClusterResponse(params apiclusters.UpdateSingleClusterParams) middleware.Responder {

	// Do this so that we only have to specify the error code when we build ErrorResponse
	reterr := func(err *models.ErrorResponse) *apiclusters.UpdateSingleClusterDefault {
		return apiclusters.NewUpdateSingleClusterDefault(int(*err.Errors[0].Status)).WithPayload(err)
	}

	// Convenience wrapper for update failure
	fail := func(err error, msg string, code int32) *models.ErrorResponse {
		return generalErr(err, "cannot update cluster", msg, code)
	}

	const clusterNameMsg = "changing the cluster name is not supported"

	clustername := params.Name

	// Before we do further checks, make sure that we have deploymentconfigs
	// If either the master or the worker deploymentconfig are missing, we
	// assume that the cluster is missing. These are the base objects that
	// we use to create a cluster
	namespace, err := info.GetNamespace()
	if namespace == "" || err != nil {
		return reterr(fail(err, nameSpaceMsg, 500))
	}

	restConfig, err := osa.GetConfig()
	if err != nil {
		return reterr(fail(err, clientMsg, 500))
	}


	// Simple things first. At this time we do not support cluster name change and
	// we do not suppport scaling the master count (likely need HA setup for that to make sense)
	if clustername != *params.Cluster.Name {
		return reterr(fail(nil, clusterNameMsg, 409))
	}

	config := assignConfig(params.Cluster.Config)
	sc, err := coreclusters.UpdateCluster(clustername, namespace, config, restConfig)
	if err != nil {
		return reterr(fail(err, "", getErrorCode(err)))
	}
	return apiclusters.NewUpdateSingleClusterAccepted().WithPayload(singleClusterResponse(sc))
}
