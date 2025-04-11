package feature

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/objkind"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

var keyToBuilderMethod = map[types.Key]func(ContextBuilder, string) ContextBuilder{
	types.KeyFeature:         ContextBuilder.Feature,
	types.KeyPlane:           ContextBuilder.Plane,
	types.KeyProvider:        ContextBuilder.Provider,
	types.KeyBrokerPlan:      ContextBuilder.BrokerPlan,
	types.KeyGlobalAccount:   ContextBuilder.GlobalAccount,
	types.KeySubAccount:      ContextBuilder.SubAccount,
	types.KeyRegion:          ContextBuilder.Region,
	types.KeyKyma:            ContextBuilder.Kyma,
	types.KeyShoot:           ContextBuilder.Shoot,
	types.KeyObjKindGroup:    ContextBuilder.ObjKindGroup,
	types.KeyCrdKindGroup:    ContextBuilder.CrdKindGroup,
	types.KeyBusolaKindGroup: ContextBuilder.BusolaKindGroup,
}

// ContextReader ===============================================================

func NewContextReaderFromFFCtx(ffCtx ffcontext.Context) *ContextReader {
	return &ContextReader{ffCtx: ffCtx}
}

func NewContextReaderFromCtx(ctx context.Context) *ContextReader {
	ffCtx := MustContextFromCtx(ctx)
	return NewContextReaderFromFFCtx(ffCtx)
}

type ContextReader struct {
	ffCtx ffcontext.Context
}

func (r *ContextReader) Landscape() string {
	return util.CastInterfaceToString(r.ffCtx.GetCustom()[types.KeyLandscape])
}

func (r *ContextReader) Plane() string {
	return util.CastInterfaceToString(r.ffCtx.GetCustom()[types.KeyPlane])
}

func (r *ContextReader) Feature() string {
	return util.CastInterfaceToString(r.ffCtx.GetCustom()[types.KeyFeature])
}

func (r *ContextReader) Provider() string {
	return util.CastInterfaceToString(r.ffCtx.GetCustom()[types.KeyProvider])
}

func (r *ContextReader) BrokerPlan() string {
	return util.CastInterfaceToString(r.ffCtx.GetCustom()[types.KeyBrokerPlan])
}

func (r *ContextReader) GlobalAccount() string {
	return util.CastInterfaceToString(r.ffCtx.GetCustom()[types.KeyGlobalAccount])
}

func (r *ContextReader) SubAccount() string {
	return util.CastInterfaceToString(r.ffCtx.GetCustom()[types.KeySubAccount])
}

func (r *ContextReader) Region() string {
	return util.CastInterfaceToString(r.ffCtx.GetCustom()[types.KeyRegion])
}

func (r *ContextReader) Kyma() string {
	return util.CastInterfaceToString(r.ffCtx.GetCustom()[types.KeyKyma])
}

func (r *ContextReader) Shoot() string {
	return util.CastInterfaceToString(r.ffCtx.GetCustom()[types.KeyShoot])
}

func (r *ContextReader) ObjKindGroup() string {
	return util.CastInterfaceToString(r.ffCtx.GetCustom()[types.KeyObjKindGroup])
}

func (r *ContextReader) CrdKindGroup() string {
	return util.CastInterfaceToString(r.ffCtx.GetCustom()[types.KeyCrdKindGroup])
}

func (r *ContextReader) BusolaKindGroup() string {
	return util.CastInterfaceToString(r.ffCtx.GetCustom()[types.KeyBusolaKindGroup])
}

// ContextBuilder ===============================================================

