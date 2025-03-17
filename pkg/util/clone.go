package util

import "encoding/json"

func JsonClone[T any](a T) (T, error) {
	var res T
	b, err := json.Marshal(a)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(b, &res)
	return res, err
}

func JsonCloneInto[T any, R any](source T, destination R) error {
	b, err := json.Marshal(source)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &destination)
	return err
}
