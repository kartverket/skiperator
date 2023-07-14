package cleaner

import (
	"github.com/dlclark/regexp2"
	"github.com/go-logr/logr"
	"time"
)

var instanaLabelMatcher = regexp2.MustCompile("(?!^\\binstana-autotrace$\\b)(\\binstana\\b)", regexp2.Multiline)

func init() {
	instanaLabelMatcher.MatchTimeout = 5 * time.Second
}

func InstanaCleaner(log *logr.Logger) *DeploymentCleaner {
	return &DeploymentCleaner{
		RemovePrefixPredicate: func(candidate string) bool {
			match, err := instanaLabelMatcher.MatchString(candidate)
			if err != nil {
				log.Error(err, "could not match string", "candidate", candidate)
				return false
			}
			return match
		},
		RemoveEnvVars: []string{
			"LD_PRELOAD",
			"ACE_ENABLE_OPEN_TRACING",
			"MQ_ENABLE_OPEN_TRACING",
			"NODE_OPTIONS",
		},
		RemoveVolumes:        []string{"instana-instrumentation-volume"},
		RemoveInitContainers: []string{"instana-instrumentation-init"},
	}
}
