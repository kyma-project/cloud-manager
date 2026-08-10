package e2e

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/go-multierror"
)

type HttpOperation struct {
	Url            string
	Method         string
	ContentType    string
	Data           string
	MaxTime        int
	Retry          int
	ExpectedOutput string
}

func (h *HttpOperation) Validate() error {
	var result error
	if _, err := url.Parse(h.Url); err != nil {
		result = multierror.Append(result, err)
	}
	if h.ExpectedOutput == "" {
		result = multierror.Append(result, fmt.Errorf("expectedOutput is required"))
	}
	if h.MaxTime == 0 {
		h.MaxTime = 10
	}

	return result
}

func (h *HttpOperation) Args() []string {
	curlArgs := []string{
		"curl",
		"-L", // follow location 3xx redirects
		"-m", fmt.Sprintf("%d", h.MaxTime),
	}
	if h.Retry > 0 {
		// By default, curl's --retry only retries on transient HTTP-level errors (5xx responses, transfer timeouts). It does not retry on connection-level failures like:
		// - Error 7: Connection refused
		// - Error 6: Could not resolve host
		// Without --retry-all-errors, curl exits immediately on these errors instead of retrying.
		curlArgs = append(curlArgs, "--retry", fmt.Sprintf("%d", h.Retry))
		curlArgs = append(curlArgs, "--retry-all-errors")
	}
	if h.Method != "" {
		curlArgs = append(curlArgs, "-X", h.Method)
	}
	if h.ContentType != "" {
		curlArgs = append(curlArgs, "-H", "Content-Type: "+h.ContentType)
	}
	if h.Data != "" {
		curlArgs = append(curlArgs, "-d", h.Data)
	}
	curlArgs = append(curlArgs, h.Url)

	return []string{strings.Join(curlArgs, " ")}
}
