package v1alpha1

import (
	"github.com/kartverket/skiperator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *SKIPJob) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.SKIPJob)

	// Copy metadata
	dst.ObjectMeta = src.ObjectMeta

	// Copy Job and Cron settings (these are simple structs with primitive types)
	dst.Spec.Job = src.Spec.Job
	dst.Spec.Cron = src.Spec.Cron

	// Copy IstioSettings and Prometheus (these should be compatible)
	dst.Spec.IstioSettings = src.Spec.IstioSettings
	dst.Spec.Prometheus = src.Spec.Prometheus

	// Copy primitive fields from Container
	dst.Spec.Image = src.Spec.Container.Image
	dst.Spec.Priority = src.Spec.Container.Priority
	dst.Spec.Command = src.Spec.Container.Command
	dst.Spec.Env = src.Spec.Container.Env
	dst.Spec.RestartPolicy = src.Spec.Container.RestartPolicy

	// Convert custom struct fields
	dst.Spec.Resources = src.Spec.Container.Resources

	dst.Spec.AccessPolicy = src.Spec.Container.AccessPolicy

	dst.Spec.GCP = src.Spec.Container.GCP

	dst.Spec.PodSettings = src.Spec.Container.PodSettings

	dst.Spec.Liveness = src.Spec.Container.Liveness

	dst.Spec.Readiness = src.Spec.Container.Readiness

	dst.Spec.Startup = src.Spec.Container.Startup

	dst.Spec.EnvFrom = src.Spec.Container.EnvFrom
	dst.Spec.FilesFrom = src.Spec.Container.FilesFrom
	dst.Spec.AdditionalPorts = src.Spec.Container.AdditionalPorts

	// Copy status
	dst.Status = src.Status

	return nil
}
