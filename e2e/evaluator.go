package e2e

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

type EvaluatorBuilder struct {
	evaluator *defaultEvaluatorImpl
	evalData  evalData
}

func NewEvaluatorBuilder(skrNamespace string) *EvaluatorBuilder {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	util.MustVoid(vm.GlobalObject().Set("id", func(fc goja.FunctionCall, r *goja.Runtime) goja.Value {
		length := 7
		if len(fc.Arguments) > 0 {
			length = int(fc.Arguments[0].ToInteger())
		}
		id := uuid.New()
		result := strings.ReplaceAll(id.String(), "-", "")
		f := fmt.Sprintf("%%.%ds", length)
		result = fmt.Sprintf(f, result)
		return r.ToValue(result)
	}))
	_ = util.Must(vm.RunString(`
function findConditionTrue(obj, tp) {
  if (obj && obj.status && obj.status.conditions && obj.status.conditions.length) {
    for (var cond of obj.status.conditions) {
      if (cond.type == tp) {
        return cond
      }
    }
  }
  return undefined
}
`))

	return &EvaluatorBuilder{
		evaluator: &defaultEvaluatorImpl{
			vm:           vm,
			skrNamespace: skrNamespace,
		},
	}
}

type evalData []*evalObj

type evalObj struct {
	ri     *ResourceInfo
	data   map[string]interface{}
	getter func(ctx context.Context, alias string) (map[string]interface{}, error)
	mapper func(alias string) (*meta.RESTMapping, error)
	loaded bool
}

func (obj evalObj) toEvalMetadata() *evalObjMetadata {
	return &evalObjMetadata{
		ApiVersion: obj.ri.ApiVersion,
		Kind:       obj.ri.Kind,
		Name:       obj.ri.Name,
		Namespace:  obj.ri.Namespace,
		Evaluated:  obj.ri.Evaluated,
	}
}

type evalObjMetadata struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Evaluated  bool   `json:"evaluated"`
}

func (b *EvaluatorBuilder) Add(c ClusterEvaluationHandle) *EvaluatorBuilder {
	for _, ri := range c.AllResources() {
		b.evalData = append(b.evalData, &evalObj{
			ri:     ri,
			data:   nil,
			getter: c.Get,
			mapper: c.RestMapping,
		})
	}
	return b
}

func (b *EvaluatorBuilder) setToVm() error {
	metadata := map[string]*evalObjMetadata{}
	for _, obj := range b.evalData {
		metadata[obj.ri.Alias] = obj.toEvalMetadata()
		if err := b.evaluator.vm.GlobalObject().Set(obj.ri.Alias, obj.data); err != nil {
			return err
		}
	}
	if err := b.evaluator.vm.GlobalObject().Set("_", metadata); err != nil {
		return err
	}
	return nil
}

func (b *EvaluatorBuilder) evalDeclarations() ([]string, error) {
	var result []string
	for _, obj := range b.evalData {
		if obj.ri.Evaluated {
			continue
		}
		if err := b.evaluator.evalResource(obj.ri); err != nil {
			return nil, err
		}
		if obj.ri.Evaluated {
			result = append(result, obj.ri.Name)
		}
	}
	return result, nil
}

func (b *EvaluatorBuilder) load(ctx context.Context) ([]string, error) {
	var result []string
	for _, obj := range b.evalData {
		if obj.loaded {
			continue
		}
		if !obj.ri.Evaluated {
			continue
		}

		data, err := obj.getter(ctx, obj.ri.Alias)
		if err != nil {
			return nil, err
		}

		obj.data = data

		// TODO: mapper is not call so far, maybe it won't be needed at all

		obj.loaded = true
		result = append(result, obj.ri.Alias)
	}
	return result, nil
}

func (b *EvaluatorBuilder) Build(ctx context.Context) (Evaluator, error) {
	if err := b.setToVm(); err != nil {
		return nil, err
	}

	_, err := b.evalDeclarations()
	if err != nil {
		return nil, err
	}

	var loadedResources []string
	loadedResources, err = b.load(ctx)
	if err != nil {
		return nil, err
	}

	if len(loadedResources) > 0 {
		return b.Build(ctx)
	}

	return b.evaluator, nil
}

// ======================================================

type Evaluator interface {
	Eval(txt string) (interface{}, error)
	EvalTemplate(txt string) (string, error)
}

type defaultEvaluatorImpl struct {
	vm           *goja.Runtime
	skrNamespace string
}

func (e *defaultEvaluatorImpl) Eval(txt string) (interface{}, error) {
	v, err := e.vm.RunString(txt)
	if err != nil {
		// a bit hacky way to find out if it's a ReferenceError
		// that happens if expression has some undefined variable
		// like in the case when object is not evaluated or loaded
		var x *goja.Exception
		ok := errors.As(err, &x)
		if ok {
			obj, ok := x.Value().(*goja.Object)
			if ok {
				if obj.ClassName() == "Error" {
					return nil, nil
				}
				// to get the error name you can do
				// obj.Get("name")
				// it will return string Value of "ReferenceError", "TypeError"...
			}
		}
		return "", fmt.Errorf("error evaluating template %q: %v", txt, err)
	}
	if goja.IsUndefined(v) || goja.IsNull(v) || goja.IsNaN(v) || goja.IsInfinity(v) {
		return nil, nil
	}
	return v.Export(), nil
}

func (e *defaultEvaluatorImpl) EvalTemplate(txt string) (string, error) {
	v, err := e.Eval(fmt.Sprintf("`%s`", txt))
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", nil
	}
	return fmt.Sprintf("%s", v), nil
}

func (e *defaultEvaluatorImpl) evalResource(ri *ResourceInfo) error {
	if ri.Namespace != "" {
		namespace, err := e.EvalTemplate(ri.Namespace)
		if err != nil {
			return fmt.Errorf("error evaluating %q namespace %q: %w", ri.Alias, ri.Namespace, err)
		}
		ri.Namespace = namespace
	} else if e.skrNamespace != "" {
		ri.Namespace = e.skrNamespace
	} else {
		ri.Namespace = "default"
	}

	name, err := e.EvalTemplate(ri.Name)
	if err != nil {
		return fmt.Errorf("error evaluating %q name %q: %w", ri.Alias, ri.Name, err)
	}

	if name != "" {
		ri.Name = name
		ri.Evaluated = true
	}

	return nil
}
