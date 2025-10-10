package v1alpha1

import (
	v1alpha1podtypes "github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/api/v1beta1"
	v1beta1podtypes "github.com/kartverket/skiperator/api/v1beta1/podtypes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	if src.Spec.Container.Resources != nil {
		dst.Spec.Resources = convertResourceRequirements(src.Spec.Container.Resources)
	}

	if src.Spec.Container.AccessPolicy != nil {
		dst.Spec.AccessPolicy = convertAccessPolicy(src.Spec.Container.AccessPolicy)
	}

	if src.Spec.Container.GCP != nil {
		dst.Spec.GCP = convertGCP(src.Spec.Container.GCP)
	}

	if src.Spec.Container.PodSettings != nil {
		dst.Spec.PodSettings = convertPodSettings(src.Spec.Container.PodSettings)
	}

	if src.Spec.Container.Liveness != nil {
		dst.Spec.Liveness = convertProbe(src.Spec.Container.Liveness)
	}

	if src.Spec.Container.Readiness != nil {
		dst.Spec.Readiness = convertProbe(src.Spec.Container.Readiness)
	}

	if src.Spec.Container.Startup != nil {
		dst.Spec.Startup = convertProbe(src.Spec.Container.Startup)
	}

	dst.Spec.EnvFrom = convertEnvFromSlice(src.Spec.Container.EnvFrom)
	dst.Spec.FilesFrom = convertFilesFromSlice(src.Spec.Container.FilesFrom)
	dst.Spec.AdditionalPorts = convertInternalPortSlice(src.Spec.Container.AdditionalPorts)

	// Copy status
	dst.Status = convertSkiperatorStatus(&src.Status)

	return nil
}

func (dst *SKIPJob) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.SKIPJob)

	// Copy metadata
	dst.ObjectMeta = src.ObjectMeta

	// Copy Job and Cron settings
	dst.Spec.Job = src.Spec.Job
	dst.Spec.Cron = src.Spec.Cron

	// Copy IstioSettings and Prometheus
	dst.Spec.IstioSettings = src.Spec.IstioSettings
	dst.Spec.Prometheus = src.Spec.Prometheus

	// Copy primitive fields to Container
	dst.Spec.Container.Image = src.Spec.Image
	dst.Spec.Container.Priority = src.Spec.Priority
	dst.Spec.Container.Command = src.Spec.Command
	dst.Spec.Container.Env = src.Spec.Env
	dst.Spec.Container.RestartPolicy = src.Spec.RestartPolicy

	// Convert custom struct fields
	if src.Spec.Resources != nil {
		dst.Spec.Container.Resources = convertResourceRequirementsReverse(src.Spec.Resources)
	}

	if src.Spec.AccessPolicy != nil {
		dst.Spec.Container.AccessPolicy = convertAccessPolicyReverse(src.Spec.AccessPolicy)
	}

	if src.Spec.GCP != nil {
		dst.Spec.Container.GCP = convertGCPReverse(src.Spec.GCP)
	}

	if src.Spec.PodSettings != nil {
		dst.Spec.Container.PodSettings = convertPodSettingsReverse(src.Spec.PodSettings)
	}

	if src.Spec.Liveness != nil {
		dst.Spec.Container.Liveness = convertProbeReverse(src.Spec.Liveness)
	}

	if src.Spec.Readiness != nil {
		dst.Spec.Container.Readiness = convertProbeReverse(src.Spec.Readiness)
	}

	if src.Spec.Startup != nil {
		dst.Spec.Container.Startup = convertProbeReverse(src.Spec.Startup)
	}

	dst.Spec.Container.EnvFrom = convertEnvFromSliceReverse(src.Spec.EnvFrom)
	dst.Spec.Container.FilesFrom = convertFilesFromSliceReverse(src.Spec.FilesFrom)
	dst.Spec.Container.AdditionalPorts = convertInternalPortSliceReverse(src.Spec.AdditionalPorts)

	// Copy status
	dst.Status = convertSkiperatorStatus(src.Status)

	src.status

	return nil
}

