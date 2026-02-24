package webhook

import (
	"context"
	"fmt"
	"maps"

	v1beta1 "github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/pkg/log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:docs-gen:collapse=Imports
// Set up logger
var skipJobLog = log.NewLogger().WithName("skipjob-controller")

// nolint:unused
// +kubebuilder:webhook:path=/mutate-skiperator-kartverket-no-v1beta1-skipjob,mutating=true,failurePolicy=fail,sideEffects=None,groups=skiperator.kartverket.no,resources=skipjobs,verbs=create;update,versions=v1beta1;v1alpha1,name=mskipjob.skiperator.kartverket.no,admissionReviewVersions=v1
// Add a SKIPJob Defaulter that fills the default spec for the SKIPJob
type SKIPJobCustomDefaulter struct {
	DefaultJobSettings             v1beta1.JobSettings
	DefaultTTLSecondsAfterFinished int32
	DefaultBackoffLimit            int32
	DefaultSuspend                 bool
}

var _ admission.Defaulter[*v1beta1.SKIPJob] = &SKIPJobCustomDefaulter{}

func SetupSkipJobWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &v1beta1.SKIPJob{}).
		WithDefaulter(&SKIPJobCustomDefaulter{
			DefaultJobSettings:             v1beta1.JobSettings{},
			DefaultTTLSecondsAfterFinished: v1beta1.DefaultTTLSecondsAfterFinished,
			DefaultBackoffLimit:            v1beta1.DefaultBackoffLimit,
			DefaultSuspend:                 v1beta1.DefaultSuspend,
		}).
		Complete()
}

/*
We use the `webhook.CustomDefaulter`interface to set defaults to our CRD.
A webhook will automatically be served that calls this defaulting.

The `Default`method is expected to mutate the receiver, setting the defaults.
*/
// Default implements admission.Defaulter so a webhook will be registered for the Kind SKIPJob.
func (d *SKIPJobCustomDefaulter) Default(ctx context.Context, skipJob *v1beta1.SKIPJob) error {
	if skipJob == nil {
		return fmt.Errorf("expected a SKIPJob object, but got nil")
	}
	skipJobLog.Debug("Defaulting for skipJob", "name", skipJob.GetName())
	// The mutating webhook should only set defaults on admission, e.g they are static
	skipJob.FillDefaultSpec()
	d.applySKIPJobLabels(skipJob)
	return nil
}

func (d *SKIPJobCustomDefaulter) applySKIPJobLabels(skipJob *v1beta1.SKIPJob) {
	labels := skipJob.GetLabels()
	if len(labels) == 0 {
		labels = make(map[string]string)
	}
	maps.Copy(labels, skipJob.Spec.Labels)
	maps.Copy(labels, skipJob.GetDefaultLabels())
	skipJob.SetLabels(labels)
}
