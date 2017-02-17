package clusters

import (
	"fmt"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"strconv"
	"strings"
)

type ClusterConfig struct {
	MasterCount       int
	WorkerCount       int
	Name              string
	SparkMasterConfig string
	SparkWorkerConfig string
	SparkImage        string
}

var defaultConfig ClusterConfig = ClusterConfig{
	MasterCount:       1,
	WorkerCount:       1,
	Name:              "default",
	SparkMasterConfig: "",
	SparkWorkerConfig: "",
	SparkImage:        ""}

const Defaultname = "default"
const failOnMissing = true
const allowMissing = false

const MasterCountMustBeOne = "cluster configuration must have a masterCount of 1"
const WorkerCountMustBeAtLeastOne = "cluster configuration may not have a workerCount less than 1"
const ErrorWhileProcessing = "'%s', %s"
const NamedConfigDoesNotExist = "named config '%s' does not exist"

// This function is meant to support testability
func GetDefaultConfig() ClusterConfig {
	return defaultConfig
}

func assignConfig(res *ClusterConfig, src ClusterConfig) {
	if src.Name != "" {
		res.Name = src.Name
	}
	if src.MasterCount != 0 {
		res.MasterCount = src.MasterCount
	}
	if src.WorkerCount != 0 {
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
}

func checkConfiguration(config ClusterConfig) error {
	var err error
	if config.MasterCount != 1 {
		err = NewClusterError(MasterCountMustBeOne, ClusterConfigCode)
	} else if config.WorkerCount < 1 {
		err = NewClusterError(WorkerCountMustBeAtLeastOne, ClusterConfigCode)
	}
	return err
}

func getInt(value, configmapname string) (int, error) {
	i, err := strconv.Atoi(strings.Trim(value, "\n"))
	if err != nil {
		err = NewClusterError(fmt.Sprintf(ErrorWhileProcessing, configmapname, "expected integer"), ClusterConfigCode)
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
		config.MasterCount, err = getInt(value, configmapname+".mastercount")
	case "workercount":
		config.WorkerCount, err = getInt(value, configmapname+".workercount")
	case "sparkmasterconfig":
		config.SparkMasterConfig = strings.Trim(value, "\n")
	case "sparkworkerconfig":
		config.SparkWorkerConfig = strings.Trim(value, "\n")
	case "sparkimage":
		config.SparkImage = strings.Trim(value, "\n")
	}
	return err
}

func readConfig(name string, res *ClusterConfig, failOnMissing bool, cm kclient.ConfigMapsInterface) (err error) {

	cmap, err := cm.Get(name)
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
		for n, v := range cmap.Data {
			err = process(res, n, v, name)
			if err != nil {
				break
			}
		}
	}
	return err
}

func loadConfig(name string, cm kclient.ConfigMapsInterface) (res ClusterConfig, err error) {
	// If the default config has been modified use those mods.
	res = defaultConfig
	err = readConfig(Defaultname, &res, allowMissing, cm)
	if err == nil && name != "" && name != Defaultname {
		err = readConfig(name, &res, failOnMissing, cm)
	}
	return res, err
}

func GetClusterConfig(config *ClusterConfig, cm kclient.ConfigMapsInterface) (res ClusterConfig, err error) {
	var name string = ""
	if config != nil {
		name = config.Name
	}
	res, err = loadConfig(name, cm)
	if err == nil && config != nil {
		assignConfig(&res, *config)
	}

	// Check that the final configuration is valid
	if err == nil {
		err = checkConfiguration(res)
	}
	return res, err
}
