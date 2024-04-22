package config

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"sync"
)

func NewConfig(env abstractions.Environment) Config {
	return &config{
		defaults: "{}",
		env:      env,
	}
}

var _ Config = &config{}

type config struct {
	defaults string
	env      abstractions.Environment
	sources  []Source
	bindings []Binding
	js       string
}

func (c *config) Path(path FieldPath) *PathBuilder {
	return &PathBuilder{
		cfg:  c,
		path: path,
	}
}

func (c *config) GetAsString(path FieldPath) string {
	return gjson.Get(c.js, path.String()).String()
}

func (c *config) DefaultScalar(path FieldPath, scalar interface{}) {
	changed, err := sjson.Set(c.defaults, path.String(), scalar)
	if err != nil {
		return
	}
	c.defaults = changed
}

func (c *config) DefaultObj(path FieldPath, obj interface{}) {
	js, err := json.Marshal(obj)
	if err != nil {
		return
	}
	changed, err := sjson.SetRaw(c.defaults, path.String(), string(js))
	if err != nil {
		return
	}
	c.defaults = changed
}

func (c *config) DefaultJson(path FieldPath, js string) {
	changed, err := sjson.SetRaw(c.defaults, path.String(), js)
	if err != nil {
		return
	}
	c.defaults = changed
}

func (c *config) SourceFile(fieldPath FieldPath, file string) {
	c.sources = append(c.sources, &sourceFile{
		fieldPath: PrependFieldPath(fieldPath, "root"),
		file:      file,
	})
}

func (c *config) SourceEnv(fieldPath FieldPath, envVarPrefix string) {
	c.sources = append(c.sources, &sourceEnv{
		env:          c.env,
		fieldPath:    PrependFieldPath(fieldPath, "root"),
		envVarPrefix: envVarPrefix,
	})
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
		data := map[string]interface{}{}
		err := json.Unmarshal([]byte(c.js), &data)
		if err != nil {
			return
		}
		for _, b := range c.bindings {
			b.Copy(data)
		}
	}
}

func (c *config) Watch(stopCh chan struct{}, onConfigChange func(event fsnotify.Event)) error {
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
			//c.logger.Error(fmt.Sprintf("failed to create watcher: %s", err))
			returnErr = fmt.Errorf("error creating new watcher: %w", err)
			return
		}
		defer watcher.Close()

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

func (c *config) Bind(fieldPath FieldPath, obj any) {
	c.bindings = append(c.bindings, &binding{
		fieldPath: fieldPath,
		obj:       obj,
	})
}

func (c *config) Json() string {
	return c.js
}
