package e2e

import (
	"fmt"
	"net/url"

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
	result := []string{
		"-L", // follow location 3xx redicts
		"-m",
	}
	if h.Method != "" {
		result = append(result, "-X", h.Method)
	}
	if h.ContentType != "" {
		result = append(result, "-H", "Content-Type: "+h.ContentType)
	}
	if h.Data != "" {
		result = append(result, "-d", h.Data)
	}
	result = append(result, h.Url)
	return result
}
