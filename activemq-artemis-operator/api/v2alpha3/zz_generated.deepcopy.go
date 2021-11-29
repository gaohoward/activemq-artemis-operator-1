//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2021.

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

// Code generated by controller-gen. DO NOT EDIT.

package v2alpha3

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ActiveMQArtemisAddress) DeepCopyInto(out *ActiveMQArtemisAddress) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ActiveMQArtemisAddress.
func (in *ActiveMQArtemisAddress) DeepCopy() *ActiveMQArtemisAddress {
	if in == nil {
		return nil
	}
	out := new(ActiveMQArtemisAddress)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ActiveMQArtemisAddress) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ActiveMQArtemisAddressList) DeepCopyInto(out *ActiveMQArtemisAddressList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ActiveMQArtemisAddress, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ActiveMQArtemisAddressList.
func (in *ActiveMQArtemisAddressList) DeepCopy() *ActiveMQArtemisAddressList {
	if in == nil {
		return nil
	}
	out := new(ActiveMQArtemisAddressList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ActiveMQArtemisAddressList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ActiveMQArtemisAddressSpec) DeepCopyInto(out *ActiveMQArtemisAddressSpec) {
	*out = *in
	if in.QueueName != nil {
		in, out := &in.QueueName, &out.QueueName
		*out = new(string)
		**out = **in
	}
	if in.RoutingType != nil {
		in, out := &in.RoutingType, &out.RoutingType
		*out = new(string)
		**out = **in
	}
	if in.User != nil {
		in, out := &in.User, &out.User
		*out = new(string)
		**out = **in
	}
	if in.Password != nil {
		in, out := &in.Password, &out.Password
		*out = new(string)
		**out = **in
	}
	if in.QueueConfiguration != nil {
		in, out := &in.QueueConfiguration, &out.QueueConfiguration
		*out = new(QueueConfigurationType)
		(*in).DeepCopyInto(*out)
	}
	if in.ApplyToCrNames != nil {
		in, out := &in.ApplyToCrNames, &out.ApplyToCrNames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ActiveMQArtemisAddressSpec.
func (in *ActiveMQArtemisAddressSpec) DeepCopy() *ActiveMQArtemisAddressSpec {
	if in == nil {
		return nil
	}
	out := new(ActiveMQArtemisAddressSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ActiveMQArtemisAddressStatus) DeepCopyInto(out *ActiveMQArtemisAddressStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ActiveMQArtemisAddressStatus.
func (in *ActiveMQArtemisAddressStatus) DeepCopy() *ActiveMQArtemisAddressStatus {
	if in == nil {
		return nil
	}
	out := new(ActiveMQArtemisAddressStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QueueConfigurationType) DeepCopyInto(out *QueueConfigurationType) {
	*out = *in
	if in.IgnoreIfExists != nil {
		in, out := &in.IgnoreIfExists, &out.IgnoreIfExists
		*out = new(bool)
		**out = **in
	}
	if in.RoutingType != nil {
		in, out := &in.RoutingType, &out.RoutingType
		*out = new(string)
		**out = **in
	}
	if in.FilterString != nil {
		in, out := &in.FilterString, &out.FilterString
		*out = new(string)
		**out = **in
	}
	if in.Durable != nil {
		in, out := &in.Durable, &out.Durable
		*out = new(bool)
		**out = **in
	}
	if in.User != nil {
		in, out := &in.User, &out.User
		*out = new(string)
		**out = **in
	}
	if in.MaxConsumers != nil {
		in, out := &in.MaxConsumers, &out.MaxConsumers
		*out = new(int32)
		**out = **in
	}
	if in.Exclusive != nil {
		in, out := &in.Exclusive, &out.Exclusive
		*out = new(bool)
		**out = **in
	}
	if in.GroupRebalance != nil {
		in, out := &in.GroupRebalance, &out.GroupRebalance
		*out = new(bool)
		**out = **in
	}
	if in.GroupRebalancePauseDispatch != nil {
		in, out := &in.GroupRebalancePauseDispatch, &out.GroupRebalancePauseDispatch
		*out = new(bool)
		**out = **in
	}
	if in.GroupBuckets != nil {
		in, out := &in.GroupBuckets, &out.GroupBuckets
		*out = new(int32)
		**out = **in
	}
	if in.GroupFirstKey != nil {
		in, out := &in.GroupFirstKey, &out.GroupFirstKey
		*out = new(string)
		**out = **in
	}
	if in.LastValue != nil {
		in, out := &in.LastValue, &out.LastValue
		*out = new(bool)
		**out = **in
	}
	if in.LastValueKey != nil {
		in, out := &in.LastValueKey, &out.LastValueKey
		*out = new(string)
		**out = **in
	}
	if in.NonDestructive != nil {
		in, out := &in.NonDestructive, &out.NonDestructive
		*out = new(bool)
		**out = **in
	}
	if in.PurgeOnNoConsumers != nil {
		in, out := &in.PurgeOnNoConsumers, &out.PurgeOnNoConsumers
		*out = new(bool)
		**out = **in
	}
	if in.Enabled != nil {
		in, out := &in.Enabled, &out.Enabled
		*out = new(bool)
		**out = **in
	}
	if in.ConsumersBeforeDispatch != nil {
		in, out := &in.ConsumersBeforeDispatch, &out.ConsumersBeforeDispatch
		*out = new(int32)
		**out = **in
	}
	if in.DelayBeforeDispatch != nil {
		in, out := &in.DelayBeforeDispatch, &out.DelayBeforeDispatch
		*out = new(int64)
		**out = **in
	}
	if in.ConsumerPriority != nil {
		in, out := &in.ConsumerPriority, &out.ConsumerPriority
		*out = new(int32)
		**out = **in
	}
	if in.AutoDelete != nil {
		in, out := &in.AutoDelete, &out.AutoDelete
		*out = new(bool)
		**out = **in
	}
	if in.AutoDeleteDelay != nil {
		in, out := &in.AutoDeleteDelay, &out.AutoDeleteDelay
		*out = new(int64)
		**out = **in
	}
	if in.AutoDeleteMessageCount != nil {
		in, out := &in.AutoDeleteMessageCount, &out.AutoDeleteMessageCount
		*out = new(int64)
		**out = **in
	}
	if in.RingSize != nil {
		in, out := &in.RingSize, &out.RingSize
		*out = new(int64)
		**out = **in
	}
	if in.ConfigurationManaged != nil {
		in, out := &in.ConfigurationManaged, &out.ConfigurationManaged
		*out = new(bool)
		**out = **in
	}
	if in.Temporary != nil {
		in, out := &in.Temporary, &out.Temporary
		*out = new(bool)
		**out = **in
	}
	if in.AutoCreateAddress != nil {
		in, out := &in.AutoCreateAddress, &out.AutoCreateAddress
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QueueConfigurationType.
func (in *QueueConfigurationType) DeepCopy() *QueueConfigurationType {
	if in == nil {
		return nil
	}
	out := new(QueueConfigurationType)
	in.DeepCopyInto(out)
	return out
}
