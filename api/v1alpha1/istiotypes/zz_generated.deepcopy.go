//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package istiotypes

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Authentication) DeepCopyInto(out *Authentication) {
	*out = *in
	if in.SecretName != nil {
		in, out := &in.SecretName, &out.SecretName
		*out = new(string)
		**out = **in
	}
	if in.OutputClaimToHeaders != nil {
		in, out := &in.OutputClaimToHeaders, &out.OutputClaimToHeaders
		*out = new([]ClaimToHeader)
		if **in != nil {
			in, out := *in, *out
			*out = make([]ClaimToHeader, len(*in))
			copy(*out, *in)
		}
	}
	if in.IgnorePaths != nil {
		in, out := &in.IgnorePaths, &out.IgnorePaths
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Authentication.
func (in *Authentication) DeepCopy() *Authentication {
	if in == nil {
		return nil
	}
	out := new(Authentication)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IstioSettings) DeepCopyInto(out *IstioSettings) {
	*out = *in
	in.Telemetry.DeepCopyInto(&out.Telemetry)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IstioSettings.
func (in *IstioSettings) DeepCopy() *IstioSettings {
	if in == nil {
		return nil
	}
	out := new(IstioSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Telemetry) DeepCopyInto(out *Telemetry) {
	*out = *in
	if in.Tracing != nil {
		in, out := &in.Tracing, &out.Tracing
		*out = make([]*Tracing, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Tracing)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Telemetry.
func (in *Telemetry) DeepCopy() *Telemetry {
	if in == nil {
		return nil
	}
	out := new(Telemetry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Tracing) DeepCopyInto(out *Tracing) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Tracing.
func (in *Tracing) DeepCopy() *Tracing {
	if in == nil {
		return nil
	}
	out := new(Tracing)
	in.DeepCopyInto(out)
	return out
}