// Helper conversion functions
func convertSkiperatorStatus(src *SkiperatorStatus) *v1beta1.SkiperatorStatus {
	if src == nil {
		return nil
	}
	return &v1beta1.SkiperatorStatus{
		Summary: v1beta1.Status{
			Status:    v1beta1.StatusNames(src.Summary.Status),
			Message:   src.Summary.Message,
			TimeStamp: src.Summary.TimeStamp,
		},
		SubResources: func() map[string]v1beta1.Status {
			if src.SubResources == nil {
				return nil
			}
			out := make(map[string]v1beta1.Status, len(src.SubResources))
			for k, v := range src.SubResources {
				out[k] = v1beta1.Status{
					Status:    v1beta1.StatusNames(v.Status),
					Message:   v.Message,
					TimeStamp: v.TimeStamp,
				}
			}
			return out
		}(),
		Conditions: func() []metav1.Condition {
			// If the metav1.Condition type is identical across versions, a shallow copy is fine.
			if src.Conditions == nil {
				return nil
			}
			out := make([]metav1.Condition, len(src.Conditions))
			copy(out, src.Conditions)
			return out
		}(),
		AccessPolicies: v1beta1.StatusNames(src.AccessPolicies),
	}
}

func convertResourceRequirements(src *v1alpha1podtypes.ResourceRequirements) *v1beta1podtypes.ResourceRequirements {
	if src == nil {
		return nil
	}
	return &v1beta1podtypes.ResourceRequirements{
		Limits:   src.Limits,
		Requests: src.Requests,
	}
}

func convertResourceRequirementsReverse(src *v1beta1podtypes.ResourceRequirements) *v1alpha1podtypes.ResourceRequirements {
	if src == nil {
		return nil
	}
	return &v1alpha1podtypes.ResourceRequirements{
		Limits:   src.Limits,
		Requests: src.Requests,
	}
}

func convertAccessPolicy(src *v1alpha1podtypes.AccessPolicy) *v1beta1podtypes.AccessPolicy {
	if src == nil {
		return nil
	}
	dst := &v1beta1podtypes.AccessPolicy{
		Outbound: &v1beta1podtypes.OutboundPolicy{},
		Inbound:  &v1beta1podtypes.InboundPolicy{},
	}

	if src.Outbound != nil {
		dst.Outbound.External = convertExternalRuleSlice(src.Outbound.External)
		dst.Outbound.Rules = convertInternalRuleSlice(src.Outbound.Rules)
	}

	if src.Inbound != nil {
		dst.Inbound.Rules = convertInternalRuleSlice(src.Inbound.Rules)
	}

	return dst
}

func convertAccessPolicyReverse(src *v1beta1podtypes.AccessPolicy) *v1alpha1podtypes.AccessPolicy {
	if src == nil {
		return nil
	}
	dst := &v1alpha1podtypes.AccessPolicy{
		Outbound: &v1alpha1podtypes.OutboundPolicy{},
		Inbound:  &v1alpha1podtypes.InboundPolicy{},
	}

	if src.Outbound != nil {
		dst.Outbound.External = convertExternalRuleSliceReverse(src.Outbound.External)
		dst.Outbound.Rules = convertInternalRuleSliceReverse(src.Outbound.Rules)
	}

	if src.Inbound != nil {
		dst.Inbound.Rules = convertInternalRuleSliceReverse(src.Inbound.Rules)
	}

	return dst
}

func convertExternalRuleSlice(src []v1alpha1podtypes.ExternalRule) []v1beta1podtypes.ExternalRule {
	if src == nil {
		return nil
	}
	dst := make([]v1beta1podtypes.ExternalRule, len(src))
	for i, rule := range src {
		// Convert ports
		prts := make([]v1beta1podtypes.ExternalPort, len(rule.Ports))
		for j, p := range rule.Ports {
			prts[j] = v1beta1podtypes.ExternalPort{
				Name:     p.Name,
				Port:     p.Port,
				Protocol: p.Protocol,
			}
		}

		dst[i] = v1beta1podtypes.ExternalRule{
			Host:  rule.Host,
			Ports: prts,
		}
	}
	return dst
}

