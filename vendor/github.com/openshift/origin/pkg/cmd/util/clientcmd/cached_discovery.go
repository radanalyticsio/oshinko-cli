package clientcmd

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/typed/discovery"
	"k8s.io/kubernetes/pkg/runtime"
)

// CachedDiscoveryClient implements the functions that discovery server-supported API groups,
// versions and resources.
type CachedDiscoveryClient struct {
	discovery.DiscoveryInterface

	// cacheDirectory is the directory where discovery docs are held.  It must be unique per host:port combination to work well.
	cacheDirectory string

	// ttl is how long the cache should be considered valid
	ttl time.Duration
}

// ServerResourcesForGroupVersion returns the supported resources for a group and version.
func (d *CachedDiscoveryClient) ServerResourcesForGroupVersion(groupVersion string) (*unversioned.APIResourceList, error) {
	filename := filepath.Join(d.cacheDirectory, groupVersion, "serverresources.json")
	cachedBytes, err := d.getCachedFile(filename)
	// don't fail on errors, we either don't have a file or won't be able to run the cached check. Either way we can fallback.
	if err == nil {
		cachedResources := &unversioned.APIResourceList{}
		if err := runtime.DecodeInto(kapi.Codecs.UniversalDecoder(), cachedBytes, cachedResources); err == nil {
			glog.V(6).Infof("returning cached discovery info from %v", filename)
			return cachedResources, nil
		}
	}

	liveResources, err := d.DiscoveryInterface.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return liveResources, err
	}

	if err := d.writeCachedFile(filename, liveResources); err != nil {
		glog.V(3).Infof("failed to write cache to %v due to %v", filename, err)
	}

	return liveResources, nil
}

// ServerResources returns the supported resources for all groups and versions.
func (d *CachedDiscoveryClient) ServerResources() (map[string]*unversioned.APIResourceList, error) {
	apiGroups, err := d.ServerGroups()
	if err != nil {
		return nil, err
	}
	groupVersions := unversioned.ExtractGroupVersions(apiGroups)
	result := map[string]*unversioned.APIResourceList{}
	for _, groupVersion := range groupVersions {
		resources, err := d.ServerResourcesForGroupVersion(groupVersion)
		if err != nil {
			return nil, err
		}
		result[groupVersion] = resources
	}
	return result, nil
}

func (d *CachedDiscoveryClient) ServerGroups() (*unversioned.APIGroupList, error) {
	filename := filepath.Join(d.cacheDirectory, "servergroups.json")
	cachedBytes, err := d.getCachedFile(filename)
	// don't fail on errors, we either don't have a file or won't be able to run the cached check. Either way we can fallback.
	if err == nil {
		cachedGroups := &unversioned.APIGroupList{}
		if err := runtime.DecodeInto(kapi.Codecs.UniversalDecoder(), cachedBytes, cachedGroups); err == nil {
			glog.V(6).Infof("returning cached discovery info from %v", filename)
			return cachedGroups, nil
		}
	}

	liveGroups, err := d.DiscoveryInterface.ServerGroups()
	if err != nil {
		return liveGroups, err
	}

	if err := d.writeCachedFile(filename, liveGroups); err != nil {
		glog.V(3).Infof("failed to write cache to %v due to %v", filename, err)
	}

	return liveGroups, nil
}

func (d *CachedDiscoveryClient) getCachedFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if time.Now().After(fileInfo.ModTime().Add(d.ttl)) {
		return nil, errors.New("cache expired")
	}

	// the cache is present and its valid.  Try to read and use it.
	cachedBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return cachedBytes, nil
}

func (d *CachedDiscoveryClient) writeCachedFile(filename string, obj runtime.Object) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	bytes, err := runtime.Encode(kapi.Codecs.LegacyCodec(), obj)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, bytes, 0755)
}

// NewCachedDiscoveryClient creates a new DiscoveryClient.  cacheDirectory is the directory where discovery docs are held.  It must be unique per host:port combination to work well.
func NewCachedDiscoveryClient(delegate discovery.DiscoveryInterface, cacheDirectory string, ttl time.Duration) *CachedDiscoveryClient {
	return &CachedDiscoveryClient{DiscoveryInterface: delegate, cacheDirectory: cacheDirectory, ttl: ttl}
}
