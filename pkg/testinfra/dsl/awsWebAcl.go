package dsl

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAwsWebAcl(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AwsWebAcl, opts ...ObjAction) error {
	NewObjActions(opts...).ApplyOnObject(obj)

	err := clnt.Create(ctx, obj)
	return err
}
