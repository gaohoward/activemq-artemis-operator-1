/*
Copyright 2020 The Kubernetes Authors.

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

package v1alpha1

import (
	v1alpha1 "github.com/artemiscloud/activemq-artemis-operator/pkg/apis/broker/v1alpha1"
	scheme "github.com/artemiscloud/activemq-artemis-operator/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ActiveMQArtemisAuthenticationsGetter has a method to return a ActiveMQArtemisAuthenticationInterface.
// A group's client should implement this interface.
type ActiveMQArtemisAuthenticationsGetter interface {
	ActiveMQArtemisAuthentications(namespace string) ActiveMQArtemisAuthenticationInterface
}

// ActiveMQArtemisAuthenticationInterface has methods to work with ActiveMQArtemisAuthentication resources.
type ActiveMQArtemisAuthenticationInterface interface {
	Create(*v1alpha1.ActiveMQArtemisAuthentication) (*v1alpha1.ActiveMQArtemisAuthentication, error)
	Update(*v1alpha1.ActiveMQArtemisAuthentication) (*v1alpha1.ActiveMQArtemisAuthentication, error)
	UpdateStatus(*v1alpha1.ActiveMQArtemisAuthentication) (*v1alpha1.ActiveMQArtemisAuthentication, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ActiveMQArtemisAuthentication, error)
	List(opts v1.ListOptions) (*v1alpha1.ActiveMQArtemisAuthenticationList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ActiveMQArtemisAuthentication, err error)
	ActiveMQArtemisAuthenticationExpansion
}

// activeMQArtemisAuthentications implements ActiveMQArtemisAuthenticationInterface
type activeMQArtemisAuthentications struct {
	client rest.Interface
	ns     string
}

// newActiveMQArtemisAuthentications returns a ActiveMQArtemisAuthentications
func newActiveMQArtemisAuthentications(c *BrokerV1alpha1Client, namespace string) *activeMQArtemisAuthentications {
	return &activeMQArtemisAuthentications{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the activeMQArtemisAuthentication, and returns the corresponding activeMQArtemisAuthentication object, and an error if there is any.
func (c *activeMQArtemisAuthentications) Get(name string, options v1.GetOptions) (result *v1alpha1.ActiveMQArtemisAuthentication, err error) {
	result = &v1alpha1.ActiveMQArtemisAuthentication{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("activemqartemisauthentications").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ActiveMQArtemisAddresses that match those selectors.
func (c *activeMQArtemisAuthentications) List(opts v1.ListOptions) (result *v1alpha1.ActiveMQArtemisAuthenticationList, err error) {
	result = &v1alpha1.ActiveMQArtemisAuthenticationList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("activemqartemisauthentications").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested activeMQArtemisAuthentications.
func (c *activeMQArtemisAuthentications) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("activemqartemisauthentications").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a activeMQArtemisAuthentication and creates it.  Returns the server's representation of the activeMQArtemisAuthentication, and an error, if there is any.
func (c *activeMQArtemisAuthentications) Create(activeMQArtemisAuthentication *v1alpha1.ActiveMQArtemisAuthentication) (result *v1alpha1.ActiveMQArtemisAuthentication, err error) {
	result = &v1alpha1.ActiveMQArtemisAuthentication{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("activemqartemisauthentications").
		Body(activeMQArtemisAuthentication).
		Do().
		Into(result)
	return
}

// Update takes the representation of a activeMQArtemisAuthentication and updates it. Returns the server's representation of the activeMQArtemisAuthentication, and an error, if there is any.
func (c *activeMQArtemisAuthentications) Update(activeMQArtemisAuthentication *v1alpha1.ActiveMQArtemisAuthentication) (result *v1alpha1.ActiveMQArtemisAuthentication, err error) {
	result = &v1alpha1.ActiveMQArtemisAuthentication{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("activemqartemisauthentications").
		Name(activeMQArtemisAuthentication.Name).
		Body(activeMQArtemisAuthentication).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *activeMQArtemisAuthentications) UpdateStatus(activeMQArtemisAuthentication *v1alpha1.ActiveMQArtemisAuthentication) (result *v1alpha1.ActiveMQArtemisAuthentication, err error) {
	result = &v1alpha1.ActiveMQArtemisAuthentication{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("activemqartemisauthentications").
		Name(activeMQArtemisAuthentication.Name).
		SubResource("status").
		Body(activeMQArtemisAuthentication).
		Do().
		Into(result)
	return
}

// Delete takes name of the activeMQArtemisAuthentication and deletes it. Returns an error if one occurs.
func (c *activeMQArtemisAuthentications) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("activemqartemisauthentications").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *activeMQArtemisAuthentications) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("activemqartemisauthentications").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched activeMQArtemisAuthentication.
func (c *activeMQArtemisAuthentications) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ActiveMQArtemisAuthentication, err error) {
	result = &v1alpha1.ActiveMQArtemisAuthentication{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("activemqartemisauthentications").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
