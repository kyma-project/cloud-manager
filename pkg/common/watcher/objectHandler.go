package watcher

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ObjectHandler interface {
	OnObject(ctx context.Context, gvk schema.GroupVersionKind, request reconcile.Request, obj client.Object) error
}

// LogWriter owns serialization, mutex, and the "---\n" separator for a write target.
type LogWriter interface {
	Write(data map[string]any) error
}

// FileLogWriter appends YAML documents to a file.
type FileLogWriter struct {
	filePath string
	mu       sync.Mutex
}

func NewFileLogWriter(filePath string) *FileLogWriter {
	return &FileLogWriter{filePath: filePath}
}

func (w *FileLogWriter) Write(data map[string]any) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	f, err := os.OpenFile(w.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err = f.WriteString("---\n"); err != nil {
		return err
	}

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err = enc.Encode(data); err != nil {
		return err
	}
	return enc.Close()
}

// StdoutLogWriter writes YAML documents to stdout.
type StdoutLogWriter struct {
	mu sync.Mutex
}

func NewStdoutLogWriter() *StdoutLogWriter {
	return &StdoutLogWriter{}
}

func (w *StdoutLogWriter) Write(data map[string]any) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := os.Stdout.WriteString("---\n"); err != nil {
		return err
	}

	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	if err := enc.Encode(data); err != nil {
		return err
	}
	return enc.Close()
}

// YamlLoggingObjectHandler shapes the data map and delegates writing to a LogWriter.
type YamlLoggingObjectHandler struct {
	writer LogWriter
}

func NewYamlLoggingObjectHandler(writer LogWriter) *YamlLoggingObjectHandler {
	return &YamlLoggingObjectHandler{writer: writer}
}

func (h *YamlLoggingObjectHandler) OnObject(_ context.Context, gvk schema.GroupVersionKind, request reconcile.Request, obj client.Object) error {
	data := map[string]any{
		"timestamp": time.Now().Format(time.RFC3339Nano),
	}
	if request.String() != "/" {
		data["request"] = request.String()
	}
	if gvk.GroupVersion().String() != "" || gvk.Kind != "" {
		gvkmap := map[string]any{}
		if gvk.GroupVersion().String() != "" {
			gvkmap["apiVersion"] = gvk.GroupVersion().String()
		}
		if gvk.Kind != "" {
			gvkmap["kind"] = gvk.Kind
		}
		data["gvk"] = gvkmap
	}
	if obj != nil {
		data["object"] = obj.(*unstructured.Unstructured).Object
	}
	return h.writer.Write(data)
}

// DiffObjectHandler forwards only changed fields to its inner handler.
type DiffObjectHandler struct {
	inner ObjectHandler
	state map[reconcile.Request]client.Object
}

func NewDiffObjectHandler(inner ObjectHandler) ObjectHandler {
	return &DiffObjectHandler{inner: inner, state: make(map[reconcile.Request]client.Object)}
}

func (d *DiffObjectHandler) OnObject(ctx context.Context, gvk schema.GroupVersionKind, request reconcile.Request, obj client.Object) error {
	if obj == nil {
		err := d.inner.OnObject(ctx, gvk, request, nil)
		delete(d.state, request)
		return err
	}
	if request.Name == "" || obj.GetName() == "" {
		return d.inner.OnObject(ctx, gvk, request, obj)
	}
	if obj.GetObjectKind().GroupVersionKind().GroupVersion().String() == "/" {
		return d.inner.OnObject(ctx, gvk, request, obj)
	}

	stripped, err := stripManagedFields(obj)
	if err != nil {
		return err
	}

	prior, exists := d.state[request]
	if !exists {
		d.state[request] = stripped
		return d.inner.OnObject(ctx, gvk, request, stripped)
	}

	priorJSON, err := json.Marshal(prior.(*unstructured.Unstructured).Object)
	if err != nil {
		return err
	}
	currentJSON, err := json.Marshal(stripped.(*unstructured.Unstructured).Object)
	if err != nil {
		return err
	}

	patchBytes, err := jsonpatch.CreateMergePatch(priorJSON, currentJSON)
	if err != nil {
		return err
	}

	var patchMap map[string]any
	if err = json.Unmarshal(patchBytes, &patchMap); err != nil {
		return err
	}
	patchObj := &unstructured.Unstructured{Object: patchMap}
	patchObj.SetAPIVersion(gvk.GroupVersion().String())
	patchObj.SetKind(gvk.Kind)
	patchObj.SetName(obj.GetName())
	patchObj.SetNamespace(obj.GetNamespace())

	d.state[request] = stripped
	return d.inner.OnObject(ctx, gvk, request, patchObj)
}

func stripManagedFields(obj client.Object) (client.Object, error) {
	raw, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err = json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if meta, ok := m["metadata"].(map[string]any); ok {
		delete(meta, "managedFields")
	}
	result := &unstructured.Unstructured{Object: m}
	return result, nil
}

func NewDefaultObjectHandler(filePath string) ObjectHandler {
	return &DiffObjectHandler{
		inner: &YamlLoggingObjectHandler{writer: NewFileLogWriter(filePath)},
		state: make(map[reconcile.Request]client.Object),
	}
}
