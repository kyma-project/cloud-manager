package sapnfsvolumesnapshotrestore

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
)

func getLeaseName(volumeName string) string {
	return fmt.Sprintf("restore-%s", volumeName)
}

func getHolderName(ownerName types.NamespacedName) string {
	return fmt.Sprintf("%s/%s", ownerName.Namespace, ownerName.Name)
}
