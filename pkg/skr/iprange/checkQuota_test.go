package iprange

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"testing"
	"time"
)

func TestCheckQuota_Sort(t *testing.T) {
	mustParseTime := func(s string) metav1.Time {
		r, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		return metav1.NewTime(r)
	}
	list := &cloudresourcesv1beta1.IpRangeList{
		Items: []cloudresourcesv1beta1.IpRange{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "newer",
					CreationTimestamp: mustParseTime("2024-05-20T10:20:30Z"),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "older",
					CreationTimestamp: mustParseTime("2024-05-05T10:20:30Z"),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "mid",
					CreationTimestamp: mustParseTime("2024-05-10T10:20:30Z"),
				},
			},
		},
	}

	sort.Slice(list.Items, func(i, j int) bool {
		return list.Items[i].CreationTimestamp.Before(&list.Items[j].CreationTimestamp)
	})

	assert.Equal(t, "older", list.Items[0].Name)
	assert.Equal(t, "mid", list.Items[1].Name)
	assert.Equal(t, "newer", list.Items[2].Name)
}
