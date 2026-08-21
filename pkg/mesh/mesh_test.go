package mesh

import "testing"

func TestModeFromLabels(t *testing.T) {
	tests := map[string]struct {
		labels map[string]string
		want   Mode
	}{
		"no labels": {nil, ModeNone},
		"empty rev": {map[string]string{RevisionLabel: ""}, ModeNone},
		"revision":  {map[string]string{RevisionLabel: "asm-stable"}, ModeSidecar},
		"ambient":   {map[string]string{DataplaneModeLabel: AmbientDataplaneMode}, ModeAmbient},
		"mode none": {map[string]string{DataplaneModeLabel: "none"}, ModeNone},
		// Istio keeps sidecar pods out of the ambient mesh, so the sidecar
		// label decides when a namespace carries both.
		"both labels": {map[string]string{RevisionLabel: "asm-stable", DataplaneModeLabel: AmbientDataplaneMode}, ModeSidecar},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := ModeFromLabels(test.labels); got != test.want {
				t.Errorf("ModeFromLabels() = %q, want %q", got, test.want)
			}
		})
	}
}
