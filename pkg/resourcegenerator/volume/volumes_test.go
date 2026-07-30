package volume

import (
	"strings"
	"testing"

	"github.com/kartverket/skiperator/api/common/podtypes"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/validation"
)

func TestSourceVolumeName(t *testing.T) {
	longName := strings.Repeat("a", 80)

	tests := map[string]struct {
		file podtypes.FilesFrom
		want string
	}{
		"config_map": {
			file: podtypes.FilesFrom{ConfigMap: "shared-name"},
			want: "cm-shared-name",
		},
		"secret": {
			file: podtypes.FilesFrom{Secret: "shared-name"},
			want: "sec-shared-name",
		},
		"empty_dir": {
			file: podtypes.FilesFrom{EmptyDir: "shared-name"},
			want: "ed-shared-name",
		},
		"persistent_volume_claim": {
			file: podtypes.FilesFrom{PersistentVolumeClaim: "shared-name"},
			want: "pvc-shared-name",
		},
		"tmp_empty_dir": {
			file: podtypes.FilesFrom{EmptyDir: tmpVolumeName},
			want: tmpVolumeName,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			volumeName := sourceVolumeName(tc.file)

			assert.Equal(t, tc.want, volumeName)
			assert.Empty(t, validation.IsDNS1123Label(volumeName))
		})
	}

	t.Run("long_source_name", func(t *testing.T) {
		volumeName := sourceVolumeName(podtypes.FilesFrom{ConfigMap: longName})

		assert.Regexp(t, `-[0-9a-f]{8}$`, volumeName)
		assert.Len(t, volumeName, validation.DNS1123LabelMaxLength)
		assert.LessOrEqual(t, len(volumeName), validation.DNS1123LabelMaxLength)
		assert.Empty(t, validation.IsDNS1123Label(volumeName))
		assert.True(t, strings.HasPrefix(volumeName, configMapPrefix))
	})

	t.Run("normalized_names_do_not_collide", func(t *testing.T) {
		dottedName := sourceVolumeName(podtypes.FilesFrom{ConfigMap: "shared.name"})
		dashedName := sourceVolumeName(podtypes.FilesFrom{ConfigMap: "shared-name"})
		underscoredName := sourceVolumeName(podtypes.FilesFrom{ConfigMap: "shared_name"})

		assert.Regexp(t, `^cm-shared-name-[0-9a-f]{8}$`, dottedName)
		assert.NotEqual(t, dashedName, dottedName)
		assert.NotEqual(t, dottedName, underscoredName)
		assert.Regexp(t, `^cm-shared-name-[0-9a-f]{8}$`, underscoredName)
		assert.Empty(t, validation.IsDNS1123Label(dottedName))
		assert.Empty(t, validation.IsDNS1123Label(dashedName))
		assert.Empty(t, validation.IsDNS1123Label(underscoredName))
	})
}

func TestVolumeNamesMatchVolumeMountNames(t *testing.T) {
	filesFrom := []podtypes.FilesFrom{
		{MountPath: "/config", ConfigMap: strings.Repeat("a", 80)},
		{MountPath: "/secret", Secret: "shared.name"},
		{MountPath: "/pvc", PersistentVolumeClaim: "shared-name"},
	}

	volumes := GetPodVolumes(filesFrom)
	mounts := GetContainerVolumeMounts(filesFrom)

	volumeNames := map[string]struct{}{}
	for _, volume := range volumes {
		volumeNames[volume.Name] = struct{}{}
	}

	for _, mount := range mounts {
		_, exists := volumeNames[mount.Name]
		assert.Truef(t, exists, "volume mount %q has no matching volume", mount.Name)
	}
}
