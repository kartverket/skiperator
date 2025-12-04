package v1alpha1

import (
	"errors"

	"github.com/kartverket/skiperator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo: v1alpha1 -> v1beta1
func (src *SKIPJob) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*v1beta1.SKIPJob)
	if !ok {
		return errors.New("cannot convert SKIPJob from v1alpha1 to v1beta1")
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Job = src.Spec.Job
	dst.Spec.Cron = src.Spec.Cron
	dst.Spec.IstioSettings = src.Spec.IstioSettings
	dst.Spec.Prometheus = src.Spec.Prometheus

	flattenContainer(&src.Spec.Container, &dst.Spec)

	dst.Status = src.Status

	return nil
}

// ConvertFrom: v1beta1 -> v1alpha1
func (dst *SKIPJob) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*v1beta1.SKIPJob)
	if !ok {
		return errors.New("cannot convert SKIPJob from v1beta1 to v1alpha1")
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Job = src.Spec.Job
	dst.Spec.Cron = src.Spec.Cron
	dst.Spec.IstioSettings = src.Spec.IstioSettings
	dst.Spec.Prometheus = src.Spec.Prometheus

	expandContainer(&src.Spec, &dst.Spec.Container)

	dst.Status = src.Status
	return nil
}

// v1alpha1.Spec.Container -> v1beta1.Spec (flatten)
func flattenContainer(src *ContainerSettings, dst *v1beta1.SKIPJobSpec) {
	dst.Image = src.Image
	dst.Priority = src.Priority
	dst.Command = src.Command
	dst.Env = src.Env
	dst.RestartPolicy = src.RestartPolicy

	dst.Resources = src.Resources
	dst.AccessPolicy = src.AccessPolicy
	dst.GCP = src.GCP
	dst.PodSettings = src.PodSettings
	dst.Liveness = src.Liveness
	dst.Readiness = src.Readiness
	dst.Startup = src.Startup
	dst.EnvFrom = src.EnvFrom
	dst.FilesFrom = src.FilesFrom
	dst.AdditionalPorts = src.AdditionalPorts
}

// v1beta1.Spec (flattened) -> v1alpha1.Spec.Container (expand)
func expandContainer(src *v1beta1.SKIPJobSpec, dst *ContainerSettings) {
	dst.Image = src.Image
	dst.Priority = src.Priority
	dst.Command = src.Command
	dst.Env = src.Env
	dst.RestartPolicy = src.RestartPolicy

	dst.Resources = src.Resources
	dst.AccessPolicy = src.AccessPolicy
	dst.GCP = src.GCP
	dst.PodSettings = src.PodSettings
	dst.Liveness = src.Liveness
	dst.Readiness = src.Readiness
	dst.Startup = src.Startup
	dst.EnvFrom = src.EnvFrom
	dst.FilesFrom = src.FilesFrom
	dst.AdditionalPorts = src.AdditionalPorts
}