type ContextBuilder interface {
	Build(ctx context.Context) context.Context
	FFCtx() ffcontext.Context

	LoadFromKyma(u *unstructured.Unstructured) ContextBuilder
	LoadFromScope(scope *cloudcontrolv1beta1.Scope) ContextBuilder
	LoadFromMap(m map[string]interface{}) ContextBuilder
	Landscape(v string) ContextBuilder
	Feature(v string) ContextBuilder
	FeatureFromObject(obj client.Object, scheme *runtime.Scheme) ContextBuilder
	Plane(v string) ContextBuilder
	Provider(v string) ContextBuilder
	BrokerPlan(v string) ContextBuilder
	GlobalAccount(v string) ContextBuilder
	SubAccount(v string) ContextBuilder
	Kyma(v string) ContextBuilder
	Shoot(v string) ContextBuilder
	Region(v string) ContextBuilder
	// KindGroup in the format kind.group, ie awsnfsvolume.cloud-resources.kyma-project.io
	ObjKindGroup(v string) ContextBuilder
	CrdKindGroup(v string) ContextBuilder
	BusolaKindGroup(v string) ContextBuilder

	// KindsFromObject finds the obj's GVK and sets objKindGroup, crdKindGroup, and busolaKindGroup keys
	// to lower(kind).lower(group) of the GVK.
	// If the obj is CustomResourceDefinition, it will set KindGroup to the CRD defining values
	// lower(spec.names.kind).lower(spec.group).
	KindsFromObject(obj client.Object, scheme *runtime.Scheme) ContextBuilder

	Custom(key string, value interface{}) ContextBuilder
	Std(key types.Key, value string) ContextBuilder
}

func ContextFromCtx(ctx context.Context) ffcontext.Context {
	ffCtx, ok := ctx.Value(contextKey).(ffcontext.Context)
	if !ok {
		return nil
	}
	return ffCtx
}

func MustContextFromCtx(ctx context.Context) ffcontext.Context {
	ffCtx := ContextFromCtx(ctx)
	if ffCtx == nil {
		ffCtx = ffcontext.NewEvaluationContext(uuid.NewString())
	}
	return ffCtx
}

func ContextBuilderFromCtx(ctx context.Context) ContextBuilder {
	var b ffcontext.EvaluationContextBuilder
	ffCtx := ContextFromCtx(ctx)
	if ffCtx != nil {
		b = ffcontext.NewEvaluationContextBuilder(ffCtx.GetKey())
		for k, v := range ffCtx.GetCustom() {
			b = b.AddCustom(k, v)
		}
	} else {
		b = ffcontext.NewEvaluationContextBuilder(uuid.NewString())
	}
	return &contextBuilderImpl{builder: b}
}

type contextBuilderImpl struct {
	builder ffcontext.EvaluationContextBuilder
}

func (b *contextBuilderImpl) Build(ctx context.Context) context.Context {
	ffCtx := b.FFCtx()
	return context.WithValue(ctx, contextKey, ffCtx)
}

func (b *contextBuilderImpl) FFCtx() ffcontext.Context {
	ffCtx := b.builder.Build()
	ffCtx.GetCustom()[types.KeyAllKindGroups] = fmt.Sprintf("%s,%s,%s",
		util.CastInterfaceToString(ffCtx.GetCustom()[types.KeyObjKindGroup]),
		util.CastInterfaceToString(ffCtx.GetCustom()[types.KeyCrdKindGroup]),
		util.CastInterfaceToString(ffCtx.GetCustom()[types.KeyBusolaKindGroup]),
	)
	return ffCtx
}

func (b *contextBuilderImpl) LoadFromKyma(u *unstructured.Unstructured) ContextBuilder {
	b.Kyma("")
	b.BrokerPlan("")
	b.GlobalAccount("")
	b.SubAccount("")
	b.Region("")
	b.Shoot("")

	if u != nil && u.Object != nil {
		b.Kyma(u.GetName())
		if labels := u.GetLabels(); len(labels) > 0 {
			b.BrokerPlan(labels[cloudcontrolv1beta1.LabelScopeBrokerPlanName])
			b.GlobalAccount(labels[cloudcontrolv1beta1.LabelScopeGlobalAccountId])
			b.SubAccount(labels[cloudcontrolv1beta1.LabelScopeSubaccountId])
			b.Region(labels[cloudcontrolv1beta1.LabelScopeRegion])
			b.Shoot(labels[cloudcontrolv1beta1.LabelScopeShootName])
		}
	}
	return b
}

