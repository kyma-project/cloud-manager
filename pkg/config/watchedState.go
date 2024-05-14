package config

import (
	"github.com/fsnotify/fsnotify"
	"path/filepath"
	"sync"
)

type watchedState struct {
	sync.Mutex

	items map[string]*watchedStateItem
}

type watchedStateItem struct {
	previousStat *sourceFileStat
	currentStat  *sourceFileStat
}

func (w *watchedState) uniqueDirsToWatch() []string {
	unique := map[string]struct{}{}
	for _, item := range w.items {
		unique[item.currentStat.configDir] = struct{}{}
	}
	result := make([]string, len(unique))
	for k := range unique {
		result = append(result, k)
	}
	return result
}

func (w *watchedState) stat(sources []Source) {
	w.Lock()
	defer w.Unlock()

	if w.items == nil {
		w.items = map[string]*watchedStateItem{}
	}

	// take a shapshot of current file sources stats
	snapshot := map[string]*sourceFileStat{}
	for _, src := range sources {
		if fileSrc, ok := src.(*sourceFile); ok {
			stat := fileSrc.stat()
			if stat != nil {
				snapshot[stat.filename] = stat
			}
		}
	}

	// add or update stats from the taken snapshot
	for filename, stat := range snapshot {
		item, exists := w.items[filename]
		if !exists {
			w.items[filename] = &watchedStateItem{
				currentStat: stat,
			}
		} else {
			item.previousStat = item.currentStat
			item.currentStat = stat
		}
	}

	// remove items that no longer exist
	// this might happen in case file source returned this time nil if some error occurred
	var itemsToRemove []string
	for filename := range w.items {
		_, exists := w.items[filename]
		if !exists {
			itemsToRemove = append(itemsToRemove, filename)
		}
	}
	for _, filename := range itemsToRemove {
		delete(w.items, filename)
	}
}

func (w *watchedState) hasChange(event fsnotify.Event) bool {
	w.Lock()
	defer w.Unlock()

	// we only care about the config file with the following cases:
	// 1 - if the config file was modified or created
	// 2 - if the real path to the config file changed (eg: k8s ConfigMap replacement)

	// go through items and determine if change has happened
	for _, item := range w.items {
		// the config file was modified or created
		if filepath.Clean(event.Name) == item.currentStat.configFile &&
			(event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) {
			return true
		}
		// the real path to the config file changed (eg: k8s ConfigMap replacement)
		oldRealConfigFile := ""
		if item.previousStat != nil {
			oldRealConfigFile = item.previousStat.realConfigFile
		}
		currentRealConfigFile := item.currentStat.realConfigFile
		if currentRealConfigFile != "" && currentRealConfigFile != oldRealConfigFile {
			return true
		}
	}

	return false
}
