package config

import (
	"encoding/json"

	"github.com/go-viper/mapstructure/v2"
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

	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:   b.obj,
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToIPHookFunc(),
			mapstructure.StringToIPNetHookFunc(),
			mapstructure.StringToByteHookFunc(),
			mapstructure.StringToBasicTypeHookFunc(),
		),
	})
	if err != nil {
		return
	}
	err = dec.Decode(data)
	if aclObj, ok := b.obj.(AfterConfigLoaded); ok {
		aclObj.AfterConfigLoaded()
	}
	if err != nil {
		return
	}
}
