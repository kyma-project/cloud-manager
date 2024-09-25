package dsl

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraScheme"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraTypes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateObj(ctx context.Context, clnt client.Client, obj client.Object, opts ...ObjAction) error {
	if obj == nil {
		return errors.New("obj given to Create() can not be nil")
	}

	switch infraScheme.ObjToClusterType(obj) {
	case infraTypes.ClusterTypeKcp:
		opts = append(opts, WithNamespace(DefaultKcpNamespace))
	case infraTypes.ClusterTypeSkr:
		opts = append(opts, WithNamespace(DefaultSkrNamespace))
	case infraTypes.ClusterTypeGarden:
		opts = append(opts, WithNamespace(DefaultGardenNamespace))
	}

	NewObjActions(opts...).
		ApplyOnObject(obj)

	err := clnt.Create(ctx, obj)

	return err
}
