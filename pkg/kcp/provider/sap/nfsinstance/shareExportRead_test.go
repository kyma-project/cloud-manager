package nfsinstance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseExportUrl(t *testing.T) {
	testData := []struct {
		url  string
		host string
		path string
	}{
		{
			url:  "10.11.12.13:/share_1c3da1a4_3263_4779_b230_f9d496326dc6",
			host: "10.11.12.13",
			path: "share_1c3da1a4_3263_4779_b230_f9d496326dc6",
		},
		{
			url:  "10.11.12.13/share_1c3da1a4_3263_4779_b230_f9d496326dc6",
			host: "10.11.12.13",
			path: "share_1c3da1a4_3263_4779_b230_f9d496326dc6",
		},
		{
			url:  "example.com/aaa/bbb",
			host: "example.com",
			path: "aaa/bbb",
		},
	}

	for _, data := range testData {
		t.Run(data.url, func(t *testing.T) {
			h, p, err := parseExportUrl(data.url)
			assert.NoError(t, err)
			assert.Equal(t, data.host, h)
			assert.Equal(t, data.path, p)
		})
	}
}
