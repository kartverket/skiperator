package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInternalIsCaseInsensitive(t *testing.T) {
	assert.True(t, IsInternal("API.SKIP.STATKART.NO"))
	assert.True(t, IsInternal("api.KARTVERKET-INTERN.CLOUD"))
}

func TestIsInternalMatchesOnlyTheRealDomainSuffix(t *testing.T) {
	assert.True(t, IsInternal("app.skip.statkart.no"))
	assert.True(t, IsInternal("foo.bar.kartverket-intern.cloud"))

	// Lookalike hosts that merely contain the internal domain must not match.
	assert.False(t, IsInternal("app.skip.statkart.no.attacker.com"))
	assert.False(t, IsInternal("skip.statkart.no.evil.example"))
	// Unescaped dot previously let any character match before "cloud".
	assert.False(t, IsInternal("api.kartverket-internXcloud"))
}
