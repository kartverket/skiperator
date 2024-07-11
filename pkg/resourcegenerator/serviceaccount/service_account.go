package serviceaccount

import (
	"fmt"
	"github.com/kartverket/skiperator/pkg/reconciliation"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()

	//TODO refactor more so we can have more common functions
	if r.GetType() == reconciliation.ApplicationType {
		return generateForApplication(r)
	} else if r.GetType() == reconciliation.JobType {
		return generateForSKIPJob(r)
	} else {
		err := fmt.Errorf("unsupported type %s in service account", r.GetType())
		ctxLog.Error(err, "Failed to generate service account")
		return err
	}
}
