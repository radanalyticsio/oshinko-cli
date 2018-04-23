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
	radanalytics_redhat_com_v1 "github.com/radanalyticsio/oshinko-cli/pkg/apis/radanalytics.redhat.com/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSparkClusterLists implements SparkClusterListInterface
type FakeSparkClusterLists struct {
	Fake *FakeRadanalyticsV1
	ns   string
}

var sparkclusterlistsResource = schema.GroupVersionResource{Group: "radanalytics.redhat.com", Version: "v1", Resource: "sparkclusterlists"}

var sparkclusterlistsKind = schema.GroupVersionKind{Group: "radanalytics.redhat.com", Version: "v1", Kind: "SparkClusterList"}

// Get takes name of the sparkClusterList, and returns the corresponding sparkClusterList object, and an error if there is any.
func (c *FakeSparkClusterLists) Get(name string, options v1.GetOptions) (result *radanalytics_redhat_com_v1.SparkClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(sparkclusterlistsResource, c.ns, name), &radanalytics_redhat_com_v1.SparkClusterList{})

	if obj == nil {
		return nil, err
	}
	return obj.(*radanalytics_redhat_com_v1.SparkClusterList), err
}

// List takes label and field selectors, and returns the list of SparkClusterLists that match those selectors.
func (c *FakeSparkClusterLists) List(opts v1.ListOptions) (result *radanalytics_redhat_com_v1.SparkClusterListList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(sparkclusterlistsResource, sparkclusterlistsKind, c.ns, opts), &radanalytics_redhat_com_v1.SparkClusterListList{})

	if obj == nil {
		return nil, err
	}
	return obj.(*radanalytics_redhat_com_v1.SparkClusterListList), err
}

// Watch returns a watch.Interface that watches the requested sparkClusterLists.
func (c *FakeSparkClusterLists) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(sparkclusterlistsResource, c.ns, opts))

}

// Create takes the representation of a sparkClusterList and creates it.  Returns the server's representation of the sparkClusterList, and an error, if there is any.
func (c *FakeSparkClusterLists) Create(sparkClusterList *radanalytics_redhat_com_v1.SparkClusterList) (result *radanalytics_redhat_com_v1.SparkClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(sparkclusterlistsResource, c.ns, sparkClusterList), &radanalytics_redhat_com_v1.SparkClusterList{})

	if obj == nil {
		return nil, err
	}
	return obj.(*radanalytics_redhat_com_v1.SparkClusterList), err
}

// Update takes the representation of a sparkClusterList and updates it. Returns the server's representation of the sparkClusterList, and an error, if there is any.
func (c *FakeSparkClusterLists) Update(sparkClusterList *radanalytics_redhat_com_v1.SparkClusterList) (result *radanalytics_redhat_com_v1.SparkClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(sparkclusterlistsResource, c.ns, sparkClusterList), &radanalytics_redhat_com_v1.SparkClusterList{})

	if obj == nil {
		return nil, err
	}
	return obj.(*radanalytics_redhat_com_v1.SparkClusterList), err
}

// Delete takes name of the sparkClusterList and deletes it. Returns an error if one occurs.
func (c *FakeSparkClusterLists) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(sparkclusterlistsResource, c.ns, name), &radanalytics_redhat_com_v1.SparkClusterList{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSparkClusterLists) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(sparkclusterlistsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &radanalytics_redhat_com_v1.SparkClusterListList{})
	return err
}

// Patch applies the patch and returns the patched sparkClusterList.
func (c *FakeSparkClusterLists) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *radanalytics_redhat_com_v1.SparkClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(sparkclusterlistsResource, c.ns, name, data, subresources...), &radanalytics_redhat_com_v1.SparkClusterList{})

	if obj == nil {
		return nil, err
	}
	return obj.(*radanalytics_redhat_com_v1.SparkClusterList), err
}
