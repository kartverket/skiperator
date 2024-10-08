//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package istiotypes

import ()

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
