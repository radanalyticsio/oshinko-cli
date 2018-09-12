package clusters

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"strconv"
	"strings"
)

type ClusterConfig struct {
	MastersCount      int
	WorkersCount      int
	Name              string `json:"ConfigName,omitempty"`
	SparkMasterConfig string `json:"SparkMasterConfig,omitempty"`
	SparkWorkerConfig string `json:"SparkWorkerConfig,omitempty"`
	SparkImage        string `json:"SparkImage,omitempty"`
	ExposeWebUI       string
	Metrics           string
}

const Defaultname = "default-oshinko-cluster-config"

var defaultConfig ClusterConfig = ClusterConfig{
	MastersCount:      1,
	WorkersCount:      1,
	Name:              "",
	SparkMasterConfig: "",
	SparkWorkerConfig: "",
	SparkImage:        "",
	ExposeWebUI:       "true",
	Metrics:           "false",
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

	if src.MastersCount > SentinelCountValue {
		res.MastersCount = src.MastersCount
	}

	if src.WorkersCount > SentinelCountValue {
		res.WorkersCount = src.WorkersCount
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
	if config.MastersCount < 0 || config.MastersCount > 1 {
		err = NewClusterError(MasterCountMustBeZeroOrOne, ClusterConfigCode)
	} else if config.WorkersCount < 0 {
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
		val, err := getInt(value, configmapname+".mastercount")
		if err != nil {
			return err
		}
		if val > SentinelCountValue {
			config.MastersCount = val
		}
	case "workercount":
		val, err := getInt(value, configmapname+".workercount")
		if err != nil {
			return err
		}
		if val > SentinelCountValue {
			config.WorkersCount = val
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
		if err != nil {
			return err
		}
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
	default:
		err = NewClusterError(fmt.Sprintf(ErrorWhileProcessing, configmapname+"."+name, "unrecognized configuration field"), ClusterConfigCode)
	}
	return err
}

func readConfig(name string, res *ClusterConfig, failOnMissing bool, restconfig *rest.Config, namespace string) (found bool, err error) {

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

	// If we actually found a configMap then it will be non-empty
	found = err == nil && cmap != nil && cmap.Name != ""
	if found {
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
	defaultFound, err := readConfig(Defaultname, &res, allowMissing, restconfig, namespace)
	if err == nil {
		//process config if it is not named default
		if name != "" && name != Defaultname {
			_, err = readConfig(name, &res, failOnMissing, restconfig, namespace)
		} else if defaultFound {
			// If the default oshinko cluster config has been overridden by a user with a configMap
			// named Defaultname, then we want to record the name as non-empty to indicate that
			// a configmap was actually located and read vs using the hardcoded default
			res.Name = Defaultname
		}
	}
	return res, err
}

func GetClusterConfig(config *ClusterConfig, restconfig *rest.Config, namespace string) (res ClusterConfig, err error) {

	var name string = ""
	if config != nil {
		// If the default name is explicitly set, put it back to empty string
		// We will record default name in the config only if we found an overriding configmap
		if config.Name == Defaultname {
			config.Name = ""
		}
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
