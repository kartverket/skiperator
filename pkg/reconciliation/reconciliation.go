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
)

type Reconciliation interface {
	GetLogger() log.Logger
	GetCtx() context.Context
	IsIstioEnabled() bool
	GetReconciliationObject() client.Object
	GetType() ReconciliationObjectType
	GetSyncObjects() []*client.Object
	AddSyncObject(*client.Object)
}
