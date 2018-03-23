// +build !ignore_autogenerated_openshift

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package oauth

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
	unsafe "unsafe"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterRoleScopeRestriction) DeepCopyInto(out *ClusterRoleScopeRestriction) {
	*out = *in
	if in.RoleNames != nil {
		in, out := &in.RoleNames, &out.RoleNames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Namespaces != nil {
		in, out := &in.Namespaces, &out.Namespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterRoleScopeRestriction.
func (in *ClusterRoleScopeRestriction) DeepCopy() *ClusterRoleScopeRestriction {
	if in == nil {
		return nil
	}
	out := new(ClusterRoleScopeRestriction)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GrantHandlerType) DeepCopyInto(out *GrantHandlerType) {
	{
		in := (*string)(unsafe.Pointer(in))
		out := (*string)(unsafe.Pointer(out))
		*out = *in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GrantHandlerType.
func (in *GrantHandlerType) DeepCopy() *GrantHandlerType {
	if in == nil {
		return nil
	}
	out := new(GrantHandlerType)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OAuthAccessToken) DeepCopyInto(out *OAuthAccessToken) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OAuthAccessToken.
func (in *OAuthAccessToken) DeepCopy() *OAuthAccessToken {
	if in == nil {
		return nil
	}
	out := new(OAuthAccessToken)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OAuthAccessToken) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OAuthAccessTokenList) DeepCopyInto(out *OAuthAccessTokenList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OAuthAccessToken, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OAuthAccessTokenList.
func (in *OAuthAccessTokenList) DeepCopy() *OAuthAccessTokenList {
	if in == nil {
		return nil
	}
	out := new(OAuthAccessTokenList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OAuthAccessTokenList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OAuthAuthorizeToken) DeepCopyInto(out *OAuthAuthorizeToken) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OAuthAuthorizeToken.
func (in *OAuthAuthorizeToken) DeepCopy() *OAuthAuthorizeToken {
	if in == nil {
		return nil
	}
	out := new(OAuthAuthorizeToken)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OAuthAuthorizeToken) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OAuthAuthorizeTokenList) DeepCopyInto(out *OAuthAuthorizeTokenList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OAuthAuthorizeToken, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OAuthAuthorizeTokenList.
func (in *OAuthAuthorizeTokenList) DeepCopy() *OAuthAuthorizeTokenList {
	if in == nil {
		return nil
	}
	out := new(OAuthAuthorizeTokenList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OAuthAuthorizeTokenList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OAuthClient) DeepCopyInto(out *OAuthClient) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.AdditionalSecrets != nil {
		in, out := &in.AdditionalSecrets, &out.AdditionalSecrets
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.RedirectURIs != nil {
		in, out := &in.RedirectURIs, &out.RedirectURIs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ScopeRestrictions != nil {
		in, out := &in.ScopeRestrictions, &out.ScopeRestrictions
		*out = make([]ScopeRestriction, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AccessTokenMaxAgeSeconds != nil {
		in, out := &in.AccessTokenMaxAgeSeconds, &out.AccessTokenMaxAgeSeconds
		if *in == nil {
			*out = nil
		} else {
			*out = new(int32)
			**out = **in
		}
	}
	if in.AccessTokenInactivityTimeoutSeconds != nil {
		in, out := &in.AccessTokenInactivityTimeoutSeconds, &out.AccessTokenInactivityTimeoutSeconds
		if *in == nil {
			*out = nil
		} else {
			*out = new(int32)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OAuthClient.
func (in *OAuthClient) DeepCopy() *OAuthClient {
	if in == nil {
		return nil
	}
	out := new(OAuthClient)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OAuthClient) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OAuthClientAuthorization) DeepCopyInto(out *OAuthClientAuthorization) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OAuthClientAuthorization.
func (in *OAuthClientAuthorization) DeepCopy() *OAuthClientAuthorization {
	if in == nil {
		return nil
	}
	out := new(OAuthClientAuthorization)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OAuthClientAuthorization) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OAuthClientAuthorizationList) DeepCopyInto(out *OAuthClientAuthorizationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OAuthClientAuthorization, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OAuthClientAuthorizationList.
func (in *OAuthClientAuthorizationList) DeepCopy() *OAuthClientAuthorizationList {
	if in == nil {
		return nil
	}
	out := new(OAuthClientAuthorizationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OAuthClientAuthorizationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OAuthClientList) DeepCopyInto(out *OAuthClientList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OAuthClient, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OAuthClientList.
func (in *OAuthClientList) DeepCopy() *OAuthClientList {
	if in == nil {
		return nil
	}
	out := new(OAuthClientList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OAuthClientList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OAuthRedirectReference) DeepCopyInto(out *OAuthRedirectReference) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Reference = in.Reference
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OAuthRedirectReference.
func (in *OAuthRedirectReference) DeepCopy() *OAuthRedirectReference {
	if in == nil {
		return nil
	}
	out := new(OAuthRedirectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OAuthRedirectReference) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedirectReference) DeepCopyInto(out *RedirectReference) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedirectReference.
func (in *RedirectReference) DeepCopy() *RedirectReference {
	if in == nil {
		return nil
	}
	out := new(RedirectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScopeRestriction) DeepCopyInto(out *ScopeRestriction) {
	*out = *in
	if in.ExactValues != nil {
		in, out := &in.ExactValues, &out.ExactValues
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ClusterRole != nil {
		in, out := &in.ClusterRole, &out.ClusterRole
		if *in == nil {
			*out = nil
		} else {
			*out = new(ClusterRoleScopeRestriction)
			(*in).DeepCopyInto(*out)
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScopeRestriction.
func (in *ScopeRestriction) DeepCopy() *ScopeRestriction {
	if in == nil {
		return nil
	}
	out := new(ScopeRestriction)
	in.DeepCopyInto(out)
	return out
}
