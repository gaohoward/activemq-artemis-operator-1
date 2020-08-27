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

package versioned

import (
	brokerv2alpha1 "github.com/artemiscloud/activemq-artemis-operator/pkg/client/clientset/versioned/typed/broker/v2alpha1"
	brokerv2alpha2 "github.com/artemiscloud/activemq-artemis-operator/pkg/client/clientset/versioned/typed/broker/v2alpha2"
	brokerv3alpha1 "github.com/artemiscloud/activemq-artemis-operator/pkg/client/clientset/versioned/typed/broker/v3alpha1"
	glog "github.com/golang/glog"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	BrokerV2alpha1() brokerv2alpha1.BrokerV2alpha1Interface
	BrokerV2alpha2() brokerv2alpha2.BrokerV2alpha2Interface
	BrokerV3alpha1() brokerv3alpha1.BrokerV3alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Broker() brokerv3alpha1.BrokerV3alpha1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	brokerV2alpha1 *brokerv2alpha1.BrokerV2alpha1Client
	brokerV2alpha2 *brokerv2alpha2.BrokerV2alpha2Client
	brokerV3alpha1 *brokerv3alpha1.BrokerV3alpha1Client
}

// BrokerV2alpha1 retrieves the BrokerV2alpha1Client
func (c *Clientset) BrokerV2alpha1() brokerv2alpha1.BrokerV2alpha1Interface {
	return c.brokerV2alpha1
}

// BrokerV2alpha2 retrieves the BrokerV2alpha2Client
func (c *Clientset) BrokerV2alpha2() brokerv2alpha2.BrokerV2alpha2Interface {
	return c.brokerV2alpha2
}

// BrokerV3alpha1 retrieves the BrokerV3alpha1Client
func (c *Clientset) BrokerV3alpha1() brokerv3alpha1.BrokerV3alpha1Interface {
	return c.brokerV3alpha1
}

// Deprecated: Broker retrieves the default version of BrokerClient.
// Please explicitly pick a version.
func (c *Clientset) Broker() brokerv3alpha1.BrokerV3alpha1Interface {
	return c.brokerV3alpha1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.brokerV2alpha1, err = brokerv2alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.brokerV2alpha2, err = brokerv2alpha2.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.brokerV3alpha1, err = brokerv3alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		glog.Errorf("failed to create the DiscoveryClient: %v", err)
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.brokerV2alpha1 = brokerv2alpha1.NewForConfigOrDie(c)
	cs.brokerV2alpha2 = brokerv2alpha2.NewForConfigOrDie(c)
	cs.brokerV3alpha1 = brokerv3alpha1.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.brokerV2alpha1 = brokerv2alpha1.New(c)
	cs.brokerV2alpha2 = brokerv2alpha2.New(c)
	cs.brokerV3alpha1 = brokerv3alpha1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
