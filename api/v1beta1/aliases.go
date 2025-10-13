package v1beta1

import (
	"github.com/kartverket/skiperator/api/common"
	commontypes "github.com/kartverket/skiperator/api/common"
	commondigdiratortypes "github.com/kartverket/skiperator/api/common/digdirator"
	commonistiotypes "github.com/kartverket/skiperator/api/common/istiotypes"
	commonpodtypes "github.com/kartverket/skiperator/api/common/podtypes"
	commonprometheustypes "github.com/kartverket/skiperator/api/common/prometheus"
)

// ===== digdirator aliases =====

// digdirator
type DigdiratorName = commondigdiratortypes.DigdiratorName
type DigdiratorInfo = commondigdiratortypes.DigdiratorInfo
type DigdiratorClient = commondigdiratortypes.DigdiratorClient
type DigdiratorProvider = commondigdiratortypes.DigdiratorProvider

// ID Porten
type IDPorten = commondigdiratortypes.IDPorten

// maskinporten
type Maskinporten = commondigdiratortypes.Maskinporten
type MaskinportenClient = commondigdiratortypes.MaskinportenClient

// ===== istiotypes aliases =====
// istiosettings
type IstioSettingsBase = commonistiotypes.IstioSettingsBase
type IstioSettingsApplication = commonistiotypes.IstioSettingsApplication

// jwt authentication
type RequestAuthentication = commonistiotypes.RequestAuthentication
type ClaimToHeader = commonistiotypes.ClaimToHeader

// ===== podtypes aliases =====
// access policies
type AccessPolicy = commonpodtypes.AccessPolicy
type InboundPolicy = commonpodtypes.InboundPolicy
type OutboundPolicy = commonpodtypes.OutboundPolicy
type InternalRule = commonpodtypes.InternalRule
type ExternalRule = commonpodtypes.ExternalRule
type ExternalPort = commonpodtypes.ExternalPort

// files from env
type EnvFrom = commonpodtypes.EnvFrom
type FilesFrom = commonpodtypes.FilesFrom

// GCP
type GCP = commonpodtypes.GCP
type Auth = commonpodtypes.Auth
type CloudSQLProxySettings = commonpodtypes.CloudSQLProxySettings

// internal port
type InternalPort = commonpodtypes.InternalPort

// pod settings
type PodSettings = commonpodtypes.PodSettings

// probe
type Probe = commonpodtypes.Probe

// resource requirements
type ResourceRequirements = commonpodtypes.ResourceRequirements

// ===== prometheus config =====
type PrometheusConfig = commonprometheustypes.PrometheusConfig

// ===== cron settings aliases =====
type CronSettings = commontypes.CronSettings

// ===== job settings aliases =====
type JobSettings = commontypes.JobSettings

// ===== skiperator status =====
type SkiperatorStatus = common.SkiperatorStatus
type Status = common.Status
type StatusNames = common.StatusNames

const (
	SYNCED        = common.SYNCED
	PROGRESSING   = common.PROGRESSING
	ERROR         = common.ERROR
	PENDING       = common.PENDING
	READY         = common.READY
	INVALIDCONFIG = common.INVALIDCONFIG
)
