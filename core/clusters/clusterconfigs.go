package clusters

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
	"k8s.io/client-go/rest"
)

type ClusterConfig struct {
	MasterCount       int
	WorkerCount       int
	Name              string
	SparkMasterConfig string
	SparkWorkerConfig string
	SparkImage        string
	ExposeWebUI       string
	Metrics           string
}

const Defaultname = "default-oshinko-cluster-config"

var defaultConfig ClusterConfig = ClusterConfig{
	MasterCount:       1,
	WorkerCount:       1,
	Name:              "",
	SparkMasterConfig: "",
	SparkWorkerConfig: "",
	SparkImage:        "",
	ExposeWebUI:       "true",
	Metrics: 	   "false",
}

const failOnMissing = true
const allowMissing = false

const MasterCountMustBeZeroOrOne = "cluster configuration must have a master count of 0 or 1"
const WorkerCountMustBeAtLeastZero = "cluster configuration may not have a worker count less than 0"
const ErrorWhileProcessing = "'%s', %s"
const NamedConfigDoesNotExist = "named config '%s' does not exist"

const SentinelCountValue = -1

// This function is meant to support testability
func GetDefaultConfig() ClusterConfig {
	return defaultConfig
}

func assignConfig(res *ClusterConfig, src ClusterConfig) {
	if src.Name != "" {
		res.Name = src.Name
	}

	if src.MasterCount > SentinelCountValue {
		res.MasterCount = src.MasterCount
	}

	if src.WorkerCount > SentinelCountValue {
		res.WorkerCount = src.WorkerCount
	}

	if src.SparkMasterConfig != "" {
		res.SparkMasterConfig = src.SparkMasterConfig
	}
	if src.SparkWorkerConfig != "" {
		res.SparkWorkerConfig = src.SparkWorkerConfig
	}
	if src.SparkImage != "" {
		res.SparkImage = src.SparkImage
	}
	if src.ExposeWebUI != "" {
		res.ExposeWebUI = src.ExposeWebUI
	}
	if src.Metrics != "" {
		res.Metrics = src.Metrics
	}
}

func checkConfiguration(config ClusterConfig) error {
	var err error
	if config.MasterCount < 0 || config.MasterCount > 1 {
		err = NewClusterError(MasterCountMustBeZeroOrOne, ClusterConfigCode)
	} else if config.WorkerCount < 0 {
		err = NewClusterError(WorkerCountMustBeAtLeastZero, ClusterConfigCode)
	}
	return err
}

func getInt(value, configmapname string) (int, error) {
	i, err := strconv.Atoi(value)
	if err != nil {
		err = NewClusterError(fmt.Sprintf(ErrorWhileProcessing, configmapname, fmt.Sprintf("expected integer, got '%s'", value)), ClusterConfigCode)
	}
	return i, err
}

func process(config *ClusterConfig, name, value, configmapname string) error {

	var err error

	// At present we only have a single level of configs, but if/when we have
	// nested configs then we would descend through the levels beginning here with
	// the first element in the name
	switch name {
	case "mastercount":
		val, err := getInt(value, configmapname + ".mastercount")
		if err != nil {
			return err
		}
		if val > SentinelCountValue {
			config.MasterCount = val
		}
	case "workercount":
		val, err := getInt(value, configmapname + ".workercount")
		if err != nil {
			return err
		}
		if val > SentinelCountValue {
			config.WorkerCount = val
		}
	case "sparkmasterconfig":
		config.SparkMasterConfig = value
	case "sparkworkerconfig":
		config.SparkWorkerConfig = value
	case "sparkimage":
		config.SparkImage = value
	case "exposeui":
		config.ExposeWebUI = value
		_, err = strconv.ParseBool(config.ExposeWebUI)
	case "metrics":
		// default will be "false" if the string is empty
		if value != "" {
			val, err := strconv.ParseBool(value)
			if err == nil {
				//  Support 'true' and default to jolokia
				// Normalize truth values
				if val {
					config.Metrics = "true"
				} else {
					config.Metrics = "false"
				}
			} else {
				if value != "jolokia" && value != "prometheus" {
					msg := fmt.Sprintf("expected 'jolokia' or 'prometheus', got '%s'", value)
					return NewClusterError(fmt.Sprintf(ErrorWhileProcessing, configmapname, msg), ClusterConfigCode)
				}
				config.Metrics = value
			}
		}
	}
	return err
}

func readConfig(name string, res *ClusterConfig, failOnMissing bool, restconfig *rest.Config, namespace string) (found bool, err error) {

	found = false
	cmap, err := getKubeClient(restconfig).CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if strings.Index(err.Error(), "not found") != -1 {
			if !failOnMissing {
				err = nil
			} else {
				err = NewClusterError(fmt.Sprintf(NamedConfigDoesNotExist, name), ClusterConfigCode)
			}
		} else {
			err = NewClusterError(err.Error(), ClientOperationCode)
		}
	}
	if err == nil && cmap != nil {
		// Kube will give us an empty configmap if the named one does not exist,
		// so we test for a Name to see if we foud it
		found = cmap.Name != ""
		for n, v := range cmap.Data {
			err = process(res, strings.Trim(n, "\n"), strings.Trim(v, "\n"), name)
			if err != nil {
				break
			}
		}
	}
	return found, err
}

func loadConfig(name string, restconfig *rest.Config, namespace string) (res ClusterConfig, err error) {
	// If the default config has been modified use those mods.
	res = defaultConfig
	found, err := readConfig(Defaultname, &res, allowMissing, restconfig, namespace)
	if err == nil {
		if name != "" && name != Defaultname {
			_, err = readConfig(name, &res, failOnMissing, restconfig, namespace)
		} else if found {
			res.Name = Defaultname
		}
	}
	return res, err
}

func GetClusterConfig(config *ClusterConfig, restconfig *rest.Config, namespace string) (res ClusterConfig, err error) {

	var name string = ""
	if config != nil {
		name = config.Name
	}
	res, err = loadConfig(name, restconfig, namespace)
	if err == nil && config != nil {
		assignConfig(&res, *config)
	}

	// Check that the final configuration is valid
	if err == nil {
		err = checkConfiguration(res)
	}
	return res, err
}
