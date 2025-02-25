package auth

import (
	"github.com/kartverket/skiperator/api/v1alpha1/digdirator"
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
)

type AutoLoginConfig struct {
	Spec         istiotypes.AutoLogin
	IsEnabled    bool
	IgnorePaths  []string
	ProviderURIs digdirator.DigdiratorURIs
	AuthScopes   []string
	ClientSecret string
}
