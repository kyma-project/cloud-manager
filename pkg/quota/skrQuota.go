package quota

import (
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"strings"
)

func InitConfig(cfg config.Config) {
	cfg.Path(
		"resourceQuota.skr",
		config.SourceFile("resourceQuotaSkr.yaml"),
		config.DefaultObj(DefaultSkrQuota()),
		config.Bind(SkrQuota),
	)
}

type SkrQuotaIntf interface {
	TotalCountForObj(obj runtime.Object, scheme *runtime.Scheme, skr string) int
	Override(obj runtime.Object, scheme *runtime.Scheme, skr string, value int)
}

func DefaultSkrQuota() SkrQuotaIntf {
	return &skrQuotaConfig{
		Defaults: map[string]int{
			// sigs.k8s.io/controller-runtime@v0.16.3/pkg/builder/controller.go#getControllerName
			// quota names are in the form `[lower(Kind)].[Group]/[quotaName]`
			"iprange.cloud-resources.kyma-project.io/totalCount":      1,
			"awsnfsvolume.cloud-resources.kyma-project.io/totalCount": 5,
		},
		Overrides: map[string]skrQuotaSpec{},
	}
}

type skrQuotaConfig struct {
	Defaults  skrQuotaSpec            `json:"defaults,omitempty" yaml:"defaults,omitempty"`
	Overrides map[string]skrQuotaSpec `json:"overrides,omitempty" yaml:"overrides,omitempty"`
}

type skrQuotaSpec map[string]int

var SkrQuota SkrQuotaIntf = DefaultSkrQuota()

func (q *skrQuotaConfig) TotalCountForObj(obj runtime.Object, scheme *runtime.Scheme, skr string) (result int) {
	result = util.MaxInt
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return
	}
	name := fmt.Sprintf("%s.%s/totalCount", strings.ToLower(gvk.Kind), gvk.Group)
	if q.Overrides != nil {
		skrSpec, ok := q.Overrides[skr]
		if ok {
			val, ok := skrSpec[name]
			if ok {
				result = val
				return
			}
		}
	}
	val, ok := q.Defaults[name]
	if ok {
		result = val
	}
	return
}

func (q *skrQuotaConfig) Override(obj runtime.Object, scheme *runtime.Scheme, skr string, value int) {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return
	}
	name := fmt.Sprintf("%s.%s/totalCount", strings.ToLower(gvk.Kind), gvk.Group)
	if len(skr) > 0 {
		if q.Overrides == nil {
			q.Overrides = map[string]skrQuotaSpec{}
		}
		spec, ok := q.Overrides[skr]
		if !ok {
			spec := skrQuotaSpec{}
			q.Overrides[skr] = spec
		}
		spec[name] = value
		return
	}

	q.Defaults[name] = value
}
