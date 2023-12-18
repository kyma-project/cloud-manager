package common

/*

This file must exist empty to make controller-gen happy, so it thinks
package github.com/kyma-project/cloud-resources/components/skr/pkg/common exists.

Otherwise, if there are no go files in this dir, it will brake with this error:

go/src/github.com/kyma-project/cloud-resources/components/skr/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
pkg/cloudresources/reconcile/reconciler.go:6:2: no required module provides package github.com/kyma-project/cloud-resources/components/skr/pkg/common; to add it:
	go get github.com/kyma-project/cloud-resources/components/skr/pkg/common

*/
