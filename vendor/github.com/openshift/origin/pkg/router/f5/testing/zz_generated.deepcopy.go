// +build !ignore_autogenerated_openshift

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package testing

import (
	unsafe "unsafe"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Datagroup) DeepCopyInto(out *Datagroup) {
	{
		in := (*map[string]string)(unsafe.Pointer(in))
		out := (*map[string]string)(unsafe.Pointer(out))
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Datagroup.
func (in *Datagroup) DeepCopy() *Datagroup {
	if in == nil {
		return nil
	}
	out := new(Datagroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IRule) DeepCopyInto(out *IRule) {
	{
		in := (*string)(unsafe.Pointer(in))
		out := (*string)(unsafe.Pointer(out))
		*out = *in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IRule.
func (in *IRule) DeepCopy() *IRule {
	if in == nil {
		return nil
	}
	out := new(IRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyCondition) DeepCopyInto(out *PolicyCondition) {
	*out = *in
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyCondition.
func (in *PolicyCondition) DeepCopy() *PolicyCondition {
	if in == nil {
		return nil
	}
	out := new(PolicyCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyRule) DeepCopyInto(out *PolicyRule) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]PolicyCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyRule.
func (in *PolicyRule) DeepCopy() *PolicyRule {
	if in == nil {
		return nil
	}
	out := new(PolicyRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Pool) DeepCopyInto(out *Pool) {
	{
		in := (*map[string]bool)(unsafe.Pointer(in))
		out := (*map[string]bool)(unsafe.Pointer(out))
		*out = make(map[string]bool, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Pool.
func (in *Pool) DeepCopy() *Pool {
	if in == nil {
		return nil
	}
	out := new(Pool)
	in.DeepCopyInto(out)
	return out
}
