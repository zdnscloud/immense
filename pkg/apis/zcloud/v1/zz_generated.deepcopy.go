// +build !ignore_autogenerated

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Capacity) DeepCopyInto(out *Capacity) {
	*out = *in
	out.Total = in.Total
	if in.Instances != nil {
		in, out := &in.Instances, &out.Instances
		*out = make([]Instance, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Capacity.
func (in *Capacity) DeepCopy() *Capacity {
	if in == nil {
		return nil
	}
	out := new(Capacity)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Cluster) DeepCopyInto(out *Cluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Cluster.
func (in *Cluster) DeepCopy() *Cluster {
	if in == nil {
		return nil
	}
	out := new(Cluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Cluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterList) DeepCopyInto(out *ClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Cluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterList.
func (in *ClusterList) DeepCopy() *ClusterList {
	if in == nil {
		return nil
	}
	out := new(ClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterSpec) DeepCopyInto(out *ClusterSpec) {
	*out = *in
	if in.Hosts != nil {
		in, out := &in.Hosts, &out.Hosts
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterSpec.
func (in *ClusterSpec) DeepCopy() *ClusterSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterStatus) DeepCopyInto(out *ClusterStatus) {
	*out = *in
	in.Capacity.DeepCopyInto(&out.Capacity)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterStatus.
func (in *ClusterStatus) DeepCopy() *ClusterStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Instance) DeepCopyInto(out *Instance) {
	*out = *in
	out.Info = in.Info
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Instance.
func (in *Instance) DeepCopy() *Instance {
	if in == nil {
		return nil
	}
	out := new(Instance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Size) DeepCopyInto(out *Size) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Size.
func (in *Size) DeepCopy() *Size {
	if in == nil {
		return nil
	}
	out := new(Size)
	in.DeepCopyInto(out)
	return out
}

func (in *Iscsi) DeepCopyInto(out *Iscsi) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Iscsi.
func (in *Iscsi) DeepCopy() *Iscsi {
	if in == nil {
		return nil
	}
	out := new(Iscsi)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Iscsi) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *IscsiList) DeepCopyInto(out *IscsiList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Iscsi, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IscsiList.
func (in *IscsiList) DeepCopy() *IscsiList {
	if in == nil {
		return nil
	}
	out := new(IscsiList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IscsiList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *IscsiSpec) DeepCopyInto(out *IscsiSpec) {
	*out = *in
	if in.Initiators != nil {
		in, out := &in.Initiators, &out.Initiators
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IscsiSpec.
func (in *IscsiSpec) DeepCopy() *IscsiSpec {
	if in == nil {
		return nil
	}
	out := new(IscsiSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IscsiStatus) DeepCopyInto(out *IscsiStatus) {
	*out = *in
	in.Capacity.DeepCopyInto(&out.Capacity)
	return
}

func (in *IscsiStatus) DeepCopy() *IscsiStatus {
	if in == nil {
		return nil
	}
	out := new(IscsiStatus)
	in.DeepCopyInto(out)
	return out
}

func (in *Nfs) DeepCopyInto(out *Nfs) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Nfs.
func (in *Nfs) DeepCopy() *Nfs {
	if in == nil {
		return nil
	}
	out := new(Nfs)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Nfs) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *NfsList) DeepCopyInto(out *NfsList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Nfs, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsList.
func (in *NfsList) DeepCopy() *NfsList {
	if in == nil {
		return nil
	}
	out := new(NfsList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NfsList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *NfsSpec) DeepCopyInto(out *NfsSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsSpec.
func (in *NfsSpec) DeepCopy() *NfsSpec {
	if in == nil {
		return nil
	}
	out := new(NfsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NfsStatus) DeepCopyInto(out *NfsStatus) {
	*out = *in
	in.Capacity.DeepCopyInto(&out.Capacity)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NfsStatus.
func (in *NfsStatus) DeepCopy() *NfsStatus {
	if in == nil {
		return nil
	}
	out := new(NfsStatus)
	in.DeepCopyInto(out)
	return out
}
