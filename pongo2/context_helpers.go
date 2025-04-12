package pongo2

import (
	"reflect"

	"github.com/mitchellh/mapstructure"
)

func MarshalContext(data any) (Context, error) {
	result := make(Context)
	if data == nil {
		return result, nil
	}

	dataMap, ok := data.(map[string]any)
	if !ok {
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName: "pongo2",
			Result:  &dataMap,
		})
		if err != nil {
			return nil, err
		}
		if err := decoder.Decode(data); err != nil {
			return nil, err
		}
	}
	result = result.Update(dataMap)
	return result, nil
}

func UnmarshalContext(c Context, dst any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "pongo2",
		WeaklyTypedInput: true,
		Result:           dst,
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(valueConvertHook),
	})
	if err != nil {
		return err
	}
	if err := decoder.Decode(c); err != nil {
		return err
	}
	return nil
}

func valueConvertHook(from reflect.Type, to reflect.Type, data any) (any, error) {
	if from == reflect.TypeOf(&Value{}) {
		v := data.(*Value).Interface()
		if v == nil {
			return reflect.Zero(to).Interface(), nil
		}
		return v, nil
	}
	return data, nil
}
