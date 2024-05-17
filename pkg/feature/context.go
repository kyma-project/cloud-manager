package feature

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"strings"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

var keyToBuilderMethod = map[types.Key]func(ContextBuilder, string) ContextBuilder{
	types.KeyFeature:       ContextBuilder.Feature,
	types.KeyPlane:         ContextBuilder.Plane,
	types.KeyProvider:      ContextBuilder.Provider,
	types.KeyBrokerPlan:    ContextBuilder.BrokerPlan,
	types.KeyGlobalAccount: ContextBuilder.GlobalAccount,
	types.KeySubAccount:    ContextBuilder.SubAccount,
	types.KeyKyma:          ContextBuilder.Kyma,
	types.KeyShoot:         ContextBuilder.Shoot,
	types.KeyObjKindGroup:  ContextBuilder.ObjKindGroup,
}

type ContextBuilder interface {
	Build(ctx context.Context) context.Context
	FFCtx() ffcontext.Context

	LoadFromKyma(o client.Object) ContextBuilder
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
		ffCtx = ffcontext.NewEvaluationContext("")
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
		b = ffcontext.NewEvaluationContextBuilder("")
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

func (b *contextBuilderImpl) LoadFromKyma(o client.Object) ContextBuilder {
	b.Kyma(o.GetName())
	if labels := o.GetLabels(); len(labels) > 0 {
		b.BrokerPlan(labels["kyma-project.io/broker-plan-name"])
		b.GlobalAccount(labels["kyma-project.io/global-account-id"])
		b.SubAccount(labels["kyma-project.io/subaccount-id"])
		b.Region(labels["kyma-project.io/region"])
		b.Shoot(labels["kyma-project.io/shoot-name"])
	}
	return b
}

func (b *contextBuilderImpl) LoadFromScope(scope *cloudcontrolv1beta1.Scope) ContextBuilder {
	b.Provider(string(scope.Spec.Provider))
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

	var err error
	gvk := obj.GetObjectKind().GroupVersionKind()
	if gvk.Kind == "" {
		gvk, err = apiutil.GVKForObject(obj, scheme)
		if err != nil {
			return b
		}
	}

	kg := strings.ToLower(gvk.Kind)
	if gvk.Group != "" {
		kg = fmt.Sprintf("%s.%s", strings.ToLower(gvk.Kind), strings.ToLower(gvk.Group))
	}
	b.ObjKindGroup(kg)

	if kg == "customresourcedefinition.apiextensions.k8s.io" {
		if u, ok := obj.(*unstructured.Unstructured); ok {
			crdGroup, groupFound, groupErr := unstructured.NestedString(u.Object, "spec", "group")
			crdKind, kindFound, kindErr := unstructured.NestedString(u.Object, "spec", "names", "kind")
			if groupFound && kindFound && groupErr == nil && kindErr == nil {
				crdKg := fmt.Sprintf("%s.%s", strings.ToLower(crdKind), strings.ToLower(crdGroup))
				b.CrdKindGroup(crdKg)
			}
		}
		if crd, ok := obj.(*apiextensions.CustomResourceDefinition); ok {
			crdGroup := crd.Spec.Group
			crdKind := crd.Spec.Names.Kind
			crdKg := fmt.Sprintf("%s.%s", strings.ToLower(crdKind), strings.ToLower(crdGroup))
			b.CrdKindGroup(crdKg)
		}
	}

	if kg == "configmap" &&
		obj.GetLabels() != nil && obj.GetLabels()["busola.io/extension"] != "" {

		var general string
		if cm, ok := obj.(*unstructured.Unstructured); ok {
			gen, found, err := unstructured.NestedString(cm.Object, "data", "general")
			if found && err == nil {
				general = gen
			}
		}
		if cm, ok := obj.(*corev1.ConfigMap); ok {
			gen, found := cm.Data["general"]
			if found {
				general = gen
			}
		}

		if len(general) > 0 {
			obj := map[string]interface{}{}
			if err := yaml.Unmarshal([]byte(general), &obj); err == nil {
				cmGroup, groupFound, groupErr := unstructured.NestedString(obj, "resource", "group")
				cmKind, kindFound, kindErr := unstructured.NestedString(obj, "resource", "kind")
				if groupFound && kindFound && groupErr == nil && kindErr == nil {
					busolaKg := fmt.Sprintf("%s.%s", strings.ToLower(cmKind), strings.ToLower(cmGroup))
					b.BusolaKindGroup(busolaKg)
				}
			}
		}
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
