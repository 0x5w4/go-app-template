package shared

import (
	"reflect"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

func Sanitize(data any, policy *bluemonday.Policy) {
	if policy == nil {
		policy = bluemonday.UGCPolicy()
	}

	val := reflect.ValueOf(data)

	if val.Kind() != reflect.Pointer || val.IsNil() {
		return
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return
	}

	for i := range elem.NumField() {
		field := elem.Field(i)

		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			originalString := field.String()
			trimmed := strings.TrimSpace(originalString)
			sanitized := policy.Sanitize(trimmed)
			field.SetString(sanitized)

		case reflect.Struct:
			if field.Addr().CanInterface() {
				Sanitize(field.Addr().Interface(), policy)
			}

		case reflect.Pointer:
			if !field.IsNil() && field.Elem().Kind() == reflect.Struct {
				Sanitize(field.Interface(), policy)
			}
		}
	}
}
