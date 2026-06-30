package pod

import (
	"testing"

	"github.com/kartverket/skiperator/api/common/podtypes"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestCreateExtraContainers_SplitsByType(t *testing.T) {
	specs := []podtypes.ContainerSpec{
		{Name: "logging-agent", Image: "logging:1.0"},
		{Name: "config-loader", Image: "loader:1.0", Type: podtypes.ContainerTypeInit},
	}

	sidecars, initContainers, _ := CreateExtraContainers(specs, PodOpts{})

	assert.Len(t, sidecars, 1)
	assert.Equal(t, "logging-agent", sidecars[0].Name)
	assert.Nil(t, sidecars[0].RestartPolicy)

	assert.Len(t, initContainers, 1)
	assert.Equal(t, "config-loader", initContainers[0].Name)
	if assert.NotNil(t, initContainers[0].RestartPolicy) {
		assert.Equal(t, corev1.ContainerRestartPolicyAlways, *initContainers[0].RestartPolicy)
	}
}

func TestCreateExtraContainers_EnforcesSecurityContext(t *testing.T) {
	sidecars, _, _ := CreateExtraContainers([]podtypes.ContainerSpec{
		{Name: "side", Image: "side:1.0"},
	}, PodOpts{})

	sc := sidecars[0].SecurityContext
	if assert.NotNil(t, sc) {
		assert.False(t, *sc.AllowPrivilegeEscalation)
		assert.True(t, *sc.ReadOnlyRootFilesystem)
		assert.True(t, *sc.RunAsNonRoot)
		assert.Contains(t, sc.Capabilities.Drop, corev1.Capability("ALL"))
	}
}

func TestCreateExtraContainers_VolumesDedupedByAppendUniqueVolumes(t *testing.T) {
	specs := []podtypes.ContainerSpec{
		{Name: "a", Image: "a:1.0", FilesFrom: []podtypes.FilesFrom{{MountPath: "/etc/cfg", ConfigMap: "shared"}}},
		{Name: "b", Image: "b:1.0", FilesFrom: []podtypes.FilesFrom{{MountPath: "/etc/cfg", ConfigMap: "shared"}}},
	}

	_, _, volumes := CreateExtraContainers(specs, PodOpts{})

	// Merging through AppendUniqueVolumes (as the deployment/statefulset callers
	// do) is the single dedup point across the whole pod.
	merged := AppendUniqueVolumes(nil, volumes...)

	names := map[string]int{}
	for _, v := range merged {
		names[v.Name]++
	}
	for name, count := range names {
		assert.Equalf(t, 1, count, "volume %q duplicated", name)
	}
	assert.Equal(t, 1, names["shared"])
}

func TestCreateExtraContainers_Empty(t *testing.T) {
	sidecars, initContainers, volumes := CreateExtraContainers(nil, PodOpts{})
	assert.Empty(t, sidecars)
	assert.Empty(t, initContainers)
	assert.Empty(t, volumes)
}

func TestCreateExtraContainers_ImagePullPolicyFromOpts(t *testing.T) {
	specs := []podtypes.ContainerSpec{{Name: "side", Image: "side:1.0"}}

	remote, _, _ := CreateExtraContainers(specs, PodOpts{})
	assert.Equal(t, corev1.PullAlways, remote[0].ImagePullPolicy)

	local, _, _ := CreateExtraContainers(specs, PodOpts{LocalBuiltImages: true})
	assert.Equal(t, corev1.PullNever, local[0].ImagePullPolicy)
}

func TestAppendUniqueVolumes(t *testing.T) {
	existing := []corev1.Volume{{Name: "tmp"}, {Name: "shared"}}
	result := AppendUniqueVolumes(existing,
		corev1.Volume{Name: "tmp"},    // duplicate, skipped
		corev1.Volume{Name: "shared"}, // duplicate, skipped
		corev1.Volume{Name: "new"},    // appended
	)

	assert.Len(t, result, 3)
	names := []string{result[0].Name, result[1].Name, result[2].Name}
	assert.Equal(t, []string{"tmp", "shared", "new"}, names)
}
