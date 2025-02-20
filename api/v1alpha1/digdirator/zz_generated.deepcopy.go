//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package digdirator

import (
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IDPorten) DeepCopyInto(out *IDPorten) {
	*out = *in
	if in.ClientName != nil {
		in, out := &in.ClientName, &out.ClientName
		*out = new(string)
		**out = **in
	}
	if in.AccessTokenLifetime != nil {
		in, out := &in.AccessTokenLifetime, &out.AccessTokenLifetime
		*out = new(int)
		**out = **in
	}
	if in.PostLogoutRedirectURIs != nil {
		in, out := &in.PostLogoutRedirectURIs, &out.PostLogoutRedirectURIs
		*out = new([]v1.IDPortenURI)
		if **in != nil {
			in, out := *in, *out
			*out = make([]v1.IDPortenURI, len(*in))
			copy(*out, *in)
		}
	}
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SessionLifetime != nil {
		in, out := &in.SessionLifetime, &out.SessionLifetime
		*out = new(int)
		**out = **in
	}
	if in.Authentication != nil {
		in, out := &in.Authentication, &out.Authentication
		*out = new(istiotypes.RequestAuthentication)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IDPorten.
func (in *IDPorten) DeepCopy() *IDPorten {
	if in == nil {
		return nil
	}
	out := new(IDPorten)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Maskinporten) DeepCopyInto(out *Maskinporten) {
	*out = *in
	if in.ClientName != nil {
		in, out := &in.ClientName, &out.ClientName
		*out = new(string)
		**out = **in
	}
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = new(v1.MaskinportenScope)
		(*in).DeepCopyInto(*out)
	}
	if in.Authentication != nil {
		in, out := &in.Authentication, &out.Authentication
		*out = new(istiotypes.RequestAuthentication)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Maskinporten.
func (in *Maskinporten) DeepCopy() *Maskinporten {
	if in == nil {
		return nil
	}
	out := new(Maskinporten)
	in.DeepCopyInto(out)
	return out
}
