package redisinstance

import "cloud.google.com/go/redis/apiv1/redispb"

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

func AreEqualPolicies(current, desired *redispb.MaintenancePolicy) bool {
	if current == nil && desired == nil {
		return true
	}

	if (current != nil && desired == nil) || (current == nil && desired != nil) {
		return false
	}

	currentWindow := current.WeeklyMaintenanceWindow[0]
	desiredWindow := desired.WeeklyMaintenanceWindow[0]

	if currentWindow.Day.String() != desiredWindow.Day.String() {
		return false
	}

	if currentWindow.StartTime.Hours != desiredWindow.StartTime.Hours {
		return false
	}

	if currentWindow.StartTime.Minutes != desiredWindow.StartTime.Minutes {
		return false
	}

	return true
}
