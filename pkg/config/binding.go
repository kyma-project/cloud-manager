package config

import (
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
)

type AfterConfigLoaded interface {
	AfterConfigLoaded()
}

type binding struct {
	fieldPath string
	obj       any
}

func (b *binding) Copy(in string) {
	str := in
	if len(b.fieldPath) > 0 {
		res := gjson.Get(in, b.fieldPath)
		if res.Type != gjson.JSON {
			return
		}
		str = res.String()
	}
	data := map[string]interface{}{}
	err := json.Unmarshal([]byte(str), &data)
	if err != nil {
		return
	}

	err = mapstructure.Decode(data, b.obj)
	if err != nil {
		return
	}
	if aclObj, ok := b.obj.(AfterConfigLoaded); ok {
		aclObj.AfterConfigLoaded()
	}
}
