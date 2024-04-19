package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchBindWhenFileContentChanges(t *testing.T) {
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
	cfg.SourceFile(nil, filepath.Join(dir, "file.yaml"))
	obj := &fileStruct{}
	cfg.Bind(nil, obj)
	cfg.Read()

	stopCh := make(chan struct{})
	reloaded := false
	err = cfg.Watch(stopCh, func(_ fsnotify.Event) {
		reloaded = true
	})
	assert.NoError(t, err)

	// Then bound object will have values from file
	assert.NotNilf(t, obj.Defaults, "expected obj.defaults not to be nil")
	assert.Equal(t, 1, obj.Defaults.IpRanges)
	assert.Equal(t, 3, obj.Defaults.AwsNfsVolumes)

	// When file content changes
	err = copyFile("testdata/file-v2.yaml", filepath.Join(dir, "file.yaml"))
	assert.NoError(t, err)

	// give it time to consume fsnotify event and reload
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		if reloaded {
			break
		}
	}
	assert.True(t, reloaded, "expected config to reload")

	// Then bound object will have values from file
	assert.Equal(t, 2, obj.Defaults.IpRanges)
	assert.Equal(t, 4, obj.Defaults.AwsNfsVolumes)

	// stop the watcher
	close(stopCh)
}

func TestWatchRawWhenSymlinkReplaced(t *testing.T) {
	// Given a symlink exists
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	cwd, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Symlink(filepath.Join(cwd, "testdata/file-v1.yaml"), filepath.Join(dir, "file.yaml"))
	assert.NoError(t, err)

	// When config is created from the watched file
	env := abstractions.NewMockedEnvironment(nil)
	cfg := NewConfig(env)
	cfg.SourceFile(nil, filepath.Join(dir, "file.yaml"))
	cfg.Read()

	reloaded := false
	err = cfg.Watch(nil, func(_ fsnotify.Event) {
		reloaded = true
	})
	assert.NoError(t, err)

	// Then config will give values from file
	res := gjson.Get(cfg.Json(), "defaults.ipranges")
	assert.Equal(t, gjson.Number, res.Type)
	assert.Equal(t, "1", res.String())

	// When symlink changes
	err = os.Symlink(filepath.Join(cwd, "testdata/file-v2.yaml"), filepath.Join(dir, "file.yaml.tmp"))
	assert.NoError(t, err)
	err = os.Rename(filepath.Join(dir, "file.yaml.tmp"), filepath.Join(dir, "file.yaml"))
	assert.NoError(t, err)

	// give it time to consume fsnotify event and reload
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		if reloaded {
			break
		}
	}
	assert.True(t, reloaded, "expected config to reload")

	// Then config is reloaded and provides new values
	res = gjson.Get(cfg.Json(), "defaults.ipranges")
	assert.Equal(t, gjson.Number, res.Type)
	assert.Equal(t, "2", res.String())
}
