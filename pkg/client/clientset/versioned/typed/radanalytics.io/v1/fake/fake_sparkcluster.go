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

package fake

import (
	radanalytics_io_v1 "github.com/radanalyticsio/oshinko-cli/pkg/apis/radanalytics.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSparkClusters implements SparkClusterInterface
type FakeSparkClusters struct {
	Fake *FakeRadanalyticsV1
	ns   string
}

var sparkclustersResource = schema.GroupVersionResource{Group: "radanalytics.io", Version: "v1", Resource: "sparkclusters"}

var sparkclustersKind = schema.GroupVersionKind{Group: "radanalytics.io", Version: "v1", Kind: "SparkCluster"}

// Get takes name of the sparkCluster, and returns the corresponding sparkCluster object, and an error if there is any.
func (c *FakeSparkClusters) Get(name string, options v1.GetOptions) (result *radanalytics_io_v1.SparkCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(sparkclustersResource, c.ns, name), &radanalytics_io_v1.SparkCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*radanalytics_io_v1.SparkCluster), err
}

// List takes label and field selectors, and returns the list of SparkClusters that match those selectors.
func (c *FakeSparkClusters) List(opts v1.ListOptions) (result *radanalytics_io_v1.SparkClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(sparkclustersResource, sparkclustersKind, c.ns, opts), &radanalytics_io_v1.SparkClusterList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &radanalytics_io_v1.SparkClusterList{}
	for _, item := range obj.(*radanalytics_io_v1.SparkClusterList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested sparkClusters.
func (c *FakeSparkClusters) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(sparkclustersResource, c.ns, opts))

}

// Create takes the representation of a sparkCluster and creates it.  Returns the server's representation of the sparkCluster, and an error, if there is any.
func (c *FakeSparkClusters) Create(sparkCluster *radanalytics_io_v1.SparkCluster) (result *radanalytics_io_v1.SparkCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(sparkclustersResource, c.ns, sparkCluster), &radanalytics_io_v1.SparkCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*radanalytics_io_v1.SparkCluster), err
}

// Update takes the representation of a sparkCluster and updates it. Returns the server's representation of the sparkCluster, and an error, if there is any.
func (c *FakeSparkClusters) Update(sparkCluster *radanalytics_io_v1.SparkCluster) (result *radanalytics_io_v1.SparkCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(sparkclustersResource, c.ns, sparkCluster), &radanalytics_io_v1.SparkCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*radanalytics_io_v1.SparkCluster), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSparkClusters) UpdateStatus(sparkCluster *radanalytics_io_v1.SparkCluster) (*radanalytics_io_v1.SparkCluster, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(sparkclustersResource, "status", c.ns, sparkCluster), &radanalytics_io_v1.SparkCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*radanalytics_io_v1.SparkCluster), err
}

// Delete takes name of the sparkCluster and deletes it. Returns an error if one occurs.
func (c *FakeSparkClusters) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(sparkclustersResource, c.ns, name), &radanalytics_io_v1.SparkCluster{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSparkClusters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(sparkclustersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &radanalytics_io_v1.SparkClusterList{})
	return err
}

// Patch applies the patch and returns the patched sparkCluster.
func (c *FakeSparkClusters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *radanalytics_io_v1.SparkCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(sparkclustersResource, c.ns, name, data, subresources...), &radanalytics_io_v1.SparkCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*radanalytics_io_v1.SparkCluster), err
}
