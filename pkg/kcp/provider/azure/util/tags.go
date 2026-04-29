package util

func AzureTags(tags map[string]string) map[string]*string {
	var azureTags map[string]*string
	if tags != nil {
		azureTags = make(map[string]*string, len(tags))
		for k, v := range tags {
			azureTags[k] = new(v)
		}
	}
	return azureTags
}
