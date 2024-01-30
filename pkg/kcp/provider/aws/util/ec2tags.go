package util

import (
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/utils/pointer"
)

func Ec2Tags(args ...string) (result []ec2types.Tag) {
	ll := len(args)
	if ll == 0 {
		return nil
	}
	resultIndex := 0
	for i := 0; i < ll; i = i + 2 {
		result = append(result, ec2types.Tag{
			Key: pointer.String(args[i]),
		})
		if i < ll-1 {
			result[resultIndex].Value = pointer.String(args[i+1])
		}
		resultIndex++
	}
	return
}
