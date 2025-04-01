package util

import ctrl "sigs.k8s.io/controller-runtime"

func RequestObjToString(req ctrl.Request) string {
	if req.Namespace != "" {
		return req.String()
	}
	return req.Name
}