func convertExternalRuleSliceReverse(src []v1beta1podtypes.ExternalRule) []v1alpha1podtypes.ExternalRule {
	if src == nil {
		return nil
	}
	dst := make([]v1alpha1podtypes.ExternalRule, len(src))
	for i, rule := range src {
		// Convert ports from v1beta1 -> v1alpha1
		var ports []v1alpha1podtypes.ExternalPort
		if rule.Ports != nil {
			ports = make([]v1alpha1podtypes.ExternalPort, len(rule.Ports))
			for j, p := range rule.Ports {
				ports[j] = v1alpha1podtypes.ExternalPort{
					Name:     p.Name,
					Port:     p.Port,
					Protocol: p.Protocol,
				}
			}
		}

		dst[i] = v1alpha1podtypes.ExternalRule{
			Host:  rule.Host,
			Ip:    rule.Ip,
			Ports: ports,
		}
	}
	return dst
}

func convertInternalRuleSlice(src []v1alpha1podtypes.InternalRule) []v1beta1podtypes.InternalRule {
	if src == nil {
		return nil
	}
	dst := make([]v1beta1podtypes.InternalRule, len(src))
	for i, rule := range src {
		dst[i] = v1beta1podtypes.InternalRule{
			Application: rule.Application,
			Namespace:   rule.Namespace,
			NamespacesByLabel: func() map[string]string {
				if rule.NamespacesByLabel != nil {
					return rule.NamespacesByLabel
				}
				return nil
			}(),
			Ports: rule.Ports,
		}
	}
	return dst
}

func convertInternalRuleSliceReverse(src []v1beta1podtypes.InternalRule) []v1alpha1podtypes.InternalRule {
	if src == nil {
		return nil
	}
	dst := make([]v1alpha1podtypes.InternalRule, len(src))
	for i, rule := range src {
		dst[i] = v1alpha1podtypes.InternalRule{
			Application: rule.Application,
			Namespace:   rule.Namespace,
			NamespacesByLabel: func() map[string]string {
				if rule.NamespacesByLabel != nil {
					return rule.NamespacesByLabel
				}
				return nil
			}(),
			Ports: rule.Ports,
		}
	}
	return dst
}

func convertGCP(src *v1alpha1podtypes.GCP) *v1beta1podtypes.GCP {
	if src == nil {
		return nil
	}
	return &v1beta1podtypes.GCP{
		Auth: &v1beta1podtypes.Auth{
			ServiceAccount: src.Auth.ServiceAccount,
		},
		CloudSQLProxy: &v1beta1podtypes.CloudSQLProxySettings{
			ConnectionName: src.CloudSQLProxy.ConnectionName,
			ServiceAccount: src.CloudSQLProxy.ServiceAccount,
			IP:             src.CloudSQLProxy.IP,
			Version:        src.CloudSQLProxy.Version,
			PublicIP:       src.CloudSQLProxy.PublicIP,
		},
	}
}

func convertGCPReverse(src *v1beta1podtypes.GCP) *v1alpha1podtypes.GCP {
	if src == nil {
		return nil
	}
	return &v1alpha1podtypes.GCP{
		Auth: &v1alpha1podtypes.Auth{
			ServiceAccount: src.Auth.ServiceAccount,
		},
		CloudSQLProxy: &v1alpha1podtypes.CloudSQLProxySettings{
			ConnectionName: src.CloudSQLProxy.ConnectionName,
			ServiceAccount: src.CloudSQLProxy.ServiceAccount,
			IP:             src.CloudSQLProxy.IP,
			Version:        src.CloudSQLProxy.Version,
			PublicIP:       src.CloudSQLProxy.PublicIP,
		},
	}
}

func convertPodSettings(src *v1alpha1podtypes.PodSettings) *v1beta1podtypes.PodSettings {
	if src == nil {
		return nil
	}
	return &v1beta1podtypes.PodSettings{
		Annotations:                         src.Annotations,
		TerminationGracePeriodSeconds:       src.TerminationGracePeriodSeconds,
		DisablePodSpreadTopologyConstraints: src.DisablePodSpreadTopologyConstraints,
	}
}

func convertPodSettingsReverse(src *v1beta1podtypes.PodSettings) *v1alpha1podtypes.PodSettings {
	if src == nil {
		return nil
	}
	return &v1alpha1podtypes.PodSettings{
		Annotations:                         src.Annotations,
		TerminationGracePeriodSeconds:       src.TerminationGracePeriodSeconds,
		DisablePodSpreadTopologyConstraints: src.DisablePodSpreadTopologyConstraints,
	}
}

