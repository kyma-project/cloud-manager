package debugged

func When[T any](yes, no T) T {
	if Debugged {
		return yes
	}
	return no
}
