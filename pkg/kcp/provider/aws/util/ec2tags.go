package util

import (
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"k8s.io/utils/ptr"
)

func Ec2Tags(args ...string) (result []ec2types.Tag) {
	ll := len(args)
	if ll == 0 {
		return nil
	}
	resultIndex := 0
	for i := 0; i < ll; i = i + 2 {
		result = append(result, ec2types.Tag{
			Key: ptr.To(args[i]),
		})
		if i < ll-1 {
			result[resultIndex].Value = ptr.To(args[i+1])
		}
		resultIndex++
	}
	return
}

func GetEfsTagValue(tags []efsTypes.Tag, key string) string {
	for _, t := range tags {
		if ptr.Deref(t.Key, "") == key {
			return ptr.Deref(t.Value, "")
		}
	}
	return ""
}

func GetEc2TagValue(tags []ec2types.Tag, key string) string {
	for _, t := range tags {
		if ptr.Deref(t.Key, "") == key {
			return ptr.Deref(t.Value, "")
		}
	}
	return ""
}

func NameEc2TagEquals(tags []ec2types.Tag, name string) bool {
	val := GetEc2TagValue(tags, "Name")
	return val == name
}
