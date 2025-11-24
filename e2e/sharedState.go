package e2e

//
//type SharedState struct {
//	Runtimes []string `yaml:"runtimes,omitempty" json:"runtimes,omitempty"`
//}
//
//func LoadSharedState(filename string) (*SharedState, error) {
//	fi, err := os.Stat(filename)
//	if os.IsNotExist(err) {
//		return nil, err
//	}
//	if err != nil {
//		return nil, err
//	}
//	if fi.IsDir() {
//		return nil, fmt.Errorf("%s is a directory but file is expected", filename)
//	}
//
//	content, err := os.ReadFile(filename)
//	if err != nil {
//		return nil, err
//	}
//	state := &SharedState{}
//	if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
//		err = yaml.Unmarshal(content, state)
//	} else if strings.HasSuffix(filename, ".json") {
//		err = json.Unmarshal(content, state)
//	} else {
//		return nil, fmt.Errorf("file %s must have .yaml or .json extension", filename)
//	}
//	if err != nil {
//		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
//	}
//	return state, nil
//}
//
//func SaveSharedState(state *SharedState, filename string) error {
//	var content []byte
//	var err error
//	if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
//		content, err = yaml.Marshal(state)
//	} else if strings.HasSuffix(filename, ".json") {
//		content, err = json.Marshal(state)
//	} else {
//		return fmt.Errorf("file %s must have .yaml or .json extension", filename)
//	}
//	if err != nil {
//		return fmt.Errorf("failed to marshal state: %w", err)
//	}
//	err = os.WriteFile(filename, content, 0644)
//	if err != nil {
//		return fmt.Errorf("failed to save state: %w", err)
//	}
//	return nil
//}
