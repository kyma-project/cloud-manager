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
