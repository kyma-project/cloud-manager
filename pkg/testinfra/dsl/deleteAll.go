package dsl

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeleteAllOfSKR(infra testinfra.Infra) error {
	for _, obj := range []client.Object{
		&cloudresourcesv1beta1.IpRange{},
		&cloudresourcesv1beta1.AwsNfsVolume{},
	} {
		err := infra.SKR().Client().DeleteAllOf(infra.Ctx(), obj)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}
