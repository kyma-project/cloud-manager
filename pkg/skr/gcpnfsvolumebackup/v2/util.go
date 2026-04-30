package v2

func getLocation(specLocation string, scopeRegion string) string {
	if len(specLocation) != 0 {
		return specLocation
	}
	return scopeRegion
}
