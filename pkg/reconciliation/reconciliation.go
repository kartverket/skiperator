package reconciliation

import (
	"context"
	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReconciliationObjectType string

const (
	ApplicationType ReconciliationObjectType = "Application"
	JobType         ReconciliationObjectType = "SKIPJob"
	NamespaceType   ReconciliationObjectType = "Namespace"
	RoutingType     ReconciliationObjectType = "Routing"
)

type Reconciliation interface {
	GetLogger() log.Logger
	GetCtx() context.Context
	IsIstioEnabled() bool
	GetSKIPObject() v1alpha1.SKIPObject
	GetCommonSpec() *CommonType
	GetType() ReconciliationObjectType
	GetResources() []client.Object
	AddResource(client.Object)
	GetIdentityConfigMap() *corev1.ConfigMap
	GetRestConfig() *rest.Config
}

// TODO Move to types?
type CommonType struct {
	AccessPolicy *podtypes.AccessPolicy
	GCP          *podtypes.GCP
}
