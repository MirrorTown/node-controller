/*
Copyright The Kubernetes Authors.

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
	"context"
	v1alpha1 "node-controller/api/virtulmachinecontroller/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVirtulMachines implements VirtulMachineInterface
type FakeVirtulMachines struct {
	Fake *FakeNodecontrollerV1alpha1
	ns   string
}

var virtulmachinesResource = schema.GroupVersionResource{Group: "nodecontroller.k8s.io", Version: "v1alpha1", Resource: "virtulmachines"}

var virtulmachinesKind = schema.GroupVersionKind{Group: "nodecontroller.k8s.io", Version: "v1alpha1", Kind: "VirtulMachine"}

// Get takes name of the virtulMachine, and returns the corresponding virtulMachine object, and an error if there is any.
func (c *FakeVirtulMachines) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.VirtulMachine, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(virtulmachinesResource, c.ns, name), &v1alpha1.VirtulMachine{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VirtulMachine), err
}

// List takes label and field selectors, and returns the list of VirtulMachines that match those selectors.
func (c *FakeVirtulMachines) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.VirtulMachineList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(virtulmachinesResource, virtulmachinesKind, c.ns, opts), &v1alpha1.VirtulMachineList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.VirtulMachineList{ListMeta: obj.(*v1alpha1.VirtulMachineList).ListMeta}
	for _, item := range obj.(*v1alpha1.VirtulMachineList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested virtulMachines.
func (c *FakeVirtulMachines) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(virtulmachinesResource, c.ns, opts))

}

// Create takes the representation of a virtulMachine and creates it.  Returns the server's representation of the virtulMachine, and an error, if there is any.
func (c *FakeVirtulMachines) Create(ctx context.Context, virtulMachine *v1alpha1.VirtulMachine, opts v1.CreateOptions) (result *v1alpha1.VirtulMachine, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(virtulmachinesResource, c.ns, virtulMachine), &v1alpha1.VirtulMachine{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VirtulMachine), err
}

// Update takes the representation of a virtulMachine and updates it. Returns the server's representation of the virtulMachine, and an error, if there is any.
func (c *FakeVirtulMachines) Update(ctx context.Context, virtulMachine *v1alpha1.VirtulMachine, opts v1.UpdateOptions) (result *v1alpha1.VirtulMachine, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(virtulmachinesResource, c.ns, virtulMachine), &v1alpha1.VirtulMachine{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VirtulMachine), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeVirtulMachines) UpdateStatus(ctx context.Context, virtulMachine *v1alpha1.VirtulMachine, opts v1.UpdateOptions) (*v1alpha1.VirtulMachine, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(virtulmachinesResource, "status", c.ns, virtulMachine), &v1alpha1.VirtulMachine{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VirtulMachine), err
}

// Delete takes name of the virtulMachine and deletes it. Returns an error if one occurs.
func (c *FakeVirtulMachines) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(virtulmachinesResource, c.ns, name), &v1alpha1.VirtulMachine{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVirtulMachines) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(virtulmachinesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.VirtulMachineList{})
	return err
}

// Patch applies the patch and returns the patched virtulMachine.
func (c *FakeVirtulMachines) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.VirtulMachine, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(virtulmachinesResource, c.ns, name, pt, data, subresources...), &v1alpha1.VirtulMachine{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VirtulMachine), err
}