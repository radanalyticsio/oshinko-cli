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
	v1 "github.com/radanalyticsio/oshinko-cli/pkg/apis/radanalytics.io/v1"
	scheme "github.com/radanalyticsio/oshinko-cli/pkg/client/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SparkClustersGetter has a method to return a SparkClusterInterface.
// A group's client should implement this interface.
type SparkClustersGetter interface {
	SparkClusters(namespace string) SparkClusterInterface
}

// SparkClusterInterface has methods to work with SparkCluster resources.
type SparkClusterInterface interface {
	Create(*v1.SparkCluster) (*v1.SparkCluster, error)
	Update(*v1.SparkCluster) (*v1.SparkCluster, error)
	UpdateStatus(*v1.SparkCluster) (*v1.SparkCluster, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.SparkCluster, error)
	List(opts meta_v1.ListOptions) (*v1.SparkClusterList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SparkCluster, err error)
	SparkClusterExpansion
}

// sparkClusters implements SparkClusterInterface
type sparkClusters struct {
	client rest.Interface
	ns     string
}

// newSparkClusters returns a SparkClusters
func newSparkClusters(c *RadanalyticsV1Client, namespace string) *sparkClusters {
	return &sparkClusters{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the sparkCluster, and returns the corresponding sparkCluster object, and an error if there is any.
func (c *sparkClusters) Get(name string, options meta_v1.GetOptions) (result *v1.SparkCluster, err error) {
	result = &v1.SparkCluster{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sparkclusters").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SparkClusters that match those selectors.
func (c *sparkClusters) List(opts meta_v1.ListOptions) (result *v1.SparkClusterList, err error) {
	result = &v1.SparkClusterList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sparkclusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested sparkClusters.
func (c *sparkClusters) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("sparkclusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a sparkCluster and creates it.  Returns the server's representation of the sparkCluster, and an error, if there is any.
func (c *sparkClusters) Create(sparkCluster *v1.SparkCluster) (result *v1.SparkCluster, err error) {
	result = &v1.SparkCluster{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("sparkclusters").
		Body(sparkCluster).
		Do().
		Into(result)
	return
}

// Update takes the representation of a sparkCluster and updates it. Returns the server's representation of the sparkCluster, and an error, if there is any.
func (c *sparkClusters) Update(sparkCluster *v1.SparkCluster) (result *v1.SparkCluster, err error) {
	result = &v1.SparkCluster{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sparkclusters").
		Name(sparkCluster.Name).
		Body(sparkCluster).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *sparkClusters) UpdateStatus(sparkCluster *v1.SparkCluster) (result *v1.SparkCluster, err error) {
	result = &v1.SparkCluster{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sparkclusters").
		Name(sparkCluster.Name).
		SubResource("status").
		Body(sparkCluster).
		Do().
		Into(result)
	return
}

// Delete takes name of the sparkCluster and deletes it. Returns an error if one occurs.
func (c *sparkClusters) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sparkclusters").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *sparkClusters) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sparkclusters").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched sparkCluster.
func (c *sparkClusters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SparkCluster, err error) {
	result = &v1.SparkCluster{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("sparkclusters").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
