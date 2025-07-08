package gcpnfsvolumerestore

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
)

func getLeaseName(resourceName, prefix string) string {
	if prefix != "" {
		return fmt.Sprintf("%s-%s", prefix, resourceName)
	}
	return resourceName
}
func getHolderName(ownerName types.NamespacedName) string {
	return fmt.Sprintf("%s/%s", ownerName.Namespace, ownerName.Name)
}
