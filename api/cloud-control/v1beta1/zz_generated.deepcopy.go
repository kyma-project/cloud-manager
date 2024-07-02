//go:build !ignore_autogenerated

/*
Copyright 2023.

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

package v1beta1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AwsNetwork) DeepCopyInto(out *AwsNetwork) {
	*out = *in
	out.VPC = in.VPC
	if in.Zones != nil {
		in, out := &in.Zones, &out.Zones
		*out = make([]AwsZone, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AwsNetwork.
func (in *AwsNetwork) DeepCopy() *AwsNetwork {
	if in == nil {
		return nil
	}
	out := new(AwsNetwork)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AwsScope) DeepCopyInto(out *AwsScope) {
	*out = *in
	in.Network.DeepCopyInto(&out.Network)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AwsScope.
func (in *AwsScope) DeepCopy() *AwsScope {
	if in == nil {
		return nil
	}
	out := new(AwsScope)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AwsVPC) DeepCopyInto(out *AwsVPC) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AwsVPC.
func (in *AwsVPC) DeepCopy() *AwsVPC {
	if in == nil {
		return nil
	}
	out := new(AwsVPC)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AwsVpcPeering) DeepCopyInto(out *AwsVpcPeering) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AwsVpcPeering.
func (in *AwsVpcPeering) DeepCopy() *AwsVpcPeering {
	if in == nil {
		return nil
	}
	out := new(AwsVpcPeering)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AwsZone) DeepCopyInto(out *AwsZone) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AwsZone.
func (in *AwsZone) DeepCopy() *AwsZone {
	if in == nil {
		return nil
	}
	out := new(AwsZone)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureScope) DeepCopyInto(out *AzureScope) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureScope.
func (in *AzureScope) DeepCopy() *AzureScope {
	if in == nil {
		return nil
	}
	out := new(AzureScope)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureVpcPeering) DeepCopyInto(out *AzureVpcPeering) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureVpcPeering.
func (in *AzureVpcPeering) DeepCopy() *AzureVpcPeering {
	if in == nil {
		return nil
	}
	out := new(AzureVpcPeering)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GcpNetwork) DeepCopyInto(out *GcpNetwork) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GcpNetwork.
func (in *GcpNetwork) DeepCopy() *GcpNetwork {
	if in == nil {
		return nil
	}
	out := new(GcpNetwork)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GcpScope) DeepCopyInto(out *GcpScope) {
	*out = *in
	out.Network = in.Network
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GcpScope.
func (in *GcpScope) DeepCopy() *GcpScope {
	if in == nil {
		return nil
	}
	out := new(GcpScope)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GcpVpcPeering) DeepCopyInto(out *GcpVpcPeering) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GcpVpcPeering.
func (in *GcpVpcPeering) DeepCopy() *GcpVpcPeering {
	if in == nil {
		return nil
	}
	out := new(GcpVpcPeering)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IpRange) DeepCopyInto(out *IpRange) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IpRange.
func (in *IpRange) DeepCopy() *IpRange {
	if in == nil {
		return nil
	}
	out := new(IpRange)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IpRange) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IpRangeAws) DeepCopyInto(out *IpRangeAws) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IpRangeAws.
func (in *IpRangeAws) DeepCopy() *IpRangeAws {
	if in == nil {
		return nil
	}
	out := new(IpRangeAws)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IpRangeAzure) DeepCopyInto(out *IpRangeAzure) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IpRangeAzure.
func (in *IpRangeAzure) DeepCopy() *IpRangeAzure {
	if in == nil {
		return nil
	}
	out := new(IpRangeAzure)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IpRangeGcp) DeepCopyInto(out *IpRangeGcp) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IpRangeGcp.
func (in *IpRangeGcp) DeepCopy() *IpRangeGcp {
	if in == nil {
		return nil
	}
	out := new(IpRangeGcp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IpRangeList) DeepCopyInto(out *IpRangeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]IpRange, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IpRangeList.
func (in *IpRangeList) DeepCopy() *IpRangeList {
	if in == nil {
		return nil
	}
	out := new(IpRangeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IpRangeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IpRangeOptions) DeepCopyInto(out *IpRangeOptions) {
	*out = *in
	if in.Gcp != nil {
		in, out := &in.Gcp, &out.Gcp
		*out = new(IpRangeGcp)
		**out = **in
	}
	if in.Azure != nil {
		in, out := &in.Azure, &out.Azure
		*out = new(IpRangeAzure)
		**out = **in
	}
	if in.Aws != nil {
		in, out := &in.Aws, &out.Aws
		*out = new(IpRangeAws)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IpRangeOptions.
func (in *IpRangeOptions) DeepCopy() *IpRangeOptions {
	if in == nil {
		return nil
	}
	out := new(IpRangeOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IpRangeRef) DeepCopyInto(out *IpRangeRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IpRangeRef.
func (in *IpRangeRef) DeepCopy() *IpRangeRef {
	if in == nil {
		return nil
	}
	out := new(IpRangeRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IpRangeSpec) DeepCopyInto(out *IpRangeSpec) {
	*out = *in
	out.RemoteRef = in.RemoteRef
	out.Scope = in.Scope
	in.Options.DeepCopyInto(&out.Options)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IpRangeSpec.
func (in *IpRangeSpec) DeepCopy() *IpRangeSpec {
	if in == nil {
		return nil
	}
	out := new(IpRangeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IpRangeStatus) DeepCopyInto(out *IpRangeStatus) {
	*out = *in
	if in.Ranges != nil {
		in, out := &in.Ranges, &out.Ranges
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Subnets != nil {
		in, out := &in.Subnets, &out.Subnets
		*out = make(IpRangeSubnets, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IpRangeStatus.
func (in *IpRangeStatus) DeepCopy() *IpRangeStatus {
	if in == nil {
		return nil
	}
	out := new(IpRangeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IpRangeSubnet) DeepCopyInto(out *IpRangeSubnet) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IpRangeSubnet.
func (in *IpRangeSubnet) DeepCopy() *IpRangeSubnet {
	if in == nil {
		return nil
	}
	out := new(IpRangeSubnet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in IpRangeSubnets) DeepCopyInto(out *IpRangeSubnets) {
	{
		in := &in
		*out = make(IpRangeSubnets, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IpRangeSubnets.
func (in IpRangeSubnets) DeepCopy() IpRangeSubnets {
	if in == nil {
		return nil
	}
	out := new(IpRangeSubnets)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfsInstance) DeepCopyInto(out *NfsInstance) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsInstance.
func (in *NfsInstance) DeepCopy() *NfsInstance {
	if in == nil {
		return nil
	}
	out := new(NfsInstance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NfsInstance) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfsInstanceAws) DeepCopyInto(out *NfsInstanceAws) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsInstanceAws.
func (in *NfsInstanceAws) DeepCopy() *NfsInstanceAws {
	if in == nil {
		return nil
	}
	out := new(NfsInstanceAws)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfsInstanceAzure) DeepCopyInto(out *NfsInstanceAzure) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsInstanceAzure.
func (in *NfsInstanceAzure) DeepCopy() *NfsInstanceAzure {
	if in == nil {
		return nil
	}
	out := new(NfsInstanceAzure)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfsInstanceGcp) DeepCopyInto(out *NfsInstanceGcp) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsInstanceGcp.
func (in *NfsInstanceGcp) DeepCopy() *NfsInstanceGcp {
	if in == nil {
		return nil
	}
	out := new(NfsInstanceGcp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfsInstanceInfo) DeepCopyInto(out *NfsInstanceInfo) {
	*out = *in
	if in.Gcp != nil {
		in, out := &in.Gcp, &out.Gcp
		*out = new(NfsInstanceGcp)
		**out = **in
	}
	if in.Azure != nil {
		in, out := &in.Azure, &out.Azure
		*out = new(NfsInstanceAzure)
		**out = **in
	}
	if in.Aws != nil {
		in, out := &in.Aws, &out.Aws
		*out = new(NfsInstanceAws)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsInstanceInfo.
func (in *NfsInstanceInfo) DeepCopy() *NfsInstanceInfo {
	if in == nil {
		return nil
	}
	out := new(NfsInstanceInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfsInstanceList) DeepCopyInto(out *NfsInstanceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NfsInstance, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsInstanceList.
func (in *NfsInstanceList) DeepCopy() *NfsInstanceList {
	if in == nil {
		return nil
	}
	out := new(NfsInstanceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NfsInstanceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfsInstanceSpec) DeepCopyInto(out *NfsInstanceSpec) {
	*out = *in
	out.RemoteRef = in.RemoteRef
	out.IpRange = in.IpRange
	out.Scope = in.Scope
	in.Instance.DeepCopyInto(&out.Instance)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsInstanceSpec.
func (in *NfsInstanceSpec) DeepCopy() *NfsInstanceSpec {
	if in == nil {
		return nil
	}
	out := new(NfsInstanceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfsInstanceStatus) DeepCopyInto(out *NfsInstanceStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Hosts != nil {
		in, out := &in.Hosts, &out.Hosts
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsInstanceStatus.
func (in *NfsInstanceStatus) DeepCopy() *NfsInstanceStatus {
	if in == nil {
		return nil
	}
	out := new(NfsInstanceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfsOptionsGcp) DeepCopyInto(out *NfsOptionsGcp) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsOptionsGcp.
func (in *NfsOptionsGcp) DeepCopy() *NfsOptionsGcp {
	if in == nil {
		return nil
	}
	out := new(NfsOptionsGcp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisInstance) DeepCopyInto(out *RedisInstance) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisInstance.
func (in *RedisInstance) DeepCopy() *RedisInstance {
	if in == nil {
		return nil
	}
	out := new(RedisInstance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RedisInstance) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisInstanceAws) DeepCopyInto(out *RedisInstanceAws) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisInstanceAws.
func (in *RedisInstanceAws) DeepCopy() *RedisInstanceAws {
	if in == nil {
		return nil
	}
	out := new(RedisInstanceAws)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisInstanceAzure) DeepCopyInto(out *RedisInstanceAzure) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisInstanceAzure.
func (in *RedisInstanceAzure) DeepCopy() *RedisInstanceAzure {
	if in == nil {
		return nil
	}
	out := new(RedisInstanceAzure)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisInstanceGcp) DeepCopyInto(out *RedisInstanceGcp) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisInstanceGcp.
func (in *RedisInstanceGcp) DeepCopy() *RedisInstanceGcp {
	if in == nil {
		return nil
	}
	out := new(RedisInstanceGcp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisInstanceInfo) DeepCopyInto(out *RedisInstanceInfo) {
	*out = *in
	if in.Gcp != nil {
		in, out := &in.Gcp, &out.Gcp
		*out = new(RedisInstanceGcp)
		**out = **in
	}
	if in.Azure != nil {
		in, out := &in.Azure, &out.Azure
		*out = new(RedisInstanceAzure)
		**out = **in
	}
	if in.Aws != nil {
		in, out := &in.Aws, &out.Aws
		*out = new(RedisInstanceAws)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisInstanceInfo.
func (in *RedisInstanceInfo) DeepCopy() *RedisInstanceInfo {
	if in == nil {
		return nil
	}
	out := new(RedisInstanceInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisInstanceList) DeepCopyInto(out *RedisInstanceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RedisInstance, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisInstanceList.
func (in *RedisInstanceList) DeepCopy() *RedisInstanceList {
	if in == nil {
		return nil
	}
	out := new(RedisInstanceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RedisInstanceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisInstanceSpec) DeepCopyInto(out *RedisInstanceSpec) {
	*out = *in
	out.RemoteRef = in.RemoteRef
	out.IpRange = in.IpRange
	out.Scope = in.Scope
	in.Instance.DeepCopyInto(&out.Instance)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisInstanceSpec.
func (in *RedisInstanceSpec) DeepCopy() *RedisInstanceSpec {
	if in == nil {
		return nil
	}
	out := new(RedisInstanceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisInstanceStatus) DeepCopyInto(out *RedisInstanceStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisInstanceStatus.
func (in *RedisInstanceStatus) DeepCopy() *RedisInstanceStatus {
	if in == nil {
		return nil
	}
	out := new(RedisInstanceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RemoteRef) DeepCopyInto(out *RemoteRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RemoteRef.
func (in *RemoteRef) DeepCopy() *RemoteRef {
	if in == nil {
		return nil
	}
	out := new(RemoteRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Scope) DeepCopyInto(out *Scope) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Scope.
func (in *Scope) DeepCopy() *Scope {
	if in == nil {
		return nil
	}
	out := new(Scope)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Scope) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScopeInfo) DeepCopyInto(out *ScopeInfo) {
	*out = *in
	if in.Gcp != nil {
		in, out := &in.Gcp, &out.Gcp
		*out = new(GcpScope)
		**out = **in
	}
	if in.Azure != nil {
		in, out := &in.Azure, &out.Azure
		*out = new(AzureScope)
		**out = **in
	}
	if in.Aws != nil {
		in, out := &in.Aws, &out.Aws
		*out = new(AwsScope)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScopeInfo.
func (in *ScopeInfo) DeepCopy() *ScopeInfo {
	if in == nil {
		return nil
	}
	out := new(ScopeInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScopeList) DeepCopyInto(out *ScopeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Scope, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScopeList.
func (in *ScopeList) DeepCopy() *ScopeList {
	if in == nil {
		return nil
	}
	out := new(ScopeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ScopeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScopeRef) DeepCopyInto(out *ScopeRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScopeRef.
func (in *ScopeRef) DeepCopy() *ScopeRef {
	if in == nil {
		return nil
	}
	out := new(ScopeRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScopeSpec) DeepCopyInto(out *ScopeSpec) {
	*out = *in
	in.Scope.DeepCopyInto(&out.Scope)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScopeSpec.
func (in *ScopeSpec) DeepCopy() *ScopeSpec {
	if in == nil {
		return nil
	}
	out := new(ScopeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScopeStatus) DeepCopyInto(out *ScopeStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.GcpOperations != nil {
		in, out := &in.GcpOperations, &out.GcpOperations
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScopeStatus.
func (in *ScopeStatus) DeepCopy() *ScopeStatus {
	if in == nil {
		return nil
	}
	out := new(ScopeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VpcPeering) DeepCopyInto(out *VpcPeering) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VpcPeering.
func (in *VpcPeering) DeepCopy() *VpcPeering {
	if in == nil {
		return nil
	}
	out := new(VpcPeering)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VpcPeering) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VpcPeeringInfo) DeepCopyInto(out *VpcPeeringInfo) {
	*out = *in
	if in.Gcp != nil {
		in, out := &in.Gcp, &out.Gcp
		*out = new(GcpVpcPeering)
		**out = **in
	}
	if in.Azure != nil {
		in, out := &in.Azure, &out.Azure
		*out = new(AzureVpcPeering)
		**out = **in
	}
	if in.Aws != nil {
		in, out := &in.Aws, &out.Aws
		*out = new(AwsVpcPeering)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VpcPeeringInfo.
func (in *VpcPeeringInfo) DeepCopy() *VpcPeeringInfo {
	if in == nil {
		return nil
	}
	out := new(VpcPeeringInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VpcPeeringList) DeepCopyInto(out *VpcPeeringList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VpcPeering, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VpcPeeringList.
func (in *VpcPeeringList) DeepCopy() *VpcPeeringList {
	if in == nil {
		return nil
	}
	out := new(VpcPeeringList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VpcPeeringList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VpcPeeringSpec) DeepCopyInto(out *VpcPeeringSpec) {
	*out = *in
	out.RemoteRef = in.RemoteRef
	out.Scope = in.Scope
	in.VpcPeering.DeepCopyInto(&out.VpcPeering)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VpcPeeringSpec.
func (in *VpcPeeringSpec) DeepCopy() *VpcPeeringSpec {
	if in == nil {
		return nil
	}
	out := new(VpcPeeringSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VpcPeeringStatus) DeepCopyInto(out *VpcPeeringStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VpcPeeringStatus.
func (in *VpcPeeringStatus) DeepCopy() *VpcPeeringStatus {
	if in == nil {
		return nil
	}
	out := new(VpcPeeringStatus)
	in.DeepCopyInto(out)
	return out
}
