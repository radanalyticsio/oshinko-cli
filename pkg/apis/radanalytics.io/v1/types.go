/*
Copyright The Red Hat Authors.
https://radanalytics.io/

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GroupName       = "radanalytics.io"
	SparkListPlural = "sparkclusters"
	FullCRDName     = SparkListPlural + "." + GroupName
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SparkCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SparkClusterSpec   `json:"spec"`
	Status            SparkClusterStatus `json:"status,omitempty"`
}

type SparkPod struct {
	IP     string `json:"ip"`
	Status string `json:"status"`
	Type   string `json:"type"`
}

type ClusterConfig struct {
	MasterCount       int `json:"masterCount"`
	WorkerCount       int `json:"workerCount"`
	Name              string `json:"name"`
	SparkMasterConfig string `json:"sparkMasterConfig"`
	SparkWorkerConfig string `json:"sparkWorkerConfig"`
	SparkImage        string `json:"sparkImage"`
	ExposeWebUI       string `json:"exposeWebUI"`
	Metrics           string `json:"metrics,omitempty"`
}

// TODO: Following same format as ClusterConfig.go in oshinko-cli
type SparkClusterSpec struct {
	Namespace    string `json:"namespace,omitempty"`
	//Name         string `json:"name,omitempty"`
	Href         string `json:"href"`
	Image        string `json:"image"`
	MasterURL    string `json:"masterUrl"`
	MasterWebURL string `json:"masterWebUrl"`
	MasterWebRoute string `json:"masterWebRoute"`
	Status       string `json:"status"`
	WorkerCount  int    `json:"workerCount"`
	MasterCount  int    `json:"masterCount"`
	Config       ClusterConfig
	Ephemeral    string `json:"ephemeral,omitempty"`
	Pods         []SparkPod
	Alertrules      string `json:"alertrules,omitempty"`

	//SparkMasterName string `json"sparkmastername"`
	//SparkWorkerName string `json"sparkworkername"`
	//Image           string `json"image"`
	//Workers         int32  `json:"workers"`
	//SparkMetrics    string `json:"sparkmetrics"`
	//Alertrules      string `json:"alertrules,omitempty"`
	//NodeSelector string `json:"nodeselector"`
}

type SparkClusterStatus struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}


// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SparkClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SparkCluster `json:"items"`
}
