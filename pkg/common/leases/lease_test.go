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

func (ste *leaseSuite) SetupTest() {
	ste.ctx = context.Background()
	env := abstractions.NewMockedEnvironment(map[string]string{})
	cfg := config.NewConfig(env)
	config2.InitConfig(cfg)
	cfg.Read()
}

func (ste *leaseSuite) TestAcquireAndRelease() {
	// arrange
	fakeClient := fake.NewClientBuilder().Build()
	skrScheme := runtime.NewScheme()
	client := composed.NewStateCluster(fakeClient, fakeClient, nil, skrScheme)
	leaseName := "test-lease"
	leaseNamespace := "test-namespace"
	owner := "test-owner"
	leaseDurationSec := int32(600)

	// act (owner acquires lease)
	res, err := Acquire(ste.ctx, client, leaseName, leaseNamespace, owner, leaseDurationSec)

	// assert
	ste.NoError(err)
	ste.Equal(AcquiredLease, res)

	// arrange
	lease := &coordinationv1.Lease{}
	err = fakeClient.Get(ste.ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	ste.NoError(err)
	time1 := lease.Spec.RenewTime.Time

	//act (owner extends lease)
	res, err = Acquire(ste.ctx, client, leaseName, leaseNamespace, owner, leaseDurationSec)

	// assert
	ste.NoError(err)
	err = fakeClient.Get(ste.ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	ste.NoError(err)
	ste.Equal(RenewedLease, res)
	time2 := lease.Spec.RenewTime.Time
	ste.True(time2.After(time1)) // make sure time2 is greater than time1

	// arrange
	otherOwner := "test-owner2"

	// act (owner owner tries to acquire same lease - fails)
	res, err = Acquire(ste.ctx, client, leaseName, leaseNamespace, otherOwner, leaseDurationSec)

	// assert
	ste.NoError(err)
	ste.Equal(OtherLeased, res)

	// act (other owner tries to release lease he doesnt own - fails)
	err = Release(ste.ctx, client, leaseName, leaseNamespace, otherOwner)

	// assert
	ste.Error(err)
	err = fakeClient.Get(ste.ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	ste.NoError(err)
	ste.Equal(owner, *lease.Spec.HolderIdentity)

	// act (lease is released by original owner)
	err = Release(ste.ctx, client, leaseName, leaseNamespace, owner)

	// assert
	ste.NoError(err)
	err = fakeClient.Get(ste.ctx, types.NamespacedName{Name: leaseName, Namespace: leaseNamespace}, lease)
	ste.Error(err)
	ste.True(apierrors.IsNotFound(err))

	// act (nothing happens when releasing non-existant lease)
	err = Release(ste.ctx, client, leaseName, leaseNamespace, owner)

	// assert
	ste.NoError(err)
}

func TestLeaseSuite(t *testing.T) {
	suite.Run(t, new(leaseSuite))
}
