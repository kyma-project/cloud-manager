package e2e

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
)

func debugLog(ctx context.Context, onOff string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	v, err := strconv.ParseBool(onOff)
	if err != nil {
		return ctx, err
	}

	session.DebugLog(v)

	return ctx, nil
}

func debugWait(ctx context.Context, suffix string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	name := fmt.Sprintf("e2e-lock-%s-%s", suffix, util.RandomString(6))
	alias := strings.ReplaceAll(name, "-", "_")

	err := session.CurrentCluster().AddResources(ctx, &ResourceDeclaration{
		Alias:      alias,
		Kind:       "ConfigMap",
		ApiVersion: "v1",
		Name:       name,
		Namespace:  world.Config().SkrNamespace,
	})
	if err != nil {
		return ctx, err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: world.Config().SkrNamespace,
			Name:      name,
			Annotations: map[string]string{
				e2elib.AliasLabel:             alias,
				e2elib.ScenarioNameAnnotation: session.GetScenarioName(),
				e2elib.StepNameAnnotation:     session.GetStepName(),
			},
		},
	}
	err = session.CurrentCluster().GetClient().Create(ctx, cm)
	if err != nil {
		return ctx, fmt.Errorf("error creating debug wait configmap: %w", err)
	}

	err = session.EventuallyResourceDoesNotExist(ctx, alias)

	return ctx, err
}

func errEvalContextBuilding(err error) error {
	return fmt.Errorf("error building evaluation context: %w", err)
}

func eventuallyTimeoutIs(ctx context.Context, d time.Duration) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}
	session.Timing().EventuallyTimeout = d
	return ctx, nil
}

func thereIsSharedSKRWithProvider(ctx context.Context, provider string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}
	pt, err := cloudcontrolv1beta1.ParseProviderType(provider)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse provider type: %w", err)
	}
	alias := fmt.Sprintf("shared-%s", pt)
	_, err = session.AddExistingCluster(ctx, alias)
	return ctx, err
}

