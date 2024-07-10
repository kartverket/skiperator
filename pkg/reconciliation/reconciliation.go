package reconciliation

import (
	"context"
	"github.com/kartverket/skiperator/pkg/log"
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
	GetReconciliationObject() client.Object
	GetType() ReconciliationObjectType
	GetResources() []*client.Object
	AddResource(*client.Object)
}
