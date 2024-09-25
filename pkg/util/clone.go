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

func JsonCloneInto[T any, R any](a T, r R) error {
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &r)
	return err
}