func moduleIsAdded(ctx context.Context, moduleName string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	err := wait.PollUntilContextTimeout(ctx, time.Millisecond, 10*time.Millisecond, true, func(ctx context.Context) (done bool, err error) {
		kyma, err := session.CurrentCluster().GetSkrKyma(ctx)
		if err != nil {
			return false, err
		}

		isFound := false
		for _, m := range kyma.Spec.Modules {
			if m.Name == moduleName {
				isFound = true
				break
			}
		}

		if isFound {
			return false, fmt.Errorf("module %q is already added", moduleName)
		}

		kyma.Spec.Modules = append(kyma.Spec.Modules, operatorv1beta2.Module{
			Name: moduleName,
		})

		err = session.CurrentCluster().GetClient().Update(ctx, kyma)
		if apierrors.IsConflict(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func moduleIsRemoved(ctx context.Context, moduleName string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	err := wait.PollUntilContextTimeout(ctx, time.Millisecond, 10*time.Millisecond, true, func(ctx context.Context) (done bool, err error) {
		kyma, err := session.CurrentCluster().GetSkrKyma(ctx)
		if err != nil {
			return false, err
		}

		isFound := false
		for _, m := range kyma.Spec.Modules {
			if m.Name == moduleName {
				isFound = true
				break
			}
		}

		if !isFound {
			return false, fmt.Errorf("module %q not active", moduleName)
		}

		kyma.Spec.Modules = pie.FilterNot(kyma.Spec.Modules, func(m operatorv1beta2.Module) bool {
			return m.Name == moduleName
		})

		err = session.CurrentCluster().GetClient().Update(ctx, kyma)
		if apierrors.IsConflict(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func moduleIsActive(ctx context.Context, moduleName string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	kyma, err := session.CurrentCluster().GetSkrKyma(ctx)
	if err != nil {
		return ctx, err
	}

	isFound := false
	for _, m := range kyma.Spec.Modules {
		if m.Name == moduleName {
			isFound = true
			break
		}
	}

	if isFound {
		return ctx, nil
	}

	return ctx, fmt.Errorf("module %q is expected to be active, but it is not active", moduleName)
}

func moduleIsNotActive(ctx context.Context, moduleName string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	kyma, err := session.CurrentCluster().GetSkrKyma(ctx)
	if err != nil {
		return ctx, err
	}

	isFound := false
	for _, m := range kyma.Spec.Modules {
		if m.Name == moduleName {
			isFound = true
			break
		}
	}

	if !isFound {
		return ctx, nil
	}

	return ctx, fmt.Errorf("module %q is expected not to be active, but it is active", moduleName)
}

/*
Given resource declaration:

	| Alias | Kind                      | ApiVersion              | Name                                 | Namespace  |
	| crd   | customresourcedefinitions | apiextensions.k8s.io/v1 | destinationrules.networking.istio.io | $namespace |
	| cm    | ConfigMap                 | v1                      | `test-${id(4)}`                      | `namespace` |
*/
func resourceDeclaration(ctx context.Context, tbl *godog.Table) (context.Context, error) {
	rd, err := ParseResourceDeclarations(tbl)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse resource declaration: %w", err)
	}

	if GetCurrentScenarioSession(ctx).CurrentCluster() == nil {
		return ctx, fmt.Errorf("current cluster is not defined")
	}

	err = GetCurrentScenarioSession(ctx).CurrentCluster().AddResources(ctx, rd...)
	if err != nil {
		return ctx, fmt.Errorf("error adding resource declaration: %w", err)
	}

	return ctx, nil
}

func resourceIsCreated(ctx context.Context, alias string, doc *godog.DocString) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	ri := session.CurrentCluster().GetResource(alias)

	eval, err := session.Eval(ctx)
	if err != nil {
		return ctx, errEvalContextBuilding(err)
	}
	if !eval.IsEvaluated(alias) {
		return ctx, fmt.Errorf("resource with alias %q is not evaluated", alias)
	}

	txt, err := eval.EvalTemplate(doc.Content)
	if err != nil {
		return ctx, err
	}
	if txt == "" {
		return ctx, fmt.Errorf("resource manifest with alias %q is empty", alias)
	}
	arr, err := util.YamlMultiDecodeToUnstructured([]byte(txt))
	if err != nil {
		return ctx, fmt.Errorf("failed to parse resource yaml: %w", err)
	}
	if len(arr) != 1 {
		return ctx, fmt.Errorf("expected one resource in yaml but got %d", len(arr))
	}
	obj := arr[0]

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[e2elib.AliasLabel] = alias
	annotations[e2elib.ScenarioNameAnnotation] = session.GetScenarioName()
	annotations[e2elib.StepNameAnnotation] = session.GetStepName()
	obj.SetAnnotations(annotations)

	if obj.GetNamespace() == "" {
		obj.SetNamespace(ri.Namespace)
	}
	if obj.GetName() == "" {
		obj.SetName(ri.Name)
	}
	if obj.GetKind() == "" {
		obj.SetKind(ri.Kind)
	}
	if obj.GetAPIVersion() == "" {
		obj.SetAPIVersion(ri.ApiVersion)
	}

	err = session.CurrentCluster().GetClient().Create(ctx, obj)

	if err == nil {
		session.CurrentCluster().DeleteOnTerminate(obj)
	}

	return ctx, err
}

func resourceIsDeleted(ctx context.Context, alias string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	data, err := session.CurrentCluster().Get(ctx, alias)
	if err != nil {
		return ctx, err
	}

	u := &unstructured.Unstructured{Object: data}
	err = session.CurrentCluster().GetClient().Delete(ctx, u)

	return ctx, err
}

/*
Then eventually "cm.data.foo == 'bar'" is ok, unless:

	| cm.data.foo && cm.data.foo != 'bar' |
*/
func eventuallyValueIsOkUnless(ctx context.Context, expression string, unless *godog.Table) (context.Context, error) {
	arrUnless := pie.Map(unless.Rows, func(row *messages.PickleTableRow) string {
		return row.Cells[0].Value
	})

	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	err := session.EventuallyValueIsOK(ctx, expression, arrUnless...)

	return ctx, err
}

func eventuallyValueIsOk(ctx context.Context, expression string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	err := session.EventuallyValueIsOK(ctx, expression)

	return ctx, err
}

func valueIsOk(ctx context.Context, expression string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}
	eval, err := session.Eval(ctx)
	if err != nil {
		return ctx, errEvalContextBuilding(err)
	}
	ok, err := eval.EvalTruthy(expression)
	if err != nil {
		return ctx, err
	}
	if !ok {
		return ctx, fmt.Errorf("expected expression %s to be truthy", expression)
	}
	return ctx, nil
}

/*
PVC x file operations succeed:

		| Operation            | Path    | Content      |
		| Create               | foo.txt | some content |
		| Append               | foo.txt | some more    |
		| Delete               | foo.txt |              |
		| Contains             | foo.txt | content      |
		| Exists               | foo.txt |              |
	    | SleepOnPodStart      |         | 8888         |
	    | SleepBeforePodDelete |         | 1m           |
*/
func pvcFileOperationsSucceed(ctx context.Context, alias string, ops *godog.Table) (context.Context, error) {
	arr, err := ad.ParseSlice(ops)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse operations, the table must have first header row with colums Operation, Path, Content: %w", err)
	}
	var sleepBeforePodDelete time.Duration
	var fileOps []FileOperationFunc
	for i, row := range arr {
		opType, ok := row["Operation"]
		if !ok {
			return ctx, fmt.Errorf("missing 'Operation' column in row %d", i+1)
		}
		path, ok := row["Path"]
		if !ok {
			return ctx, fmt.Errorf("missing 'Path' column in row %d", i+1)
		}
		content := row["Content"]
		switch opType {
		case "Create":
			fileOps = append(fileOps, CreateFileOperation(path, content))
		case "Append":
			fileOps = append(fileOps, AppendFileOperation(path, content))
		case "Delete":
			fileOps = append(fileOps, DeleteFileOperation(path))
		case "Contains":
			fileOps = append(fileOps, FileContainsOperation(path, content))
		case "Exists":
			fileOps = append(fileOps, FileExistsOperation(path))
		case "SleepOnPodStart":
			fileOps = pie.Insert(fileOps, 0, func(_ string) []string {
				return []string{fmt.Sprintf("sleep %s", content)}
			})
		case "SleepBeforePodDelete":
			d, err := time.ParseDuration(content)
			if err != nil {
				return ctx, fmt.Errorf("failed to parse 'SleepBeforePodDelete' value: %w", err)
			}
			sleepBeforePodDelete = d
		default:
			return ctx, fmt.Errorf("unknown operation '%s' in row %d, valid operations are: Create, Append, Delete, Contains, Exists", opType, i+1)
		}
	}

	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	// create the evaluation context so the resource declaration evaluation and loaded state are fresh
	eval, err := session.Eval(ctx)
	if err != nil {
		return ctx, errEvalContextBuilding(err)
	}

	ri := session.CurrentCluster().GetResource(alias)
	if ri == nil {
		return ctx, fmt.Errorf("resource with alias %s not declared", alias)
	}
	if ri.GVK.Kind != "PersistentVolumeClaim" {
		return ctx, fmt.Errorf("resource with alias %s is %s but a PersistentVolumeClaim is required", alias, ri.GVK.Kind)
	}
	if !eval.IsEvaluated(alias) {
		return ctx, fmt.Errorf("resource with alias %s is not yet evaluated", alias)
	}
	if !eval.IsLoaded(alias) {
		return ctx, fmt.Errorf("pvc resource with alias %s is not yet loaded", alias)
	}

	allDone := "All done!"

	rootDir := "/mnt/" + ri.Name
	// name must be valid k8s uri name
	// alias must be valid js variable name
	// to simplify let's use same value for both but respect the constraints
	name := "e2epvcop" + util.RandomString(6)
	podAlias := name
	fileOps = append(fileOps, EchoOperation(allDone))
	scriptLines := CombineFileOperations(fileOps...)(rootDir)
	b := NewPodBuilder(name, ri.Namespace, "ubuntu").
		WithPodDetails(
			PodWithScript(scriptLines),
			PodWithMountFromPVC(ri.Name, "", ""),
		).
		WithAnnotation(e2elib.AliasLabel, alias).
		WithAnnotation(e2elib.ScenarioNameAnnotation, session.GetScenarioName()).
		WithAnnotation(e2elib.StepNameAnnotation, session.GetStepName())

	err = session.CurrentCluster().AddResources(ctx, &ResourceDeclaration{
		Alias:      podAlias,
		Kind:       "Pod",
		ApiVersion: "v1",
		Name:       name,
		Namespace:  ri.Namespace,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to declare pvc pod resource: %w", err)
	}

	session.Logger(ctx).
		WithValues("yaml", b.DumpYamlText(session.CurrentCluster().GetScheme())).
		Info("pvcFileOp creating resources")

	err = b.Create(ctx, session.CurrentCluster())

	dumpYamlText := b.DumpYamlText(session.CurrentCluster().GetScheme())

	if err != nil {
		return ctx, fmt.Errorf("error creating pvc operation resources:\n%w\n\n---\n%s", err, dumpYamlText)
	}

	failed := false

	err = session.EventuallyValueIsOK(
		ctx,
		fmt.Sprintf(`%s.status.phase == "Succeeded"`, podAlias),
		fmt.Sprintf(`%s.status.phase == "Failed"`, podAlias),
	)
	if err != nil {
		failed = true
	}

	logs, err := session.CurrentCluster().PodLogs(ctx, ri.Namespace, name, name)
	if err != nil {
		return ctx, fmt.Errorf("failed to get pvc operation pod logs: %w", err)
	}

	session.Logger(ctx).
		WithValues("log", logs).
		Info("pvcFileOp log")

	if sleepBeforePodDelete > 0 {
		time.Sleep(sleepBeforePodDelete)
	}

	err = b.Delete(ctx, session.CurrentCluster().GetClient())
	if err != nil {
		return ctx, fmt.Errorf("error deleting pvc operation resources: %w", err)
	}

	if failed {
		return ctx, fmt.Errorf("pvc operation failed:\n%s\n\n---\n%s", logs, dumpYamlText)
	}

	if strings.Contains(logs, allDone) {
		return ctx, nil
	}

	return ctx, fmt.Errorf("pvc operation did not succeeded:\n%s\n\n---\n%s", logs, dumpYamlText)
}

func eventuallyResourceDoesNotExist(ctx context.Context, alias string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	err := session.EventuallyResourceDoesNotExist(ctx, alias)

	return ctx, err
}

func resourceDoesNotExist(ctx context.Context, alias string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}
	eval, err := session.Eval(ctx)
	if err != nil {
		return ctx, errEvalContextBuilding(err)
	}

	v, err := eval.Eval(alias)
	if err != nil {
		return ctx, err
	}
	if v != nil {
		return ctx, fmt.Errorf("expected resource %s to not exist, but it does", alias)
	}
	return ctx, nil
}

func logsOfContainerInPodContain(ctx context.Context, containerName string, alias string, content string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}
	eval, err := session.Eval(ctx)
	if err != nil {
		return ctx, errEvalContextBuilding(err)
	}

	ri := session.CurrentCluster().GetResource(alias)
	if ri == nil {
		return ctx, fmt.Errorf("resource with alias %s not declared", alias)
	}
	if ri.GVK.Kind != "Pod" {
		return ctx, fmt.Errorf("resource with alias %s has kind %s but a Pod is required", alias, ri.GVK.Kind)
	}
	if !eval.IsEvaluated(alias) {
		return ctx, fmt.Errorf("resource with alias %s is not yet evaluated", alias)
	}
	if !eval.IsLoaded(alias) {
		return ctx, fmt.Errorf("resource with alias %s is not loaded/does not exist", alias)
	}

	logs, err := session.CurrentCluster().PodLogs(ctx, ri.Namespace, ri.Name, containerName)
	if err != nil {
		return ctx, fmt.Errorf("failed to get logs of container %s in pod %s: %w", containerName, alias, err)
	}

	if !strings.Contains(logs, content) {
		return ctx, fmt.Errorf("logs of container %s in pod %s do not contain expected content:\n%s", containerName, alias, logs)
	}

	return ctx, nil
}

/*
Then HTTP operation succeedes:

	| Url            | https://example.com |
	| Method         | POST                |
	| ContentType    | application/json    |
	| Data           | {"foo": "bar"}      |
	| MaxTime        | 10                  |
	| ExpectedOutput | something           |
*/
func httpOperationSucceeds(ctx context.Context, tbl *godog.Table) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	x, err := ad.CreateInstance(new(HttpOperation), tbl)
	if err != nil {
		return ctx, fmt.Errorf("invalid HTTP operation parameters: %w", err)
	}
	op := x.(*HttpOperation)
	if err := op.Validate(); err != nil {
		return ctx, err
	}

	name := "e2ehttpop" + util.RandomString(6)
	b := NewPodBuilder(name, world.Config().SkrNamespace, "curlimages/curl").
		WithPodDetails(
			PodWithArguments(op.Args()...),
		).
		WithAnnotation(e2elib.ScenarioNameAnnotation, session.GetScenarioName()).
		WithAnnotation(e2elib.StepNameAnnotation, session.GetStepName())

	if err := session.CurrentCluster().AddResources(ctx, &ResourceDeclaration{
		Alias:      name,
		Kind:       "Pod",
		ApiVersion: "v1",
		Name:       name,
		Namespace:  world.Config().SkrNamespace,
	}); err != nil {
		return ctx, fmt.Errorf("failed to declare http operation resources: %w", err)
	}
	if err := b.Create(ctx, session.CurrentCluster()); err != nil {
		return ctx, fmt.Errorf("failed to create http operation resources: %w", err)
	}

	failed := false
	err = session.EventuallyValueIsOK(
		ctx,
		fmt.Sprintf(`%s.status.phase == "Succeeded"`, name),
		fmt.Sprintf(`%s.status.phase == "Failed"`, name),
	)
	if err != nil {
		failed = true
	}

	logs, err := session.CurrentCluster().PodLogs(ctx, world.Config().SkrNamespace, name, name)
	if err != nil {
		return ctx, err
	}

	if err := b.Delete(ctx, session.CurrentCluster().GetClient()); err != nil {
		return ctx, fmt.Errorf("error deleting http operation resources: %w", err)
	}

	if failed {
		return ctx, fmt.Errorf("http operation failed:\n%s", logs)
	}

	if !strings.Contains(logs, op.ExpectedOutput) {
		return ctx, fmt.Errorf("http response do not match expected output %q:\n%s", op.ExpectedOutput, logs)
	}

	return ctx, nil
}

/*
Then Redis "PING" gives "PONG" with:

		| Host        | Secret | ${redis.metadata.name} | host       |
		| Port        | Secret | ${redis.metadata.name} | port       |
		| Auth        | Secret | ${redis.metadata.name} | authString |
		| TLS         | true   |                        |            |
		| CA          | Secret | ${redis.metadata.name} | CaCert.pem |
		| Version     | 7.4    |                        |            |
	    | ClusterMode | true   |                        |            |
*/
func redisGivesWith(ctx context.Context, cmd string, out string, tbl *godog.Table) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	eval, err := session.Eval(ctx)
	if err != nil {
		return ctx, errEvalContextBuilding(err)
	}

	for _, row := range tbl.Rows {
		for _, cell := range row.Cells {
			v, err := eval.EvalTemplate(cell.Value)
			if err != nil {
				return ctx, fmt.Errorf("failed to evaluate template %q: %w", cell.Value, err)
			}
			cell.Value = v
		}
	}

	name := "e2eredisop" + util.RandomString(6)

	b := NewPodBuilder(name, world.Config().SkrNamespace, "redis")

	makeEnv := func(envVarName string, row *messages.PickleTableRow) error {
		switch row.Cells[1].Value {
		case "Secret":
			b.WithPodDetails(PodWithEnvFromSecret(envVarName, row.Cells[2].Value, row.Cells[3].Value))
		case "ConfigMap":
			b.WithPodDetails(PodWithEnvFromConfigMap(envVarName, row.Cells[2].Value, row.Cells[3].Value))
		case "Fixed":
			b.WithPodDetails(PodWithFixedEnvVar(envVarName, row.Cells[2].Value))
		default:
			return fmt.Errorf("invalid value indicator %q", row.Cells[2].Value)
		}
		return nil
	}
	makeVol := func(volName string, row *messages.PickleTableRow) error {
		switch row.Cells[1].Value {
		case "Secret":
			b.WithPodDetails(PodWithMountFromSecret(row.Cells[2].Value, volName, ""))
		case "ConfigMap":
			b.WithPodDetails(PodWithMountFromConfigMap(row.Cells[2].Value, volName, "", nil))
		default:
			return fmt.Errorf("invalid value indicator %q", row.Cells[1].Value)
		}
		return nil
	}

	opts := RedisOptions{}

	for _, row := range tbl.Rows {
		switch row.Cells[0].Value {
		case "Host":
			opts.Host = true
			if err := makeEnv("HOST", row); err != nil {
				return ctx, err
			}
		case "Port":
			opts.Port = true
			if err := makeEnv("PORT", row); err != nil {
				return ctx, err
			}
		case "Auth":
			opts.Auth = true
			if err := makeEnv("REDISCLI_AUTH", row); err != nil {
				return ctx, err
			}
		case "TLS":
			b, err := strconv.ParseBool(row.Cells[1].Value)
			if err != nil {
				return ctx, fmt.Errorf("invalid TLS value, expected true/false: %w", err)
			}
			opts.TLS = b
		case "CA":
			if err := makeVol("cacert", row); err != nil {
				return ctx, err
			}
			b.WithPodDetails(PodWithFixedEnvVar("CA", row.Cells[3].Value))
			opts.CA = true
		case "Version":
			opts.Version = row.Cells[1].Value
		case "ClusterMode":
			b, err := strconv.ParseBool(row.Cells[1].Value)
			if err != nil {
				return ctx, fmt.Errorf("invalid ClusterMode value, expected true/false: %w", err)
			}
			opts.ClusterMode = b
		default:
			return ctx, fmt.Errorf("invalid value indicator %q", row.Cells[0].Value)
		} // switch row[0]
	} // for rows

	if !opts.Host {
		return ctx, fmt.Errorf("mandatory Redis Host param is not set")
	}

	var scriptLines []string
	//scriptLines = append(scriptLines, "sleep 88888")
	if opts.TLS && !opts.CA {
		scriptLines = append(
			scriptLines,
			"apt-get update",
			"apt-get install -y ca-certificates",
			"update-ca-certificates",
		)
	}
	command := "redis-cli -h $HOST"
	if opts.Port {
		command += " -p $PORT"
	}
	if opts.TLS {
		command += " --tls"
	}
	if opts.ClusterMode {
		command += " -c"
	}
	if opts.CA {
		command += " --cacert /mnt/cacert/$CA"
	}

	command += " " + cmd

	scriptLines = append(scriptLines, command)

	if opts.Version == "" {
		opts.Version = "latest"
	}

	b.
		WithPodDetails(
			PodWithScript(scriptLines),
			PodWithImage("redis:"+opts.Version),
		).
		WithAnnotation(e2elib.ScenarioNameAnnotation, session.GetScenarioName()).
		WithAnnotation(e2elib.StepNameAnnotation, session.GetStepName())

	if err := session.CurrentCluster().AddResources(ctx, &ResourceDeclaration{
		Alias:      name,
		Kind:       "Pod",
		ApiVersion: "v1",
		Name:       name,
		Namespace:  world.Config().SkrNamespace,
	}); err != nil {
		return ctx, fmt.Errorf("failed to declare redis operation pod: %w", err)
	}

	session.Logger(ctx).
		WithValues("yaml", b.DumpYamlText(session.CurrentCluster().GetScheme())).
		Info("redis op creating resources")

	err = b.Create(ctx, session.CurrentCluster())

	dumpYamlText := b.DumpYamlText(session.CurrentCluster().GetScheme())

	if err != nil {
		return ctx, fmt.Errorf("failed to create redis operation pod: \n%w\n\n---\n%s", err, dumpYamlText)
	}

	failed := false
	err = session.EventuallyValueIsOK(
		ctx,
		fmt.Sprintf(`%s.status.phase == "Succeeded"`, name),
		fmt.Sprintf(`%s.status.phase == "Failed"`, name),
	)
	if err != nil {
		failed = true
	}

	logs, err := session.CurrentCluster().PodLogs(ctx, b.Pod().Namespace, name, name)
	if err != nil {
		return ctx, fmt.Errorf("failed to retrieve redis operation pod logs: %w", err)
	}

	session.Logger(ctx).
		WithValues("log", logs).
		Info("redis op logs")

	if err := b.Delete(ctx, session.CurrentCluster().GetClient()); err != nil {
		return ctx, fmt.Errorf("failed to delete redis operation pod: %w", err)
	}

	if failed {
		return ctx, fmt.Errorf("redis operation failed:\n%s\n\n---\n%s", logs, dumpYamlText)
	}

	if !strings.Contains(logs, out) {
		return ctx, fmt.Errorf("redis operation did not return expected %q:\n%s\n\n---\n%s", out, logs, dumpYamlText)
	}

	return ctx, nil
}

/*
Given tf module "peeringTarget" is applied:

	| source             | terraform-aws-modules/vpc/aws ~> 6.5.1  |
	| provider           | hashicorp/aws ~> 6.0                    |
	| provider           | random 3.7.2                            |
	| input_var_string   | "some value"             |
	| input_var_int      | 123                      |
*/
func tfModuleIsApplied(ctx context.Context, alias string, tbl *godog.Table) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	data, err := ad.ParseMap(tbl)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse table: %w", err)
	}

	eval, err := session.Eval(ctx)
	if err != nil {
		return ctx, errEvalContextBuilding(err)
	}

	b := world.Cloud().WorkspaceBuilder(alias)
	for k, v := range data {
		switch k {
		case "source":
			vv := v
			if strings.HasPrefix(vv, "./") || strings.HasPrefix(vv, "../") {
				vv = path.Join(world.Config().ConfigDir, "e2e/tf", v)
			}
			b.WithSource(vv)
		case "provider":
			b.WithProvider(v)
		default:
			vv, err := eval.EvalTemplate(v)
			if err != nil {
				return ctx, fmt.Errorf("failed to evaluate tf variable %q: %w", v, err)
			}
			b.WithVariable(k, vv)
		}
	}
	if err := b.Validate(); err != nil {
		return ctx, fmt.Errorf("invalid tf module configuration: %w", err)
	}

	ws := b.Build()
	if err := session.AddTfWorkspace(ws); err != nil {
		return ctx, fmt.Errorf("failed to add tf workspace to session: %w", err)
	}

	if err := ws.Create(); err != nil {
		return ctx, fmt.Errorf("failed to create tf workspace: %w", err)
	}
	if err := ws.Init(); err != nil {
		return ctx, fmt.Errorf("failed to init tf workspace: %w", err)
	}
	if err := ws.Apply(); err != nil {
		return ctx, fmt.Errorf("failed to apply tf workspace: %w", err)
	}

	return ctx, nil
}

/*
Then tf module "peeringTarget" is destroyed
*/
func tfModuleIsDestroyed(ctx context.Context, alias string) (context.Context, error) {
	session := GetCurrentScenarioSession(ctx)
	if session == nil {
		return ctx, ErrNoSession
	}

	ws := session.GetWorkspace(alias)
	if ws == nil {
		return ctx, fmt.Errorf("tf %q is not defined", alias)
	}

	if err := ws.Destroy(); err != nil {
		return ctx, fmt.Errorf("failed to destroy tf workspace: %w", err)
	}

	return ctx, godog.ErrPending
}
