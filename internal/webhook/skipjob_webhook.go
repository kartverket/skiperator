package webhook

import (
	"context"
	"fmt"
	"maps"

	v1beta1 "github.com/kartverket/skiperator/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:docs-gen:collapse=Imports
// Set up logger
var skipJobLog = logf.Log.WithName("skipjop-resource")

// nolint:unused
// +kubebuilder:webhook:path=/mutate-skiperator-kartverket-no-v1beta1-skipjob,mutating=true,failurePolicy=fail,sideEffects=None,groups=skiperator.kartverket.no,resources=skipjobs,verbs=create;update,versions=v1beta1;v1alpha1,name=mskipjob.skiperator.kartverket.no,admissionReviewVersions=v1
// Add a SKIPJob Defaulter that fills the default spec for the SKIPJob
type SKIPJobCustomDefaulter struct {
	DefaultJobSettings             v1beta1.JobSettings
	DefaultTTLSecondsAfterFinished int32
	DefaultBackoffLimit            int32
	DefaultSuspend                 bool
}

var _ webhook.CustomDefaulter = &SKIPJobCustomDefaulter{}

func SetupSkipJobWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1beta1.SKIPJob{}).
		WithDefaulter(&SKIPJobCustomDefaulter{
			DefaultJobSettings:             v1beta1.JobSettings{},
			DefaultTTLSecondsAfterFinished: v1beta1.DefaultTTLSecondsAfterFinished,
			DefaultBackoffLimit:            v1beta1.DefaultBackoffLimit,
			DefaultSuspend:                 v1beta1.DefaultSuspend,
		}).
		// TODO: Add a validator when this works...
		Complete()
}

/*
We use the `webhook.CustomDefaulter`interface to set defaults to our CRD.
A webhook will automatically be served that calls this defaulting.

The `Default`method is expected to mutate the receiver, setting the defaults.
*/
// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind SKIPJob.
func (d *SKIPJobCustomDefaulter) Default(ctx context.Context, object runtime.Object) error {
	skipJob, ok := object.(*v1beta1.SKIPJob)
	if !ok {
		return fmt.Errorf("expected a SKIPJob object, but got %T", object)
	}
	skipJobLog.Info("Defaulting for skipJob", "name", skipJob.GetName())
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
	maps.Copy(labels, skipJob.GetDefaultLabels())
	skipJob.SetLabels(labels)
}
