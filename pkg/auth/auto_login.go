package auth

import (
	"github.com/kartverket/skiperator/api/v1alpha1/digdirator"
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
)

type AutoLoginConfig struct {
	Spec         istiotypes.AutoLogin
	IsEnabled    bool
	IgnorePaths  []string
	ProviderInfo digdirator.DigdiratorInfo
	AuthScopes   []string
	ClientSecret string
}
