package utils

import (
	"errors"
	"fmt"
	"reflect"
)

func ToMap[T any](req T, typee string) (any, error) {
	sort := map[string]string{}
	filter := map[string][]string{}
	v := reflect.ValueOf(req)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fV := v.Field(i)
		fN := t.Field(i).Name
		if !fV.IsZero() {
			switch typee {
			case "sort":
				sort[fN] = fV.Interface().(string)
			case "filter":
				filter[fN] = fV.Interface().([]string)
			default:
				return nil, errors.New("invalid map type")
			}
		}
	}
	fmt.Println(sort)
	fmt.Println(filter)
	switch typee {
	case "sort":
		return sort, nil
	case "filter":
		return filter, nil
	default:
		return nil, errors.New("invalid map type")
	}
}
