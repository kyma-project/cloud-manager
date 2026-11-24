package types

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Provider interface {
	BoolVariation(ctx context.Context, flagKey string, defaultValue bool) bool
	IntVariation(ctx context.Context, flagKey string, defaultValue int) int
	Float64Variation(ctx context.Context, flagKey string, defaultValue float64) float64
	StringVariation(ctx context.Context, flagKey string, defaultValue string) string
	JSONArrayVariation(ctx context.Context, flagKey string, defaultValue []any) []any
	JSONVariation(ctx context.Context, flagKey string, defaultValue map[string]any) map[string]any
}

// ReasonAwareProvider extends Provider with reason-aware boolean evaluation.
// byRule is true iff an explicit targeting rule decided the value, rather than the flag default.
type ReasonAwareProvider interface {
	Provider
	BoolVariationReason(ctx context.Context, flagKey string, defaultValue bool) (value bool, byRule bool)
}

type LandscapeName = string

const (
	LandscapeDev   LandscapeName = "dev"
	LandscapeStage LandscapeName = "stage"
	LandscapeProd  LandscapeName = "prod"
)

type FeatureName = string

const (
	FeatureUnknown FeatureName = "unknown"

	// FeatureFoundation represents foundational/non-feature-specific resources (always enabled, independent of config).
	FeatureFoundation FeatureName = "foundation"

	FeatureCertificate       FeatureName = "certificate"
	FeatureNfs               FeatureName = "nfs"
	FeatureNfsBackup         FeatureName = "nfsBackup"
	FeaturePeering           FeatureName = "peering"
	FeatureRedis             FeatureName = "redis"
	FeatureRedisCluster      FeatureName = "rediscluster"
	FeatureAzureManagedRedis FeatureName = "azureManagedRedis"
	FeatureVpcDnsLink        FeatureName = "vpcdnslink"
	FeatureWAF               FeatureName = "waf"
)

// AllFeatureNames returns every defined FeatureName constant in the canonical order.
// Callers (e.g. the allowlist matrix test) use this to stay in sync when new features are added.
func AllFeatureNames() []FeatureName {
	return []FeatureName{
		FeatureFoundation,
		FeatureUnknown,
		FeatureCertificate,
		FeatureNfs,
		FeatureNfsBackup,
		FeaturePeering,
		FeatureRedis,
		FeatureRedisCluster,
		FeatureAzureManagedRedis,
		FeatureVpcDnsLink,
		FeatureWAF,
	}
}

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

type FeatureAwareObject interface {
	client.Object
	// SpecificToFeature returns the FeatureName this resource belongs to. If not specific to a
	// certain feature (foundational/non-feature-specific resources) it MUST return FeatureFoundation.
	// It must never return "" or FeatureUnknown — those are untagged sentinels and would break
	// apiEnabled gate correctness (enforced by TestEverySkrCrdIsTagged).
	SpecificToFeature() FeatureName
}

type ProviderAwareObject interface {
	client.Object
	// SpecificToProviders returns slice of supported providers as defined in the Shoot resource
	// if not specific to certain providers and can work with all providers it should return nil
	SpecificToProviders() []string
}
