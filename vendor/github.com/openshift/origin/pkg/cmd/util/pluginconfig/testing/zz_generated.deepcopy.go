// +build !ignore_autogenerated_openshift

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package testing

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TestConfig) DeepCopyInto(out *TestConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	if in.Item2 != nil {
		in, out := &in.Item2, &out.Item2
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TestConfig.
func (in *TestConfig) DeepCopy() *TestConfig {
	if in == nil {
		return nil
	}
	out := new(TestConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TestConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TestConfigV1) DeepCopyInto(out *TestConfigV1) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	if in.Item2 != nil {
		in, out := &in.Item2, &out.Item2
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TestConfigV1.
func (in *TestConfigV1) DeepCopy() *TestConfigV1 {
	if in == nil {
		return nil
	}
	out := new(TestConfigV1)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TestConfigV1) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}
