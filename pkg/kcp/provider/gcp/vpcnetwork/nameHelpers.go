package vpcnetwork

import "fmt"

func RouterName(name string) string {
	return fmt.Sprintf("%s-cloud-router", name)
}