func (b *contextBuilderImpl) LoadFromScope(scope *cloudcontrolv1beta1.Scope) ContextBuilder {
	b.Provider("")
	if scope != nil {
		b.Provider(string(scope.Spec.Provider))
	}
	return b
}

func (b *contextBuilderImpl) LoadFromMap(m map[string]interface{}) ContextBuilder {
	for k, v := range m {
		b.Custom(k, v)
	}
	return b
}

func (b *contextBuilderImpl) Landscape(v string) ContextBuilder {
	b.builder = b.builder.AddCustom(types.KeyLandscape, v)
	return b
}

func (b *contextBuilderImpl) Feature(v string) ContextBuilder {
	b.builder = b.builder.AddCustom(types.KeyFeature, v)
	return b
}

func (b *contextBuilderImpl) FeatureFromObject(obj client.Object, scheme *runtime.Scheme) ContextBuilder {
	return b.Feature(ObjectToFeature(obj, scheme))
}

func (b *contextBuilderImpl) Plane(v string) ContextBuilder {
	b.builder = b.builder.AddCustom(types.KeyPlane, v)
	return b
}

func (b *contextBuilderImpl) Provider(v string) ContextBuilder {
	b.builder = b.builder.AddCustom(types.KeyProvider, v)
	return b
}

func (b *contextBuilderImpl) BrokerPlan(v string) ContextBuilder {
	b.builder = b.builder.AddCustom(types.KeyBrokerPlan, v)
	return b
}

func (b *contextBuilderImpl) GlobalAccount(v string) ContextBuilder {
	b.builder = b.builder.AddCustom(types.KeyGlobalAccount, v)
	return b
}

func (b *contextBuilderImpl) SubAccount(v string) ContextBuilder {
	b.builder = b.builder.AddCustom(types.KeySubAccount, v)
	return b
}

func (b *contextBuilderImpl) Kyma(v string) ContextBuilder {
	b.builder = b.builder.AddCustom(types.KeyKyma, v)
	return b
}

func (b *contextBuilderImpl) Shoot(v string) ContextBuilder {
	b.builder = b.builder.AddCustom(types.KeyShoot, v)
	return b
}

func (b *contextBuilderImpl) Region(v string) ContextBuilder {
	b.builder = b.builder.AddCustom(types.KeyRegion, v)
	return b
}

func (b *contextBuilderImpl) ObjKindGroup(v string) ContextBuilder {
	b.builder = b.builder.AddCustom(types.KeyObjKindGroup, v)
	return b
}

func (b *contextBuilderImpl) CrdKindGroup(v string) ContextBuilder {
	b.builder = b.builder.AddCustom(types.KeyCrdKindGroup, v)
	return b
}

func (b *contextBuilderImpl) BusolaKindGroup(v string) ContextBuilder {
	b.builder = b.builder.AddCustom(types.KeyBusolaKindGroup, v)
	return b
}

func (b *contextBuilderImpl) KindsFromObject(obj client.Object, scheme *runtime.Scheme) ContextBuilder {
	b.ObjKindGroup("")
	b.CrdKindGroup("")
	b.BusolaKindGroup("")

	kindInfo := objkind.ObjectKinds(obj, scheme)
	if !kindInfo.ObjOK {
		return b
	}

	b.ObjKindGroup(strings.ToLower(kindInfo.ObjGK.String()))
	if kindInfo.CrdOK {
		b.CrdKindGroup(strings.ToLower(kindInfo.CrdGK.String()))
	}
	if kindInfo.BusolaOK {
		b.BusolaKindGroup(strings.ToLower(kindInfo.BusolaGK.String()))
	}

	return b
}

func (b *contextBuilderImpl) Custom(key string, value interface{}) ContextBuilder {
	b.builder = b.builder.AddCustom(key, value)
	return b
}

func (b *contextBuilderImpl) Std(key types.Key, value string) ContextBuilder {
	keyToBuilderMethod[key](b, value)
	return b
}
