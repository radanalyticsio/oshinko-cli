package handlers

import (
	middleware "github.com/go-openapi/runtime/middleware"

	"github.com/redhatanalytics/oshinko-rest/restapi/operations/clusters"
)

// CreateClusterResponse create a cluster and return the representation
func CreateClusterResponse(params clusters.CreateClusterParams) middleware.Responder {
	// create a cluster here
	// INFO(elmiko) my thinking on creating clusters is that we should use a
	// label on the items we create with kubernetes so that we can get them
	// all with a request.
	// in addition to labels for general identification, we should then use
	// annotations on objects to help further refine what we are dealing with.
	payload := makeSingleErrorResponse(501, "Not Implemented",
		"operation clusters.CreateCluster has not yet been implemented")
	return clusters.NewCreateClusterDefault(501).WithPayload(payload)
}

// DeleteClusterResponse delete a cluster
func DeleteClusterResponse(params clusters.DeleteSingleClusterParams) middleware.Responder {
	payload := makeSingleErrorResponse(501, "Not Implemented",
		"operation clusters.DeleteSingleCluster has not yet been implemented")
	return clusters.NewCreateClusterDefault(501).WithPayload(payload)
}

// FindClustersResponse find a cluster and return its representation
func FindClustersResponse() middleware.Responder {
	payload := makeSingleErrorResponse(501, "Not Implemented",
		"operation clusters.FindClusters has not yet been implemented")
	return clusters.NewCreateClusterDefault(501).WithPayload(payload)
}

// FindSingleClusterResponse find a cluster and return its representation
func FindSingleClusterResponse(clusters.FindSingleClusterParams) middleware.Responder {
	payload := makeSingleErrorResponse(501, "Not Implemented",
		"operation clusters.FindSingleCluster has not yet been implemented")
	return clusters.NewCreateClusterDefault(501).WithPayload(payload)
}

// UpdateSingleClusterResponse update a cluster and return the new representation
func UpdateSingleClusterResponse(params clusters.UpdateSingleClusterParams) middleware.Responder {
	payload := makeSingleErrorResponse(501, "Not Implemented",
		"operation clusters.UpdateSingleCluster has not yet been implemented")
	return clusters.NewCreateClusterDefault(501).WithPayload(payload)
}