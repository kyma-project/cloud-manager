package tests

/*
Tests had to be extracted into a separate package due to cyclic dependency between packges

FAIL	github.com/kyma-project/cloud-manager/pkg/migrateFinalizers [setup failed]
# github.com/kyma-project/cloud-manager/pkg/migrateFinalizers
package github.com/kyma-project/cloud-manager/pkg/migrateFinalizers
	imports github.com/kyma-project/cloud-manager/pkg/testinfra
	imports github.com/kyma-project/cloud-manager/pkg/skr/runtime/looper
	imports github.com/kyma-project/cloud-manager/pkg/migrateFinalizers: import cycle not allowed in test

*/
