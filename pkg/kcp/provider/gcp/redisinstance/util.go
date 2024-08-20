package redisinstance

func AreConfigsMissmatched(currentParameters, desiredParameters map[string]string) bool {
	if len(currentParameters) != len(desiredParameters) {
		return true
	}

	for key, desiredValue := range desiredParameters {
		currentValue, exists := currentParameters[key]
		if !exists || desiredValue != currentValue {
			return true
		}

	}

	return false
}
