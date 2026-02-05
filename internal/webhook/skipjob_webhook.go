package webhook

import (
	"context"
	"fmt"
	"maps"

	"github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	client                         client.Client
}

var _ admission.Defaulter[*v1beta1.SKIPJob] = &SKIPJobCustomDefaulter{}

func SetupSkipJobWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &v1beta1.SKIPJob{}).
		WithDefaulter(&SKIPJobCustomDefaulter{
			DefaultJobSettings:             v1beta1.JobSettings{},
			DefaultTTLSecondsAfterFinished: v1beta1.DefaultTTLSecondsAfterFinished,
			DefaultBackoffLimit:            v1beta1.DefaultBackoffLimit,
			DefaultSuspend:                 v1beta1.DefaultSuspend,
			client:                         mgr.GetClient(),
		}).
		Complete()
}

// Default implements admission.Defaulter so a webhook will be registered for the Kind SKIPJob.
func (d *SKIPJobCustomDefaulter) Default(ctx context.Context, skipJob *v1beta1.SKIPJob) error {
	skipJobLog.Debug("Defaulting for skipJob", "name", skipJob.GetName())
	// The mutating webhook should only set defaults on admission, e.g they are static
	skipJob.FillDefaultSpec()
	d.applySKIPJobLabels(ctx, skipJob)
	return nil
}

func (d *SKIPJobCustomDefaulter) applySKIPJobLabels(ctx context.Context, skipJob *v1beta1.SKIPJob) {
	labels := skipJob.GetLabels()
	if len(labels) == 0 {
		labels = make(map[string]string)
	}
	maps.Copy(labels, skipJob.GetDefaultLabels())
	maps.Copy(skipJob.Labels, skipJob.Spec.Labels)

	// Adds team label from namespace if not set in spec
	if len(skipJob.Spec.Team) == 0 {
		if name, err := d.teamNameForNamespace(ctx, skipJob); err == nil {
			skipJob.Spec.Team = name
		}
	}

	skipJob.SetLabels(labels)
}

func (d *SKIPJobCustomDefaulter) teamNameForNamespace(ctx context.Context, skipObj v1beta1.SKIPObject) (string, error) {
	ns := &corev1.Namespace{}
	if err := d.client.Get(ctx, types.NamespacedName{Name: skipObj.GetNamespace()}, ns); err != nil {
		return "", err
	}

	teamValue := ns.Labels["team"]
	if len(teamValue) > 0 {
		return teamValue, nil
	}
	return "", fmt.Errorf("missing value for team label")
}
