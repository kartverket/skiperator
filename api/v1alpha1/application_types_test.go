package v1alpha1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApplicationIngressTargetPort(t *testing.T) {
	proxyPort := int32(8443)

	tests := map[string]struct {
		app  Application
		want int
	}{
		"no extra containers falls back to spec.port": {
			app:  Application{Spec: ApplicationSpec{Port: 8080}},
			want: 8080,
		},
		"extra container without ingressPort falls back to spec.port": {
			app: Application{Spec: ApplicationSpec{Port: 8080, ExtraContainers: []ContainerSpec{
				{Name: "logger", Image: "logger:1.0"},
			}}},
			want: 8080,
		},
		"fronting container ingressPort takes precedence": {
			app: Application{Spec: ApplicationSpec{Port: 8080, ExtraContainers: []ContainerSpec{
				{Name: "logger", Image: "logger:1.0"},
				{Name: "auth-proxy", Image: "proxy:1.0", IngressPort: &proxyPort},
			}}},
			want: 8443,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.app.IngressTargetPort())
		})
	}
}

func TestApplicationFillDefaultsStatusIsIdempotent(t *testing.T) {
	app := &Application{}

	app.FillDefaultsStatus()
	firstStatus := app.Status.DeepCopy()

	time.Sleep(time.Nanosecond)
	app.FillDefaultsStatus()

	assert.Equal(t, firstStatus.Summary, app.Status.Summary)
	assert.NotNil(t, app.Status.SubResources)
	assert.NotNil(t, app.Status.Conditions)
}
