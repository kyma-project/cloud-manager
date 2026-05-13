package gardener

// SetGardenerNamespaceProviderMock use only in tests!!!!
func SetGardenerNamespaceProviderMock(value string) {
	defaultGardenerNamespaceProvider = &fixedValueGardenerNamespaceProvider{
		value: value,
	}
}
