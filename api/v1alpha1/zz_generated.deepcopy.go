//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	nais_iov1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Application) DeepCopyInto(out *Application) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Application.
func (in *Application) DeepCopy() *Application {
	if in == nil {
		return nil
	}
	out := new(Application)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Application) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationList) DeepCopyInto(out *ApplicationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Application, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationList.
func (in *ApplicationList) DeepCopy() *ApplicationList {
	if in == nil {
		return nil
	}
	out := new(ApplicationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ApplicationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationSpec) DeepCopyInto(out *ApplicationSpec) {
	*out = *in
	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(podtypes.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(Replicas)
		**out = **in
	}
	out.Strategy = in.Strategy
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.EnvFrom != nil {
		in, out := &in.EnvFrom, &out.EnvFrom
		*out = make([]podtypes.EnvFrom, len(*in))
		copy(*out, *in)
	}
	if in.FilesFrom != nil {
		in, out := &in.FilesFrom, &out.FilesFrom
		*out = make([]podtypes.FilesFrom, len(*in))
		copy(*out, *in)
	}
	if in.AdditionalPorts != nil {
		in, out := &in.AdditionalPorts, &out.AdditionalPorts
		*out = make([]podtypes.InternalPort, len(*in))
		copy(*out, *in)
	}
	if in.Prometheus != nil {
		in, out := &in.Prometheus, &out.Prometheus
		*out = new(PrometheusConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Liveness != nil {
		in, out := &in.Liveness, &out.Liveness
		*out = new(podtypes.Probe)
		**out = **in
	}
	if in.Readiness != nil {
		in, out := &in.Readiness, &out.Readiness
		*out = new(podtypes.Probe)
		**out = **in
	}
	if in.Startup != nil {
		in, out := &in.Startup, &out.Startup
		*out = new(podtypes.Probe)
		**out = **in
	}
	if in.Maskinporten != nil {
		in, out := &in.Maskinporten, &out.Maskinporten
		*out = new(Maskinporten)
		(*in).DeepCopyInto(*out)
	}
	if in.IDPorten != nil {
		in, out := &in.IDPorten, &out.IDPorten
		*out = new(IDPorten)
		(*in).DeepCopyInto(*out)
	}
	if in.Ingresses != nil {
		in, out := &in.Ingresses, &out.Ingresses
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.RedirectToHTTPS != nil {
		in, out := &in.RedirectToHTTPS, &out.RedirectToHTTPS
		*out = new(bool)
		**out = **in
	}
	if in.EnablePDB != nil {
		in, out := &in.EnablePDB, &out.EnablePDB
		*out = new(bool)
		**out = **in
	}
	if in.AccessPolicy != nil {
		in, out := &in.AccessPolicy, &out.AccessPolicy
		*out = new(podtypes.AccessPolicy)
		(*in).DeepCopyInto(*out)
	}
	if in.GCP != nil {
		in, out := &in.GCP, &out.GCP
		*out = new(podtypes.GCP)
		**out = **in
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ResourceLabels != nil {
		in, out := &in.ResourceLabels, &out.ResourceLabels
		*out = make(map[string]map[string]string, len(*in))
		for key, val := range *in {
			var outVal map[string]string
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make(map[string]string, len(*in))
				for key, val := range *in {
					(*out)[key] = val
				}
			}
			(*out)[key] = outVal
		}
	}
	if in.AuthorizationSettings != nil {
		in, out := &in.AuthorizationSettings, &out.AuthorizationSettings
		*out = new(AuthorizationSettings)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationSpec.
func (in *ApplicationSpec) DeepCopy() *ApplicationSpec {
	if in == nil {
		return nil
	}
	out := new(ApplicationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationStatus) DeepCopyInto(out *ApplicationStatus) {
	*out = *in
	out.ApplicationStatus = in.ApplicationStatus
	if in.ControllersStatus != nil {
		in, out := &in.ControllersStatus, &out.ControllersStatus
		*out = make(map[string]Status, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationStatus.
func (in *ApplicationStatus) DeepCopy() *ApplicationStatus {
	if in == nil {
		return nil
	}
	out := new(ApplicationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthorizationSettings) DeepCopyInto(out *AuthorizationSettings) {
	*out = *in
	if in.AllowList != nil {
		in, out := &in.AllowList, &out.AllowList
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthorizationSettings.
func (in *AuthorizationSettings) DeepCopy() *AuthorizationSettings {
	if in == nil {
		return nil
	}
	out := new(AuthorizationSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IDPorten) DeepCopyInto(out *IDPorten) {
	*out = *in
	if in.AccessTokenLifetime != nil {
		in, out := &in.AccessTokenLifetime, &out.AccessTokenLifetime
		*out = new(int)
		**out = **in
	}
	if in.PostLogoutRedirectURIs != nil {
		in, out := &in.PostLogoutRedirectURIs, &out.PostLogoutRedirectURIs
		*out = new([]nais_iov1.IDPortenURI)
		if **in != nil {
			in, out := *in, *out
			*out = make([]nais_iov1.IDPortenURI, len(*in))
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
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = new(nais_iov1.MaskinportenScope)
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrometheusConfig) DeepCopyInto(out *PrometheusConfig) {
	*out = *in
	out.Port = in.Port
	if in.IstioEnabled != nil {
		in, out := &in.IstioEnabled, &out.IstioEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrometheusConfig.
func (in *PrometheusConfig) DeepCopy() *PrometheusConfig {
	if in == nil {
		return nil
	}
	out := new(PrometheusConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Replicas) DeepCopyInto(out *Replicas) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Replicas.
func (in *Replicas) DeepCopy() *Replicas {
	if in == nil {
		return nil
	}
	out := new(Replicas)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceRequirements) DeepCopyInto(out *ResourceRequirements) {
	*out = *in
	if in.Limits != nil {
		in, out := &in.Limits, &out.Limits
		*out = make(v1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	if in.Requests != nil {
		in, out := &in.Requests, &out.Requests
		*out = make(v1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceRequirements.
func (in *ResourceRequirements) DeepCopy() *ResourceRequirements {
	if in == nil {
		return nil
	}
	out := new(ResourceRequirements)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Status) DeepCopyInto(out *Status) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Status.
func (in *Status) DeepCopy() *Status {
	if in == nil {
		return nil
	}
	out := new(Status)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Strategy) DeepCopyInto(out *Strategy) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Strategy.
func (in *Strategy) DeepCopy() *Strategy {
	if in == nil {
		return nil
	}
	out := new(Strategy)
	in.DeepCopyInto(out)
	return out
}
