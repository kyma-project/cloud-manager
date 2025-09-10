package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/stretchr/testify/assert"
)

type fileStruct struct {
	Defaults *fileStructDefaults `mapstructure:"defaults"`
}

type fileStructDefaults struct {
	IpRanges      int           `mapstructure:"ipranges"`
	AwsNfsVolumes int           `mapstructure:"awsnfsvolumes"`
	Duration      time.Duration `mapstructure:"duration"`
}

func TestConfigBind(t *testing.T) {
	// Given a file exists
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	err = copyFile("testdata/file-v1.yaml", filepath.Join(dir, "file.yaml"))
	assert.NoError(t, err)

	// When config is created from the file
	env := abstractions.NewMockedEnvironment(nil)
	cfg := NewConfig(env)
	cfg.SourceFile("", filepath.Join(dir, "file.yaml"))
	obj := &fileStruct{}
	cfg.Bind("", obj)
	cfg.Read()

	// Then bound object will have values from file
	assert.NotNilf(t, obj.Defaults, "expected obj.defaults not to be nil")
	assert.Equal(t, 1, obj.Defaults.IpRanges)
	assert.Equal(t, 3, obj.Defaults.AwsNfsVolumes)
	assert.Equal(t, 30*time.Second, obj.Defaults.Duration)

	// When file content changes
	err = copyFile("testdata/file-v2.yaml", filepath.Join(dir, "file.yaml"))
	assert.NoError(t, err)

	// And When config reloads
	cfg.Read()

	// Then bound object will have values from file
	assert.Equal(t, 2, obj.Defaults.IpRanges)
	assert.Equal(t, 4, obj.Defaults.AwsNfsVolumes)
}
