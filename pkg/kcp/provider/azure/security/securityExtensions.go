package security

import "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"

func hasAllExtensionsEnabled(current []*armsecurity.Extension, required []armsecurity.Extension) bool {
	for _, req := range required {
		if req.Name == nil {
			continue
		}
		found := false
		for _, cur := range current {
			if cur.Name != nil && *cur.Name == *req.Name {
				if cur.IsEnabled != nil && *cur.IsEnabled == armsecurity.IsEnabledTrue {
					found = true
				}
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func toExtensionPointers(extensions []armsecurity.Extension) []*armsecurity.Extension {
	result := make([]*armsecurity.Extension, len(extensions))
	for i := range extensions {
		result[i] = &extensions[i]
	}
	return result
}
