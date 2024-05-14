package config

import (
	"encoding/json"
	"github.com/peterbourgon/mergemap"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type sourceFile struct {
	fieldPath string
	file      string
}

type sourceFileStat struct {
	filename       string // original filename
	configFile     string // filepath.Clean
	configDir      string // filepath.Split
	realConfigFile string // filepath.EvalSymlinks
}

func (s *sourceFile) stat() *sourceFileStat {
	configDir, _ := filepath.Split(s.file)
	realConfigFile, err := filepath.EvalSymlinks(s.file)
	if err != nil {
		return nil
	}
	return &sourceFileStat{
		filename:       s.file,
		configFile:     filepath.Clean(s.file),
		configDir:      configDir,
		realConfigFile: realConfigFile,
	}
}

func (s *sourceFile) Read(inJsonString string) string {
	fileInfo, err := os.Stat(s.file)
	if err != nil {
		return inJsonString
	}
	if fileInfo.IsDir() {
		return inJsonString
	}
	buf, err := os.ReadFile(s.file)
	if err != nil {
		return inJsonString
	}
	loadedFileString := string(buf)
	ext := filepath.Ext(s.file)

	var newJsonString string
	var newData map[string]interface{}

	switch ext {
	case ".json":
		newJsonString = loadedFileString
	case ".yaml", ".yml":
		newData = map[string]interface{}{}
		err = yaml.Unmarshal([]byte(loadedFileString), newData)
		if err != nil {
			return inJsonString
		}
		jsData, err := json.Marshal(newData)
		if err != nil {
			return inJsonString
		}

		newJsonString = string(jsData)
	default:
		changed, err := sjson.Set(inJsonString, s.fieldPath, loadedFileString)
		if err != nil {
			return inJsonString
		}
		return changed
	}

	existingGJsonResult := gjson.Get(inJsonString, s.fieldPath)
	if existingGJsonResult.Type == gjson.Null {
		changedJsonString, err := sjson.SetRaw(inJsonString, s.fieldPath, newJsonString)
		if err != nil {
			return inJsonString
		}
		return changedJsonString
	}

	existingData := map[string]interface{}{}
	err = json.Unmarshal([]byte(existingGJsonResult.Raw), &existingData)
	if err != nil {
		return inJsonString
	}

	if newData == nil {
		err = json.Unmarshal([]byte(newJsonString), &newData)
		if err != nil {
			return inJsonString
		}
	}

	mergedData := mergemap.Merge(existingData, newData)
	mergedJsonBytes, err := json.Marshal(mergedData)
	if err != nil {
		return inJsonString
	}
	mergedJsonString := string(mergedJsonBytes)
	changedJsonString, err := sjson.SetRaw(inJsonString, s.fieldPath, mergedJsonString)
	if err != nil {
		return inJsonString
	}
	return changedJsonString
}
