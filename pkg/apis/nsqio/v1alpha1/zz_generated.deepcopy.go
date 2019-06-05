// +build !ignore_autogenerated

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Meta) DeepCopyInto(out *Meta) {
	*out = *in
	in.LastUpdateTime.DeepCopyInto(&out.LastUpdateTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Meta.
func (in *Meta) DeepCopy() *Meta {
	if in == nil {
		return nil
	}
	out := new(Meta)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqAdmin) DeepCopyInto(out *NsqAdmin) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqAdmin.
func (in *NsqAdmin) DeepCopy() *NsqAdmin {
	if in == nil {
		return nil
	}
	out := new(NsqAdmin)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NsqAdmin) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqAdminList) DeepCopyInto(out *NsqAdminList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NsqAdmin, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqAdminList.
func (in *NsqAdminList) DeepCopy() *NsqAdminList {
	if in == nil {
		return nil
	}
	out := new(NsqAdminList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NsqAdminList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqAdminSpec) DeepCopyInto(out *NsqAdminSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqAdminSpec.
func (in *NsqAdminSpec) DeepCopy() *NsqAdminSpec {
	if in == nil {
		return nil
	}
	out := new(NsqAdminSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqAdminStatus) DeepCopyInto(out *NsqAdminStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqAdminStatus.
func (in *NsqAdminStatus) DeepCopy() *NsqAdminStatus {
	if in == nil {
		return nil
	}
	out := new(NsqAdminStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqLookupd) DeepCopyInto(out *NsqLookupd) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqLookupd.
func (in *NsqLookupd) DeepCopy() *NsqLookupd {
	if in == nil {
		return nil
	}
	out := new(NsqLookupd)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NsqLookupd) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqLookupdList) DeepCopyInto(out *NsqLookupdList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NsqLookupd, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqLookupdList.
func (in *NsqLookupdList) DeepCopy() *NsqLookupdList {
	if in == nil {
		return nil
	}
	out := new(NsqLookupdList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NsqLookupdList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqLookupdSpec) DeepCopyInto(out *NsqLookupdSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqLookupdSpec.
func (in *NsqLookupdSpec) DeepCopy() *NsqLookupdSpec {
	if in == nil {
		return nil
	}
	out := new(NsqLookupdSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqLookupdStatus) DeepCopyInto(out *NsqLookupdStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqLookupdStatus.
func (in *NsqLookupdStatus) DeepCopy() *NsqLookupdStatus {
	if in == nil {
		return nil
	}
	out := new(NsqLookupdStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Nsqd) DeepCopyInto(out *Nsqd) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Nsqd.
func (in *Nsqd) DeepCopy() *Nsqd {
	if in == nil {
		return nil
	}
	out := new(Nsqd)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Nsqd) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqdList) DeepCopyInto(out *NsqdList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Nsqd, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqdList.
func (in *NsqdList) DeepCopy() *NsqdList {
	if in == nil {
		return nil
	}
	out := new(NsqdList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NsqdList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqdScale) DeepCopyInto(out *NsqdScale) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqdScale.
func (in *NsqdScale) DeepCopy() *NsqdScale {
	if in == nil {
		return nil
	}
	out := new(NsqdScale)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NsqdScale) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqdScaleList) DeepCopyInto(out *NsqdScaleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NsqdScale, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqdScaleList.
func (in *NsqdScaleList) DeepCopy() *NsqdScaleList {
	if in == nil {
		return nil
	}
	out := new(NsqdScaleList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NsqdScaleList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqdScaleSpec) DeepCopyInto(out *NsqdScaleSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqdScaleSpec.
func (in *NsqdScaleSpec) DeepCopy() *NsqdScaleSpec {
	if in == nil {
		return nil
	}
	out := new(NsqdScaleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqdScaleStatus) DeepCopyInto(out *NsqdScaleStatus) {
	*out = *in
	in.LastScaleUpTime.DeepCopyInto(&out.LastScaleUpTime)
	in.LastScaleDownTime.DeepCopyInto(&out.LastScaleDownTime)
	if in.Metas != nil {
		in, out := &in.Metas, &out.Metas
		*out = make(map[string][]Meta, len(*in))
		for key, val := range *in {
			var outVal []Meta
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make([]Meta, len(*in))
				for i := range *in {
					(*in)[i].DeepCopyInto(&(*out)[i])
				}
			}
			(*out)[key] = outVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqdScaleStatus.
func (in *NsqdScaleStatus) DeepCopy() *NsqdScaleStatus {
	if in == nil {
		return nil
	}
	out := new(NsqdScaleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqdSpec) DeepCopyInto(out *NsqdSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqdSpec.
func (in *NsqdSpec) DeepCopy() *NsqdSpec {
	if in == nil {
		return nil
	}
	out := new(NsqdSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NsqdStatus) DeepCopyInto(out *NsqdStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NsqdStatus.
func (in *NsqdStatus) DeepCopy() *NsqdStatus {
	if in == nil {
		return nil
	}
	out := new(NsqdStatus)
	in.DeepCopyInto(out)
	return out
}
