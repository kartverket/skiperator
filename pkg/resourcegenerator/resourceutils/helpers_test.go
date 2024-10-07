package resourceutils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetImageVersionNoTag(t *testing.T) {
	imageString := "image"
	expectedImageString := "latest"

	actualImageString := GetImageVersion(imageString)

	assert.Equal(t, expectedImageString, actualImageString)
}

func TestGetImageVersionLatestTag(t *testing.T) {
	imageString := "image:latest"
	expectedImageString := "latest"

	actualImageString := GetImageVersion(imageString)

	assert.Equal(t, expectedImageString, actualImageString)
}

func TestGetImageVersionVersionTag(t *testing.T) {
	versionImageString := "image:1.2.3"
	devImageString := "image:1.2.3-dev-123abc"
	expectedVersionImageString := "1.2.3"
	expectedDevImageString := "1.2.3-dev-123abc"

	actualVersionImageString := GetImageVersion(versionImageString)
	actualDevImageString := GetImageVersion(devImageString)

	assert.Equal(t, expectedVersionImageString, actualVersionImageString)
	assert.Equal(t, expectedDevImageString, actualDevImageString)

}

func TestGetImageVersionShaTag(t *testing.T) {
	imageString := "ghcr.io/org/repo@sha256:54d7ea8b48d0e7569766e0e10b9e38da778a5f65d764168dd7db76a37d6b8"
	expectedImageString := "54d7ea8b48d0e7569766e0e10b9e38da778a5f65d764168dd7db76a37d6b8"

	actualImageString := GetImageVersion(imageString)

	assert.Equal(t, expectedImageString, actualImageString)
}
