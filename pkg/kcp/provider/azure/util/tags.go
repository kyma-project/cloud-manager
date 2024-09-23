package util

import "k8s.io/utils/ptr"

func AzureTags(tags map[string]string) map[string]*string {
	var azureTags map[string]*string
	if tags != nil {
		azureTags = make(map[string]*string, len(tags))
		for k, v := range tags {
			azureTags[k] = ptr.To(v)
		}
	}
	return azureTags
}
