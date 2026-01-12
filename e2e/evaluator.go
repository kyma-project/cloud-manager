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
	fixedData map[string]interface{}
}

func NewEvaluatorBuilder() *EvaluatorBuilder {
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
      if (cond.type == tp && cond.status == 'True') {
        return cond
      }
    }
  }
  return undefined
}

function findConditionFalse(obj, tp) {
  if (obj && obj.status && obj.status.conditions && obj.status.conditions.length) {
    for (var cond of obj.status.conditions) {
      if (cond.type == tp && cond.status == 'False') {
        return cond
      }
    }
  }
  return undefined
}

function findCondition(obj, tp) {
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
			vm:        vm,
			loaded:    map[string]struct{}{},
			evaluated: map[string]struct{}{},
		},
		fixedData: map[string]interface{}{},
	}
}

type evalData []*evalObj

type evalObj struct {
	ri        *ResourceInfo
	loaded    bool
	evaluated bool
	data      map[string]interface{}
	getter    func(ctx context.Context, alias string) (map[string]interface{}, error)
	mapper    func(alias string) (*meta.RESTMapping, error)
}

func (obj evalObj) toEvalMetadata() *evalObjMetadata {
	return &evalObjMetadata{
		ApiVersion: obj.ri.ApiVersion,
		Kind:       obj.ri.Kind,
		Name:       obj.ri.Name,
		Namespace:  obj.ri.Namespace,
		Evaluated:  obj.evaluated,
	}
}

type evalObjMetadata struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Evaluated  bool   `json:"evaluated"`
}

func (b *EvaluatorBuilder) Set(name string, value interface{}) {
	b.fixedData[name] = value
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
	for k, v := range b.fixedData {
		if err := b.evaluator.vm.GlobalObject().Set(k, v); err != nil {
			return fmt.Errorf("error setting fixed value %q to vm: %w", k, err)
		}
	}
	metadata := map[string]*evalObjMetadata{}
	for _, obj := range b.evalData {
		metadata[obj.ri.Alias] = obj.toEvalMetadata()
		if err := b.evaluator.vm.GlobalObject().Set(obj.ri.Alias, obj.data); err != nil {
			return fmt.Errorf("error setting resource %q to vm: %w", obj.ri.Alias, err)
		}
	}
	if err := b.evaluator.vm.GlobalObject().Set("_", metadata); err != nil {
		return fmt.Errorf("error setting metadata to vm: %w", err)
	}
	return nil
}

func (b *EvaluatorBuilder) evalDeclarations() ([]string, error) {
	var result []string
	for _, obj := range b.evalData {
		if obj.evaluated {
			continue
		}
		isEvaluated, err := b.evaluator.evalResource(obj.ri)
		if err != nil {
			return nil, err
		}
		if isEvaluated {
			obj.evaluated = true
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
		if !obj.evaluated {
			continue
		}

		data, err := obj.getter(ctx, obj.ri.Alias)
		if err != nil {
			return nil, err
		}

		obj.data = data

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

	b.evaluator.evaluated = map[string]struct{}{}
	b.evaluator.loaded = map[string]struct{}{}
	for _, obj := range b.evalData {
		if obj.evaluated {
			b.evaluator.evaluated[obj.ri.Alias] = struct{}{}
		}
		if obj.loaded {
			b.evaluator.loaded[obj.ri.Alias] = struct{}{}
		}
	}

	return b.evaluator, nil
}

// ======================================================

type Evaluator interface {
	Eval(txt string) (interface{}, error)
	EvalTruthy(txt string) (bool, error)
	EvalTemplate(txt string) (string, error)

	IsEvaluated(alias string) bool
	IsLoaded(alias string) bool
}

type defaultEvaluatorImpl struct {
	vm *goja.Runtime

	loaded    map[string]struct{}
	evaluated map[string]struct{}
}

func (e *defaultEvaluatorImpl) IsEvaluated(alias string) bool {
	_, ok := e.evaluated[alias]
	return ok
}

func (e *defaultEvaluatorImpl) IsLoaded(alias string) bool {
	_, ok := e.loaded[alias]
	return ok
}

func (e *defaultEvaluatorImpl) Eval(txt string) (interface{}, error) {
	v, err := e.vm.RunString(txt)
	if err != nil {
		return nil, err
	}
	if goja.IsUndefined(v) || goja.IsNull(v) || goja.IsNaN(v) || goja.IsInfinity(v) {
		return nil, nil
	}
	return v.Export(), nil
}

func (e *defaultEvaluatorImpl) EvalTruthy(txt string) (bool, error) {
	v, err := e.Eval(txt)
	// syntax error is a showstopper
	if IsSyntaxError(err) {
		return false, err
	}
	// any other error is falsy
	if err != nil {
		return false, nil
	}
	// nil is always falsy
	if v == nil {
		return false, nil
	}
	// recognize some common falsy values
	switch vv := v.(type) {
	case bool:
		if !vv {
			return false, nil
		}
	case int:
		if vv == 0 {
			return false, nil
		}
	case string:
		if len(vv) == 0 || strings.ToLower(vv) == "false" || vv == "0" {
			return false, nil
		}
	}
	// anything else is truthy
	return true, nil
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

func (e *defaultEvaluatorImpl) evalResource(ri *ResourceInfo) (bool, error) {
	if ri.Namespace != "" {
		namespace, err := e.EvalTemplate(ri.Namespace)
		if IgnoreReferenceAndTypeError(err) != nil {
			return false, fmt.Errorf("error evaluating %q namespace %q: %w", ri.Alias, ri.Namespace, err)
		}
		ri.Namespace = namespace
	}

	name, err := e.EvalTemplate(ri.Name)
	if IgnoreReferenceAndTypeError(err) != nil {
		return false, fmt.Errorf("error evaluating %q name %q: %w", ri.Alias, ri.Name, err)
	}

	if name != "" {
		ri.Name = name
		return true, nil
	}

	return false, nil
}

func GojaErrorName(err error) (string, bool) {
	var x *goja.Exception
	ok := errors.As(err, &x)
	if !ok {
		return "", false
	}

	obj, ok := x.Value().(*goja.Object)
	if !ok || obj == nil {
		return "", false
	}
	if obj.ClassName() != "Error" {
		return "", false
	}

	name := obj.Get("name")
	if name != nil {
		// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Error
		// one of RangeError, ReferenceError, TypeError, SyntaxError, URIError, AggregateError, InternalError
		return name.String(), true
	}
	return "", false
}

func IsEcmaError(err error) bool {
	_, ok := GojaErrorName(err)
	return ok
}

func IsReferenceError(err error) bool {
	name, ok := GojaErrorName(err)
	return ok && name == "ReferenceError"
}

func IsTypeError(err error) bool {
	name, ok := GojaErrorName(err)
	return ok && name == "TypeError"
}

func IsSyntaxError(err error) bool {
	name, ok := GojaErrorName(err)
	return ok && name == "SyntaxError"
}

func IgnoreReferenceAndTypeError(err error) error {
	if IsReferenceError(err) {
		return nil
	}
	if IsTypeError(err) {
		return nil
	}
	return err
}
