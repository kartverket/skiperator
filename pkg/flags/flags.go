package flags

import (
	"fmt"
	"os"
	"strconv"
)

const envPrefix = "SKIPERATOR"

var FeatureFlags *Features

// Features contains various Skiperator knobs which can be enabled/disabled runtime with
// environment variables.
type Features struct {
	DisablePodTopologySpreadConstraints bool
}

func init() {
	FeatureFlags = &Features{
		DisablePodTopologySpreadConstraints: getEnvWithFallback("DISABLE_PTSC", false),
	}
}

func getEnvWithFallback(name string, defaultValue bool) bool {
	key := fmt.Sprintf("%s_%s", envPrefix, name)

	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	v, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}

	return v
}
