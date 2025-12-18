package usage

const (
	// Common subsystem that all metrics are registered under
	metricSubsystem = "skiperator"

	// Organizational labels
	labelTeam     = "team"
	labelDivision = "division"

	// CRD label
	labelType = "type"

	// Skiperator CRD types
	typeApplication = "Application"
	typeSKIPJob     = "SKIPJob"
	typeRouting     = "Routing"

	// For massaging data
	countKey     = "count"
	unknownValue = "unknown"

	// Ignore label
	ignoreLabel = "skiperator.kartverket.no/ignore"
)
