package iprange

import (
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/utils/pointer"
)

const (
	tagKey = "cloud-manager.kyma-project.io/iprange"
)

func getTagValue(tags []ec2Types.Tag, key string) string {
	for _, t := range tags {
		if pointer.StringDeref(t.Key, "") == key {
			return pointer.StringDeref(t.Value, "")
		}
	}
	return ""
}

func nameEquals(tags []ec2Types.Tag, name string) bool {
	val := getTagValue(tags, "Name")
	return val == name
}
