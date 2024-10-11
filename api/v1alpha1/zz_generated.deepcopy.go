//go:build !ignore_autogenerated

// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MountpointClaim) DeepCopyInto(out *MountpointClaim) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MountpointClaim.
func (in *MountpointClaim) DeepCopy() *MountpointClaim {
	if in == nil {
		return nil
	}
	out := new(MountpointClaim)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MountpointClaim) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MountpointClaimList) DeepCopyInto(out *MountpointClaimList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MountpointClaim, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MountpointClaimList.
func (in *MountpointClaimList) DeepCopy() *MountpointClaimList {
	if in == nil {
		return nil
	}
	out := new(MountpointClaimList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MountpointClaimList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MountpointClaimSpec) DeepCopyInto(out *MountpointClaimSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MountpointClaimSpec.
func (in *MountpointClaimSpec) DeepCopy() *MountpointClaimSpec {
	if in == nil {
		return nil
	}
	out := new(MountpointClaimSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MountpointClaimStatus) DeepCopyInto(out *MountpointClaimStatus) {
	*out = *in
	if in.MountpointPodID != nil {
		in, out := &in.MountpointPodID, &out.MountpointPodID
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MountpointClaimStatus.
func (in *MountpointClaimStatus) DeepCopy() *MountpointClaimStatus {
	if in == nil {
		return nil
	}
	out := new(MountpointClaimStatus)
	in.DeepCopyInto(out)
	return out
}
