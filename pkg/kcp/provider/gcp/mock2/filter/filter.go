package filter

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"github.com/google/cel-go/interpreter"
	lru "github.com/hashicorp/golang-lru/v2"
	"google.golang.org/api/googleapi"
	"google.golang.org/protobuf/reflect/protoreflect"

	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

var _ = func() {

	projectId := "test-project"
	vpc := "test-vpc"
	scopeName := "test-scope"
	shootName := "test-shoot"
	subaccountId := "test-subaccount"

	// filtering addresses by network
	// pkg/kcp/provider/gcp/iprange/client/computeClient.go
	// gcpclient.GetNetworkFilter(projectId, vpc)
	const _ = "network=\"https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s\""
	var _ = gcpclient.GetNetworkFilter(projectId, vpc)

	// filtering addresses by network
	// ListGlobalAddresses
	// pkg/kcp/provider/gcp/iprange/client/oldComputeClient.go
	var _ = fmt.Sprintf("network=\"https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s\"", projectId, vpc)

	// FindRestoreOperation
	// pkg/kcp/provider/gcp/nfsrestore/client/fileRestoreClient.go
	var destFileFullPath = "projects/%s/locations/%s/instances/%s"
	var targetFilter = fmt.Sprintf("metadata.target=\"%s\"", destFileFullPath)
	var verbFilter = "metadata.verb=\"restore\""
	// filtering file operations
	var _ = fmt.Sprintf("%s AND %s", targetFilter, verbFilter)

	// client.GetSkrBackupsFilter
	var skrBackupsFilter = "labels.managed-by=\"%s\" AND labels.scope-name=\"%s\""
	var ManagedByValue = "cloud-manager"

	// loadNfsBackups
	// pkg/kcp/provider/gcp/nuke/loadNfsBackups.go
	var _ = fmt.Sprintf(skrBackupsFilter, ManagedByValue, scopeName)
	// filtering nfs backups
	var _ = gcpclient.GetSkrBackupsFilter(scopeName)

	// client.GetSharedBackupsFilter
	var sharedBackupsFilter = "labels.managed-by=\"%s\" AND " +
		"( labels.cm-allow-%s=\"" + util.GcpLabelBackupAccessibleFrom + "\"" +
		" OR labels.cm-allow-%s=\"" + util.GcpLabelBackupAccessibleFrom + "\"" +
		" OR labels.ALL=\"" + util.GcpLabelBackupAccessibleFrom + "\")"
	var _ = fmt.Sprintf(sharedBackupsFilter, ManagedByValue, shootName, subaccountId)

	var _ = gcpclient.GetSkrBackupsFilter(scopeName)

}

type FilterEngine[T any] struct {
	env                *cel.Env
	activationProvider func(obj T) (interpreter.Activation, error)
	//cache              sync.Map // map[string]cel.Program
	cache *lru.Cache[string, cel.Program]
}

func NewFilterEngine[T any]() (*FilterEngine[T], error) {
	var obj T
	var ap func(obj T) (interpreter.Activation, error)
	var opts []cel.EnvOption
	if msg, ok := (any)(obj).(protoreflect.ProtoMessage); ok {
		md := msg.ProtoReflect().Descriptor()
		opts = []cel.EnvOption{
			cel.Types(msg),
			ext.Strings(),
			cel.DeclareContextProto(md),
		}
		ap = func(obj T) (interpreter.Activation, error) {
			msg, ok := (any)(obj).(protoreflect.ProtoMessage)
			if !ok {
				return nil, fmt.Errorf("expected proto message, got %T", obj)
			}
			return cel.ContextProtoVars(msg)
		}
	} else {
		o, nameAliases, err := envOptionsFromStruct(obj)
		if err != nil {
			return nil, err
		}
		opts = append([]cel.EnvOption{ext.Strings()}, o...)
		ap = func(obj T) (interpreter.Activation, error) {
			return &anyActivation{
				obj:         obj,
				nameAliases: nameAliases,
			}, nil
		}
	}
	env, err := cel.NewEnv(opts...)
	if err != nil {
		return nil, err
	}

	cache, err := lru.New[string, cel.Program](128)
	if err != nil {
		return nil, err
	}
	return &FilterEngine[T]{
		env:                env,
		activationProvider: ap,
		cache:              cache,
	}, nil
}

func (f *FilterEngine[T]) Match(filter string, obj T) (bool, error) {
	if strings.TrimSpace(filter) == "" {
		return true, nil
	}

	mode, err := detectMode(filter)
	if err != nil {
		return false, err
	}

	celExpr, err := translate(filter, mode)
	if err != nil {
		return false, err
	}

	prog, err := f.compile(celExpr)
	if err != nil {
		return false, err
	}

	activation, err := f.activationProvider(obj)
	if err != nil {
		return false, err
	}
	out, _, err := prog.Eval(activation)
	if err != nil {
		return false, err
	}

	b, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("filter did not evaluate to bool")
	}

	return b, nil
}

func (f *FilterEngine[T]) compile(expr string) (cel.Program, error) {
	if p, ok := f.cache.Get(expr); ok {
		return p.(cel.Program), nil
	}

	ast, iss := f.env.Parse(expr)
	if iss != nil && iss.Err() != nil {
		return nil, iss.Err()
	}

	checked, iss := f.env.Check(ast)
	if iss != nil && iss.Err() != nil {
		return nil, iss.Err()
	}

	prog, err := f.env.Program(checked)
	if err != nil {
		return nil, err
	}

	_ = f.cache.Add(expr, prog)
	return prog, nil
}

