package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func NewConfig(env abstractions.Environment) Config {
	return &config{
		defaults: "{}",
		env:      env,
	}
}

var _ Config = &config{}

type config struct {
	baseDir   string
	defaults  string
	env       abstractions.Environment
	sources   []Source
	bindings  []Binding
	sensitive []string
	js        string
}

func (c *config) BaseDir(baseDir string) {
	c.baseDir = baseDir
}

func (c *config) GetBaseDir() string {
	return c.baseDir
}

func (c *config) GetEnv(key string) string {
	return c.env.Get(key)
}

func (c *config) Path(path string, actions ...PathAction) Config {
	for _, a := range actions {
		a(c, path)
	}
	return c
}

func (c *config) GetAsString(path string) string {
	return gjson.Get(c.js, path).String()
}

func (c *config) DefaultScalar(path string, scalar interface{}) {
	changed, err := sjson.Set(c.defaults, path, scalar)
	if err != nil {
		return
	}
	c.defaults = changed
}

func (c *config) DefaultObj(path string, obj interface{}) {
	js, err := json.Marshal(obj)
	if err != nil {
		return
	}
	changed, err := sjson.SetRaw(c.defaults, path, string(js))
	if err != nil {
		return
	}
	c.defaults = changed
}

func (c *config) DefaultJson(path string, js string) {
	changed, err := sjson.SetRaw(c.defaults, path, js)
	if err != nil {
		return
	}
	c.defaults = changed
}

func (c *config) SourceFile(fieldPath string, file string) {
	if len(c.baseDir) > 0 {
		file = filepath.Join(c.baseDir, file)
	}
	c.sources = append(c.sources, &sourceFile{
		fieldPath: ConcatFieldPath("root", fieldPath),
		file:      file,
	})
}

func (c *config) SourceEnv(fieldPath string, envVarPrefix string) {
	c.sources = append(c.sources, &sourceEnv{
		env:          c.env,
		fieldPath:    ConcatFieldPath("root", fieldPath),
		envVarPrefix: envVarPrefix,
	})
}

func (c *config) Sensitive(fieldPath string) {
	c.sensitive = append(c.sensitive, fieldPath)
}

func (c *config) Read() {
	emptyRoot := "{\"root\":{}}"
	js, err := sjson.SetRaw(emptyRoot, "root", c.defaults)
	if err != nil {
		return
	}
	for _, src := range c.sources {
		js = src.Read(js)
	}
	c.js = gjson.Get(js, "root").Raw

	if len(c.bindings) > 0 {
		for _, b := range c.bindings {
			b.Copy(c.js)
		}
	}
}

func (c *config) Watch(stopCh <-chan struct{}, onConfigChange func(event fsnotify.Event)) error {
	state := &watchedState{}

	// if nil provided, it won't ever be closed, aka watch stopped, make a stub instead
	if stopCh == nil {
		stopCh = make(chan struct{})
	}

	initWG := sync.WaitGroup{}
	initWG.Add(1)
	var returnErr error
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			returnErr = fmt.Errorf("error creating new watcher: %w", err)
			return
		}
		defer func(watcher *fsnotify.Watcher) {
			_ = watcher.Close()
		}(watcher)

		// we have to watch the entire directory to pick up renames/atomic saves in a cross-platform way

		// take initial snapshot
		state.stat(c.sources)

		eventsWG := sync.WaitGroup{}
		eventsWG.Add(1)
		go func() {
			for {
				select {
				case <-stopCh:
					eventsWG.Done()
					return
				case event, ok := <-watcher.Events:
					if !ok { // 'Events' channel is closed
						eventsWG.Done()
						return
					}
					// take a new snapshot
					state.stat(c.sources)

					// we only care about the config file with the following cases:
					// 1 - if the config file was modified or created
					// 2 - if the real path to the config file changed (eg: k8s ConfigMap replacement)
					if state.hasChange(event) {
						c.Read()

						if onConfigChange != nil {
							onConfigChange(event)
						}
					}

				case err, ok := <-watcher.Errors:
					if ok { // 'Errors' channel is not closed
						//v.logger.Error(fmt.Sprintf("watcher error: %s", err))
						returnErr = fmt.Errorf("watcher error: %w", err)
					}
					eventsWG.Done()
					return
				}
			}
		}()

		for _, dir := range state.uniqueDirsToWatch() {
			err := watcher.Add(dir)
			if err != nil {
				returnErr = fmt.Errorf("error adding dir %s for watching: %w", dir, err)
				_ = watcher.Close()
				break
			}
		}
		initWG.Done()   // done initializing the watch in this go routine, so the parent routine can move on...
		eventsWG.Wait() // now, wait for event loop to end in this go-routine...
	}()
	initWG.Wait() // make sure that the go routine above fully ended before returning
	return returnErr
}

func (c *config) Bind(fieldPath string, obj any) {
	c.bindings = append(c.bindings, &binding{
		fieldPath: fieldPath,
		obj:       obj,
	})
}

func (c *config) Json() string {
	return c.js
}

func (c *config) PrintJson() string {
	js := c.js
	for _, sensitivePath := range c.sensitive {
		jsjs, err := sjson.Set(js, sensitivePath, "*****")
		if err != nil {
			return fmt.Sprintf("error: %s", js)
		}
		js = jsjs
	}
	return js
}
