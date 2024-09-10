package leases

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/config"
	config2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/config"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

type leaseSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *leaseSuite) SetupTest() {
	suite.ctx = context.Background()
	env := abstractions.NewMockedEnvironment(map[string]string{})
	cfg := config.NewConfig(env)
	config2.InitConfig(cfg)
	cfg.Read()
}

func (suite *leaseSuite) TestAcquireAndRelease() {
	fakeClient := fake.NewClientBuilder().Build()
	skrScheme := runtime.NewScheme()
	client := composed.NewStateCluster(fakeClient, fakeClient, nil, skrScheme)
	resource := types.NamespacedName{Name: "test-resource", Namespace: "resource-namespace"}
	owner := types.NamespacedName{Name: "test-owner", Namespace: "owner-namespace"}
	prefix := "test-prefix"
	res, err := Acquire(suite.ctx, client, resource, owner, prefix)
	suite.NoError(err)
	suite.Equal(AcquiredLease, res)
	leaseName := getLeaseName(resource, prefix)
	leaseNamespace := resource.Namespace
	lease := &v1.Lease{}
	err = fakeClient.Get(suite.ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	suite.NoError(err)
	time1 := lease.Spec.RenewTime.Time
	res, err = Acquire(suite.ctx, client, resource, owner, prefix)
	suite.NoError(err)
	err = fakeClient.Get(suite.ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	suite.NoError(err)
	suite.Equal(RenewedLease, res)
	time2 := lease.Spec.RenewTime.Time
	// make sure time2 is greater than time1
	suite.True(time2.After(time1))
	res, err = Acquire(suite.ctx, client, resource, types.NamespacedName{Name: "test-owner", Namespace: "owner-namespace2"}, prefix)
	suite.NoError(err)
	suite.Equal(OtherLeased, res)
	err = Release(suite.ctx, client, resource, types.NamespacedName{Name: "test-owner", Namespace: "owner-namespace2"}, prefix)
	suite.NoError(err)
	err = fakeClient.Get(suite.ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	suite.NoError(err)
	suite.Equal(getHolderName(owner), *lease.Spec.HolderIdentity)
	err = Release(suite.ctx, client, resource, owner, prefix)
	suite.NoError(err)
	err = fakeClient.Get(suite.ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	suite.Error(err)
	suite.True(apierrors.IsNotFound(err))
	err = Release(suite.ctx, client, resource, owner, prefix)
	suite.NoError(err)
	err = Release(suite.ctx, client, resource, owner, prefix)
	suite.NoError(err)
}

func TestLeaseSuite(t *testing.T) {
	suite.Run(t, new(leaseSuite))
}
