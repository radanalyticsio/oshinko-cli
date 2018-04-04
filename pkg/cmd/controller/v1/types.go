package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SparkListPlural = "sparkclusters"
	FullCRDName     = SparkListPlural + "." + GroupName
)

type SparkCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SparkClusterSpec   `json:"spec"`
	Status            SparkClusterStatus `json:"status,omitempty"`
}
type SparkClusterStatus struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}
type SparkClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SparkCluster `json:"items"`
}

// TODO: Following same format as ClusterConfig.go in oshinko-cli
type SparkClusterSpec struct {
	SparkMasterName string `json"sparkmastername"`
	SparkWorkerName string `json"sparkworkername"`
	Image           string `json"image"`
	Workers         int32  `json:"workers"`
	SparkMetrics    string `json:"sparkmetrics"`
	Alertrules      string `json:"alertrules"`
	//NodeSelector string `json:"nodeselector"`
}