func convertProbe(src *v1alpha1podtypes.Probe) *v1beta1podtypes.Probe {
	if src == nil {
		return nil
	}
	return &v1beta1podtypes.Probe{
		Path:             src.Path,
		Port:             src.Port,
		InitialDelay:     src.InitialDelay,
		Period:           src.Period,
		Timeout:          src.Timeout,
		SuccessThreshold: src.SuccessThreshold,
		FailureThreshold: src.FailureThreshold,
	}
}

func convertProbeReverse(src *v1beta1podtypes.Probe) *v1alpha1podtypes.Probe {
	if src == nil {
		return nil
	}
	return &v1alpha1podtypes.Probe{
		Path:             src.Path,
		Port:             src.Port,
		InitialDelay:     src.InitialDelay,
		Period:           src.Period,
		Timeout:          src.Timeout,
		SuccessThreshold: src.SuccessThreshold,
		FailureThreshold: src.FailureThreshold,
	}
}

func convertEnvFromSlice(src []v1alpha1podtypes.EnvFrom) []v1beta1podtypes.EnvFrom {
	if src == nil {
		return nil
	}
	dst := make([]v1beta1podtypes.EnvFrom, len(src))
	for i, envFrom := range src {
		dst[i] = v1beta1podtypes.EnvFrom{
			Secret:    envFrom.Secret,
			ConfigMap: envFrom.ConfigMap,
		}
	}
	return dst
}

func convertEnvFromSliceReverse(src []v1beta1podtypes.EnvFrom) []v1alpha1podtypes.EnvFrom {
	if src == nil {
		return nil
	}
	dst := make([]v1alpha1podtypes.EnvFrom, len(src))
	for i, envFrom := range src {
		dst[i] = v1alpha1podtypes.EnvFrom{
			Secret:    envFrom.Secret,
			ConfigMap: envFrom.ConfigMap,
		}
	}
	return dst
}

func convertFilesFromSlice(src []v1alpha1podtypes.FilesFrom) []v1beta1podtypes.FilesFrom {
	if src == nil {
		return nil
	}
	dst := make([]v1beta1podtypes.FilesFrom, len(src))
	for i, filesFrom := range src {
		dst[i] = v1beta1podtypes.FilesFrom{
			MountPath:             filesFrom.MountPath,
			ConfigMap:             filesFrom.ConfigMap,
			Secret:                filesFrom.Secret,
			EmptyDir:              filesFrom.EmptyDir,
			PersistentVolumeClaim: filesFrom.PersistentVolumeClaim,
			DefaultMode:           filesFrom.DefaultMode,
		}
	}
	return dst
}

func convertFilesFromSliceReverse(src []v1beta1podtypes.FilesFrom) []v1alpha1podtypes.FilesFrom {
	if src == nil {
		return nil
	}
	dst := make([]v1alpha1podtypes.FilesFrom, len(src))
	for i, filesFrom := range src {
		dst[i] = v1alpha1podtypes.FilesFrom{
			MountPath:             filesFrom.MountPath,
			ConfigMap:             filesFrom.ConfigMap,
			Secret:                filesFrom.Secret,
			EmptyDir:              filesFrom.EmptyDir,
			PersistentVolumeClaim: filesFrom.PersistentVolumeClaim,
			DefaultMode:           filesFrom.DefaultMode,
		}
	}
	return dst
}

func convertInternalPortSlice(src []v1alpha1podtypes.InternalPort) []v1beta1podtypes.InternalPort {
	if src == nil {
		return nil
	}
	dst := make([]v1beta1podtypes.InternalPort, len(src))
	for i, port := range src {
		dst[i] = v1beta1podtypes.InternalPort{
			Name:     port.Name,
			Port:     port.Port,
			Protocol: port.Protocol,
		}
	}
	return dst
}

func convertInternalPortSliceReverse(src []v1beta1podtypes.InternalPort) []v1alpha1podtypes.InternalPort {
	if src == nil {
		return nil
	}
	dst := make([]v1alpha1podtypes.InternalPort, len(src))
	for i, port := range src {
		dst[i] = v1alpha1podtypes.InternalPort{
			Name:     port.Name,
			Port:     port.Port,
			Protocol: port.Protocol,
		}
	}
	return dst
}
