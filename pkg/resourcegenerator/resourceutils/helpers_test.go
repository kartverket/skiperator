package resourceutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersions(t *testing.T) {
	testCases := []struct {
		imageString   string
		expectedValue string
	}{
		{"image", "latest"},
		{"image:latest", "latest"},
		{"image:1.2.3-dev-123abc", "1.2.3-dev-123abc"},
		{"image:1.2.3", "1.2.3"},
		{"ghcr.io/org/repo@sha256:54d7ea8b48d0e7569766e0e10b9e38da778a5f65d764168dd7db76a37d6b8", "54d7ea8b48d0e7569766e0e10b9e38da778a5f65d764168dd7db76a37d6b8"},
		{"ghcr.io/org/one-app:sha-b15dc91c27ad2387bea81294593d5ce5a686bcc4@sha256:3cda54f1d25458f25fdde0398130da57a4ebb4a4cd759bc49035b7ebf9d83619", "sha-b15dc91c27ad2387bea81294593d5ce5a686bcc4"},
		{"ghcr.io/org/another-app:3fb7048", "3fb7048"},
		{"ghcr.io/org/some-team/third-app:v1.2.54", "v1.2.54"},
		{"ghcr.io/org/another-team/fourth-app:4.0.0.rc-36", "4.0.0.rc-36"},
		{"ghcr.io/org/another-team/fifth-app:4.0.0.rc-36-master-latest", "4.0.0.rc-36-master-latest"},
		{"ghcr.io/kartverket/vulnerability-disclosure-program@sha256:ab85022d117168585bdedc71cf9c67c3ca327533dc7cd2c5bcc42a83f308ea5d", "latest"},
		{"nginxinc/nginx-unprivileged:1.20.0-alpine", "1.20.0-alpine"},
		{"foo/bar:1.2.3+build.4", "1.2.3-build.4"},
		{"foo/bar:1.2.3+somethingLongXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", "1.2.3-somethingLongXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"},
		{"foo/bar:-1.2.3", "1.2.3"},
		{"foo/bar:__1.2.3", "1.2.3"},
		{"foo/bar:.1.2.3", "1.2.3"},
	}

	for _, tc := range testCases {
		t.Run(tc.imageString, func(t *testing.T) {
			actualValue := GetImageVersion(tc.imageString)
			assert.Equal(t, tc.expectedValue, actualValue)
		})
	}
}