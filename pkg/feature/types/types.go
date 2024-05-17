package types

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
	KeyAllKindGroups   Key = "allKindGroups"
	KeyObjKindGroup    Key = "objKindGroup"
	KeyCrdKindGroup    Key = "crdKindGroup"
	KeyBusolaKindGroup Key = "busolaKindGroup"
)

type Feature[T any] interface {
	Value(ctx context.Context) T
}

type FeatureAware interface {
	SpecificToFeature() FeatureName
}

type ProviderAware interface {
	SpecificToProviders() []string
}
