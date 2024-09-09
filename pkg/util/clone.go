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