//
// ---- Filter Mode Detection ----
//

type filterMode int

const (
	modeAIP160 filterMode = iota
	modeRegex
)

func detectMode(s string) (filterMode, error) {
	hasRegex := strings.Contains(s, " eq ") || strings.Contains(s, " ne ")
	hasAIP := strings.ContainsAny(s, "=<>:")

	if hasRegex && hasAIP {
		return 0, fmt.Errorf("cannot mix regex and AIP-160 operators")
	}

	if hasRegex {
		return modeRegex, nil
	}
	return modeAIP160, nil
}

//
// ---- Translation ----
//

func translate(s string, mode filterMode) (string, error) {
	switch mode {
	case modeAIP160:
		return translateAIP160(s), nil
	case modeRegex:
		return translateRegex(s)
	default:
		return "", fmt.Errorf("unknown mode")
	}
}

func translateAIP160(s string) string {
	s = strings.ReplaceAll(s, " AND ", " && ")
	s = strings.ReplaceAll(s, " OR ", " || ")
	s = strings.ReplaceAll(s, "=\"", " = \"")
	s = strings.ReplaceAll(s, " = ", " == ")
	s = strings.ReplaceAll(s, " != ", " != ")

	// labels.owner:*  →  has(labels["owner"])
	re := regexp.MustCompile(`labels\.([a-zA-Z0-9_-]+):\*`)
	s = re.ReplaceAllString(s, `has(labels["$1"])`)

	return s
}

func translateRegex(s string) (string, error) {
	re := regexp.MustCompile(`^\(?\s*([\w\.]+)\s+(eq|ne)\s+(.+?)\s*\)?$`)
	m := re.FindStringSubmatch(s)
	if m == nil {
		return "", fmt.Errorf("invalid regex filter")
	}

	field := m[1]
	op := m[2]
	pattern := strings.Trim(m[3], `"'`)

	expr := fmt.Sprintf(`%s.matches("%s")`, field, pattern)
	if op == "ne" {
		expr = "!" + expr
	}
	return expr, nil
}

//
// ---- Any Activation (global fields) ----
//

type anyActivation struct {
	obj         any
	nameAliases map[string]string
}

func (a *anyActivation) ResolveName(name string) (any, bool) {
	x := reflect.ValueOf(a.obj)
	if x.Kind() == reflect.Pointer {
		x = x.Elem()
	}
	if x.Kind() != reflect.Struct {
		return nil, false
	}
	n, ok := a.nameAliases[name]
	if ok {
		name = n
	}
	val := x.FieldByName(name)
	if !val.IsValid() {
		return nil, false
	}

	if val.Type().PkgPath() == "google.golang.org/api/googleapi" && val.Type().Name() == "RawMessage" {
		x, ok := val.Interface().(googleapi.RawMessage)
		if !ok {
			return nil, false
		}
		var m map[string]any
		if err := json.Unmarshal(x, &m); err != nil {
			return nil, false
		}
		return m, true
	}

	return val.Interface(), true
}

func (a *anyActivation) Parent() interpreter.Activation {
	return nil
}

func protoMapToGo(m protoreflect.Map, fd protoreflect.FieldDescriptor) any {
	result := make(map[string]any)

	m.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		var key string

		switch fd.MapKey().Kind() {
		case protoreflect.StringKind:
			key = k.String()
		default:
			key = fmt.Sprint(k.Interface())
		}

		result[key] = v.Interface()
		return true
	})

	return result
}

func protoListToGo(l protoreflect.List, fd protoreflect.FieldDescriptor) any {
	result := make([]any, l.Len())

	for i := 0; i < l.Len(); i++ {
		result[i] = l.Get(i).Interface()
	}

	return result
}

// envOptionsFromStruct generates CEL variable declarations
// for all exported fields of the provided struct value or type.
func envOptionsFromStruct(v any) ([]cel.EnvOption, map[string]string, error) {
	nameAliases := map[string]string{}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("expected struct, got %s", t.Kind())
	}

	var opts []cel.EnvOption

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// Skip unexported fields
		if f.PkgPath != "" {
			continue
		}

		name := f.Name
		tag := f.Tag.Get("json")
		if tag != "" {
			jsonName := strings.Split(tag, ",")[0]
			if jsonName == "-" {
				continue
			}
			if jsonName != "" {
				nameAliases[jsonName] = name
				name = jsonName
			}
		}
		celType := goTypeToCELType(f.Type)

		opts = append(opts, cel.Variable(name, celType))
	}

	return opts, nameAliases, nil
}

func goTypeToCELType(t reflect.Type) *cel.Type {
	if t.PkgPath() == "google.golang.org/api/googleapi" && t.Name() == "RawMessage" {
		return cel.MapType(cel.StringType, cel.DynType)
	}
	switch t.Kind() {

	case reflect.String:
		return cel.StringType

	case reflect.Bool:
		return cel.BoolType

	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return cel.IntType

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		return cel.UintType

	case reflect.Float32, reflect.Float64:
		return cel.DoubleType

	case reflect.Slice, reflect.Array:
		return cel.ListType(cel.DynType)

	case reflect.Map:
		return cel.MapType(cel.DynType, cel.DynType)

	case reflect.Struct:
		// Nested structs are accessed dynamically
		return cel.DynType

	case reflect.Pointer:
		return goTypeToCELType(t.Elem())

	default:
		return cel.DynType
	}
}
