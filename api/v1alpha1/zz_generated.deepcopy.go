//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"github.com/kartverket/skiperator/v3/api/v1alpha1/digdirator"
	"github.com/kartverket/skiperator/v3/api/v1alpha1/istiotypes"
	"github.com/kartverket/skiperator/v3/api/v1alpha1/podtypes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	if in.Ingresses != nil {
		in, out := &in.Ingresses, &out.Ingresses
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
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
		*out = new(v1.JSON)
		(*in).DeepCopyInto(*out)
	}
	out.Strategy = in.Strategy
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]corev1.EnvVar, len(*in))
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
		*out = new(digdirator.Maskinporten)
		(*in).DeepCopyInto(*out)
	}
	if in.IDPorten != nil {
		in, out := &in.IDPorten, &out.IDPorten
		*out = new(digdirator.IDPorten)
		(*in).DeepCopyInto(*out)
	}
	if in.Prometheus != nil {
		in, out := &in.Prometheus, &out.Prometheus
		*out = new(PrometheusConfig)
		**out = **in
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
		(*in).DeepCopyInto(*out)
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
				inVal := (*in)[key]
				in, out := &inVal, &outVal
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
	if in.PodSettings != nil {
		in, out := &in.PodSettings, &out.PodSettings
		*out = new(podtypes.PodSettings)
		(*in).DeepCopyInto(*out)
	}
	if in.IstioSettings != nil {
		in, out := &in.IstioSettings, &out.IstioSettings
		*out = new(istiotypes.IstioSettings)
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
func (in *ContainerSettings) DeepCopyInto(out *ContainerSettings) {
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
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]corev1.EnvVar, len(*in))
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
	if in.AccessPolicy != nil {
		in, out := &in.AccessPolicy, &out.AccessPolicy
		*out = new(podtypes.AccessPolicy)
		(*in).DeepCopyInto(*out)
	}
	if in.GCP != nil {
		in, out := &in.GCP, &out.GCP
		*out = new(podtypes.GCP)
		(*in).DeepCopyInto(*out)
	}
	if in.RestartPolicy != nil {
		in, out := &in.RestartPolicy, &out.RestartPolicy
		*out = new(corev1.RestartPolicy)
		**out = **in
	}
	if in.PodSettings != nil {
		in, out := &in.PodSettings, &out.PodSettings
		*out = new(podtypes.PodSettings)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContainerSettings.
func (in *ContainerSettings) DeepCopy() *ContainerSettings {
	if in == nil {
		return nil
	}
	out := new(ContainerSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CronSettings) DeepCopyInto(out *CronSettings) {
	*out = *in
	if in.TimeZone != nil {
		in, out := &in.TimeZone, &out.TimeZone
		*out = new(string)
		**out = **in
	}
	if in.StartingDeadlineSeconds != nil {
		in, out := &in.StartingDeadlineSeconds, &out.StartingDeadlineSeconds
		*out = new(int64)
		**out = **in
	}
	if in.Suspend != nil {
		in, out := &in.Suspend, &out.Suspend
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CronSettings.
func (in *CronSettings) DeepCopy() *CronSettings {
	if in == nil {
		return nil
	}
	out := new(CronSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobSettings) DeepCopyInto(out *JobSettings) {
	*out = *in
	if in.ActiveDeadlineSeconds != nil {
		in, out := &in.ActiveDeadlineSeconds, &out.ActiveDeadlineSeconds
		*out = new(int64)
		**out = **in
	}
	if in.BackoffLimit != nil {
		in, out := &in.BackoffLimit, &out.BackoffLimit
		*out = new(int32)
		**out = **in
	}
	if in.Suspend != nil {
		in, out := &in.Suspend, &out.Suspend
		*out = new(bool)
		**out = **in
	}
	if in.TTLSecondsAfterFinished != nil {
		in, out := &in.TTLSecondsAfterFinished, &out.TTLSecondsAfterFinished
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobSettings.
func (in *JobSettings) DeepCopy() *JobSettings {
	if in == nil {
		return nil
	}
	out := new(JobSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrometheusConfig) DeepCopyInto(out *PrometheusConfig) {
	*out = *in
	out.Port = in.Port
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
func (in *Route) DeepCopyInto(out *Route) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Route.
func (in *Route) DeepCopy() *Route {
	if in == nil {
		return nil
	}
	out := new(Route)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Routing) DeepCopyInto(out *Routing) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Routing.
func (in *Routing) DeepCopy() *Routing {
	if in == nil {
		return nil
	}
	out := new(Routing)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Routing) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoutingList) DeepCopyInto(out *RoutingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Routing, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoutingList.
func (in *RoutingList) DeepCopy() *RoutingList {
	if in == nil {
		return nil
	}
	out := new(RoutingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RoutingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoutingSpec) DeepCopyInto(out *RoutingSpec) {
	*out = *in
	if in.Routes != nil {
		in, out := &in.Routes, &out.Routes
		*out = make([]Route, len(*in))
		copy(*out, *in)
	}
	if in.RedirectToHTTPS != nil {
		in, out := &in.RedirectToHTTPS, &out.RedirectToHTTPS
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoutingSpec.
func (in *RoutingSpec) DeepCopy() *RoutingSpec {
	if in == nil {
		return nil
	}
	out := new(RoutingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoutingStatus) DeepCopyInto(out *RoutingStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoutingStatus.
func (in *RoutingStatus) DeepCopy() *RoutingStatus {
	if in == nil {
		return nil
	}
	out := new(RoutingStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SKIPJob) DeepCopyInto(out *SKIPJob) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SKIPJob.
func (in *SKIPJob) DeepCopy() *SKIPJob {
	if in == nil {
		return nil
	}
	out := new(SKIPJob)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SKIPJob) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SKIPJobList) DeepCopyInto(out *SKIPJobList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SKIPJob, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SKIPJobList.
func (in *SKIPJobList) DeepCopy() *SKIPJobList {
	if in == nil {
		return nil
	}
	out := new(SKIPJobList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SKIPJobList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SKIPJobSpec) DeepCopyInto(out *SKIPJobSpec) {
	*out = *in
	if in.Job != nil {
		in, out := &in.Job, &out.Job
		*out = new(JobSettings)
		(*in).DeepCopyInto(*out)
	}
	if in.Cron != nil {
		in, out := &in.Cron, &out.Cron
		*out = new(CronSettings)
		(*in).DeepCopyInto(*out)
	}
	in.Container.DeepCopyInto(&out.Container)
	if in.IstioSettings != nil {
		in, out := &in.IstioSettings, &out.IstioSettings
		*out = new(istiotypes.IstioSettings)
		(*in).DeepCopyInto(*out)
	}
	if in.Prometheus != nil {
		in, out := &in.Prometheus, &out.Prometheus
		*out = new(PrometheusConfig)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SKIPJobSpec.
func (in *SKIPJobSpec) DeepCopy() *SKIPJobSpec {
	if in == nil {
		return nil
	}
	out := new(SKIPJobSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SKIPJobStatus) DeepCopyInto(out *SKIPJobStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SKIPJobStatus.
func (in *SKIPJobStatus) DeepCopy() *SKIPJobStatus {
	if in == nil {
		return nil
	}
	out := new(SKIPJobStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkiperatorStatus) DeepCopyInto(out *SkiperatorStatus) {
	*out = *in
	out.Summary = in.Summary
	if in.SubResources != nil {
		in, out := &in.SubResources, &out.SubResources
		*out = make(map[string]Status, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkiperatorStatus.
func (in *SkiperatorStatus) DeepCopy() *SkiperatorStatus {
	if in == nil {
		return nil
	}
	out := new(SkiperatorStatus)
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
