package dsl

import (
	"testing"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHavingFieldValue(t *testing.T) {

	t.Run("works with string", func(t *testing.T) {
		obj := &cloudcontrolv1beta1.NfsInstance{
			Status: cloudcontrolv1beta1.NfsInstanceStatus{
				State: "Ready",
			},
		}
		assert.NoError(t, HavingFieldValue("Ready", "status", "state")(obj))
		assert.Error(t, HavingFieldValue("Error", "status", "state")(obj))
	})

	t.Run("works with resource.Quantity as string", func(t *testing.T) {
		qty := util.Must(resource.ParseQuantity("1G"))

		obj := &cloudcontrolv1beta1.NfsInstance{
			Status: cloudcontrolv1beta1.NfsInstanceStatus{
				Capacity: qty,
			},
		}

		assert.NoError(t, HavingFieldValue("1G", "status", "capacity")(obj))
		assert.Error(t, HavingFieldValue("2M", "status", "capacity")(obj))
	})

	t.Run("works with zero resource.Quantity", func(t *testing.T) {
		obj := &cloudcontrolv1beta1.NfsInstance{
			Status: cloudcontrolv1beta1.NfsInstanceStatus{
				//Capacity: resource.Quantity{},
			},
		}

		assert.NoError(t, HavingFieldValue("0", "status", "capacity")(obj))
		assert.Error(t, HavingFieldValue("2M", "status", "capacity")(obj))
	})

	t.Run("works with metav1.Time as string", func(t *testing.T) {
		tm := metav1.Time{Time: util.Must(time.Parse(time.RFC3339, "2026-04-08T08:18:27Z"))}
		obj := &cloudcontrolv1beta1.NfsInstance{
			ObjectMeta: metav1.ObjectMeta{
				CreationTimestamp: tm,
			},
		}
		assert.NoError(t, HavingFieldValue("2026-04-08T08:18:27Z", "metadata", "creationTimestamp")(obj))
		assert.Error(t, HavingFieldValue("2001-01-01T01:01:01Z", "status", "capacity")(obj))
	})
}
