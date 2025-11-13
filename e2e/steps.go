package e2e

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
)

func errEvalContextBuilding(err error) error {
	return fmt.Errorf("error building evaluation context: %w", err)
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

	txt, err := eval.EvalTemplate(doc.Content)
	if err != nil {
		return ctx, err
	}
	arr, err := util.YamlMultiDecodeToUnstructured([]byte(txt))
	if err != nil {
		return ctx, fmt.Errorf("failed to parse resource yaml: %w", err)
	}
	if len(arr) != 1 {
		return ctx, fmt.Errorf("expected one resource in yaml but got %d", len(arr))
	}
	obj := arr[0]

	if obj.GetNamespace() == "" {
		obj.SetNamespace(ri.Namespace)
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

	| Operation | Path    | Content      |
	| Create    | foo.txt | some content |
	| Append    | foo.txt | some more    |
	| Delete    | foo.txt |              |
	| Contains  | foo.txt | content      |
	| Exists    | foo.txt |              |
*/
func pvcFileOperationsSucceed(ctx context.Context, alias string, ops *godog.Table) (context.Context, error) {
	arr, err := ad.ParseSlice(ops)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse operations, the table must have first header row with colums Operation, Path, Content: %w", err)
	}
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

	rootDir := ri.Name
	name := "e2e-pvc-op-" + util.RandomString(6)
	fileOps = append(fileOps, EchoOperation(allDone))
	scriptLines := CombineFileOperations(fileOps...)(rootDir)
	b := NewPodBuilder(name, ri.Namespace, "ubuntu").
		WithPodDetails(
			PodWithScript(scriptLines),
			PodWithMountFromPVC(ri.Name, "", ""),
		)
	err = b.Create(ctx, session.CurrentCluster().GetClient())
	if err != nil {
		return ctx, fmt.Errorf("error creating pvc operation resources: %w", err)
	}
	session.CurrentCluster().DeleteOnTerminate(b.Pod())
	session.CurrentCluster().DeleteOnTerminate(b.ExtraResourceObjects()...)

	err = session.CurrentCluster().AddResources(ctx, &ResourceDeclaration{
		Alias:      name,
		Kind:       "Pod",
		ApiVersion: "v1",
		Name:       name,
		Namespace:  ri.Namespace,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to declare pvc pod resource: %w", err)
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

	logs, err := session.CurrentCluster().PodLogs(ctx, ri.Namespace, name, name)
	if err != nil {
		return ctx, err
	}

	err = b.Delete(ctx, session.CurrentCluster().GetClient())
	if err != nil {
		return ctx, fmt.Errorf("error deleting pvc operation resources: %w", err)
	}

	if failed {
		return ctx, fmt.Errorf("pvc operation failed:\n%s", logs)
	}

	if strings.Contains(logs, allDone) {
		return ctx, nil
	}

	return ctx, fmt.Errorf("pvc operation did not succeeded:\n%s", logs)
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

	name := "e2e-http-op-" + util.RandomString(6)
	b := NewPodBuilder(name, world.Config().SkrNamespace, "curlimages/curl").
		WithPodDetails(
			PodWithArguments(op.Args()...),
		)
	if err := session.CurrentCluster().AddResources(ctx, &ResourceDeclaration{
		Alias:      name,
		Kind:       "Pod",
		ApiVersion: "v1",
		Name:       name,
		Namespace:  world.Config().SkrNamespace,
	}); err != nil {
		return ctx, fmt.Errorf("failed to declare http operation resources: %w", err)
	}
	if err := b.Create(ctx, session.CurrentCluster().GetClient()); err != nil {
		return ctx, fmt.Errorf("failed to create http operation resources: %w", err)
	}

	session.CurrentCluster().DeleteOnTerminate(b.Pod())

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

	name := "e2e-redis-op-" + util.RandomString(6)

	b := NewPodBuilder(name, world.Config().SkrNamespace, "redis")

	makeEnv := func(envVarName string, row *messages.PickleTableRow) error {
		switch row.Cells[1].Value {
		case "Secret":
			b.WithPodDetails(PodWithEnvFromSecret(envVarName, row.Cells[1].Value, row.Cells[2].Value))
		case "ConfigMap":
			b.WithPodDetails(PodWithEnvFromConfigMap(envVarName, row.Cells[1].Value, row.Cells[2].Value))
		case "Fixed":
			b.WithPodDetails(PodWithFixedEnvVar(envVarName, row.Cells[1].Value))
		default:
			return fmt.Errorf("invalid value indicator %q", row.Cells[1].Value)
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
			b.WithPodDetails(PodWithFixedEnvVar("CA", row.Cells[2].Value))
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

	if err := session.CurrentCluster().AddResources(ctx, &ResourceDeclaration{
		Alias:      name,
		Kind:       "Pod",
		ApiVersion: "v1",
		Name:       name,
		Namespace:  world.Config().SkrNamespace,
	}); err != nil {
		return ctx, fmt.Errorf("failed to declare redis operation pod: %w", err)
	}

	if err := b.Create(ctx, session.CurrentCluster().GetClient()); err != nil {
		return ctx, fmt.Errorf("failed to create redis operation pod: %w", err)
	}
	session.CurrentCluster().DeleteOnTerminate(b.Pod())
	session.CurrentCluster().DeleteOnTerminate(b.ExtraResourceObjects()...)

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

	if err := b.Delete(ctx, session.CurrentCluster().GetClient()); err != nil {
		return ctx, fmt.Errorf("failed to delete redis operation pod: %w", err)
	}

	if failed {
		return ctx, fmt.Errorf("redis operation failed:\n%s", logs)
	}

	if !strings.Contains(logs, out) {
		return ctx, fmt.Errorf("redis operation did not return expected %q:\n%s", out, logs)
	}

	return ctx, nil
}
