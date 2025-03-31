package generator

import (
	"fmt"

	"github.com/kartverket/skiperator/v3/pkg/reconciliation"
)

type genFunc = func(r reconciliation.Reconciliation) error

type MultiGenerator struct {
	generators map[reconciliation.ObjectType]genFunc
}

func NewMulti() *MultiGenerator {
	return &MultiGenerator{
		generators: map[reconciliation.ObjectType]genFunc{},
	}
}

// Register will ensure that the supplied generator will be used for a given reconciliation object type.
func (g *MultiGenerator) Register(objectType reconciliation.ObjectType, generator genFunc) {
	if generator == nil {
		panic("generator cannot be nil")
	}

	g.generators[objectType] = generator
}

// Generate will look up the reconciliation object type and generate the resource using
// the appropriate generator function.
func (g *MultiGenerator) Generate(r reconciliation.Reconciliation, resourceType string) error {
	generator, found := g.generators[r.GetType()]
	if !found {
		err := fmt.Errorf("unsupported type %s for resource %s", r.GetType(), resourceType)
		r.GetLogger().Error(err, "failed to generate resource", "resourceType", resourceType, "reconciliationType", r.GetType())
		return err
	}

	return generator(r)
}
