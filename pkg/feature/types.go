package feature

import "context"

type LandscapeName = string

const (
	LandscapeDev   LandscapeName = "dev"
	LandscapeStage LandscapeName = "stage"
	LandscapeProd  LandscapeName = "prod"
)

type FeatureName = string

const (
	FeatureNfs       FeatureName = "nfs"
	FeatureNfsBackup FeatureName = "nfsBackup"
	FeaturePeering   FeatureName = "peering"
)

type PlaneName = string

const (
	PlaneSkr PlaneName = "skr"
	PlaneKcp PlaneName = "kcp"
)

type Key = string

const (
	KeyLandscape       Key = "landscape"
	KeyFeature         Key = "feature"
	KeyPlane           Key = "plane"
	KeyProvider        Key = "provider"
	KeyBrokerPlan      Key = "brokerPlan"
	KeyGlobalAccount   Key = "globalAccount"
	KeySubAccount      Key = "subAccount"
	KeyKyma            Key = "kyma"
	KeyShoot           Key = "shoot"
	KeyRegion          Key = "region"
	KeyKindGroup       Key = "kindGroup"
	KeyCrdKindGroup    Key = "crdKindGroup"
	KeyBusolaKindGroup Key = "busolaKindGroup"
)

var loggerKeys = []Key{
	KeyFeature,
	KeyPlane,
	KeyProvider,
	KeyBrokerPlan,
	KeyGlobalAccount,
	KeySubAccount,
	KeyKyma,
	KeyShoot,
	KeyRegion,
	KeyKindGroup,
}

type Feature[T any] interface {
	Value(ctx context.Context) T
}
