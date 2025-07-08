package leases

import (
	"context"
	"testing"

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/config"
	config2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/config"
	"github.com/stretchr/testify/suite"
	coordinationv1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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
	// arrange
	fakeClient := fake.NewClientBuilder().Build()
	skrScheme := runtime.NewScheme()
	client := composed.NewStateCluster(fakeClient, fakeClient, nil, skrScheme)
	leaseName := "test-lease"
	leaseNamespace := "test-namespace"
	owner := "test-owner"
	leaseDurationSec := int32(600)

	// act (owner acquires lease)
	res, err := Acquire(suite.ctx, client, leaseName, leaseNamespace, owner, leaseDurationSec)

	// assert
	suite.NoError(err)
	suite.Equal(AcquiredLease, res)

	// arrange
	lease := &coordinationv1.Lease{}
	err = fakeClient.Get(suite.ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	suite.NoError(err)
	time1 := lease.Spec.RenewTime.Time

	//act (owner extends lease)
	res, err = Acquire(suite.ctx, client, leaseName, leaseNamespace, owner, leaseDurationSec)

	// assert
	suite.NoError(err)
	err = fakeClient.Get(suite.ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	suite.NoError(err)
	suite.Equal(RenewedLease, res)
	time2 := lease.Spec.RenewTime.Time
	suite.True(time2.After(time1)) // make sure time2 is greater than time1

	// arrange
	otherOwner := "test-owner2"

	// act (owner owner tries to acquire same lease - fails)
	res, err = Acquire(suite.ctx, client, leaseName, leaseNamespace, otherOwner, leaseDurationSec)

	// assert
	suite.NoError(err)
	suite.Equal(OtherLeased, res)

	// act (other owner tries to release lease he doesnt own - fails)
	err = Release(suite.ctx, client, leaseName, leaseNamespace, otherOwner)

	// assert
	suite.Error(err)
	err = fakeClient.Get(suite.ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	suite.NoError(err)
	suite.Equal(owner, *lease.Spec.HolderIdentity)

	// act (lease is released by original owner)
	err = Release(suite.ctx, client, leaseName, leaseNamespace, owner)

	// assert
	suite.NoError(err)
	err = fakeClient.Get(suite.ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	suite.Error(err)
	suite.True(apierrors.IsNotFound(err))

	// act (nothing happens when releasing non-existant lease)
	err = Release(suite.ctx, client, leaseName, leaseNamespace, owner)

	// assert
	suite.NoError(err)
}

func TestLeaseSuite(t *testing.T) {
	suite.Run(t, new(leaseSuite))
}
