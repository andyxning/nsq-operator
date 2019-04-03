/*
Copyright 2019 The NSQ-Operator Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/andyxning/nsq-operator/pkg/apis/nsqio/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeNsqLookupds implements NsqLookupdInterface
type FakeNsqLookupds struct {
	Fake *FakeNsqV1alpha1
	ns   string
}

var nsqlookupdsResource = schema.GroupVersionResource{Group: "nsq.io", Version: "v1alpha1", Resource: "nsqlookupds"}

var nsqlookupdsKind = schema.GroupVersionKind{Group: "nsq.io", Version: "v1alpha1", Kind: "NsqLookupd"}

// Get takes name of the nsqLookupd, and returns the corresponding nsqLookupd object, and an error if there is any.
func (c *FakeNsqLookupds) Get(name string, options v1.GetOptions) (result *v1alpha1.NsqLookupd, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(nsqlookupdsResource, c.ns, name), &v1alpha1.NsqLookupd{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.NsqLookupd), err
}

// List takes label and field selectors, and returns the list of NsqLookupds that match those selectors.
func (c *FakeNsqLookupds) List(opts v1.ListOptions) (result *v1alpha1.NsqLookupdList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(nsqlookupdsResource, nsqlookupdsKind, c.ns, opts), &v1alpha1.NsqLookupdList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.NsqLookupdList{ListMeta: obj.(*v1alpha1.NsqLookupdList).ListMeta}
	for _, item := range obj.(*v1alpha1.NsqLookupdList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested nsqLookupds.
func (c *FakeNsqLookupds) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(nsqlookupdsResource, c.ns, opts))

}

// Create takes the representation of a nsqLookupd and creates it.  Returns the server's representation of the nsqLookupd, and an error, if there is any.
func (c *FakeNsqLookupds) Create(nsqLookupd *v1alpha1.NsqLookupd) (result *v1alpha1.NsqLookupd, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(nsqlookupdsResource, c.ns, nsqLookupd), &v1alpha1.NsqLookupd{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.NsqLookupd), err
}

// Update takes the representation of a nsqLookupd and updates it. Returns the server's representation of the nsqLookupd, and an error, if there is any.
func (c *FakeNsqLookupds) Update(nsqLookupd *v1alpha1.NsqLookupd) (result *v1alpha1.NsqLookupd, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(nsqlookupdsResource, c.ns, nsqLookupd), &v1alpha1.NsqLookupd{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.NsqLookupd), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeNsqLookupds) UpdateStatus(nsqLookupd *v1alpha1.NsqLookupd) (*v1alpha1.NsqLookupd, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(nsqlookupdsResource, "status", c.ns, nsqLookupd), &v1alpha1.NsqLookupd{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.NsqLookupd), err
}

// Delete takes name of the nsqLookupd and deletes it. Returns an error if one occurs.
func (c *FakeNsqLookupds) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(nsqlookupdsResource, c.ns, name), &v1alpha1.NsqLookupd{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeNsqLookupds) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(nsqlookupdsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.NsqLookupdList{})
	return err
}

// Patch applies the patch and returns the patched nsqLookupd.
func (c *FakeNsqLookupds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.NsqLookupd, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(nsqlookupdsResource, c.ns, name, pt, data, subresources...), &v1alpha1.NsqLookupd{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.NsqLookupd), err
}
