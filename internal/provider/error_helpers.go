package provider

import (
	"reflect"
	"strings"
)

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "HTTP 404") || strings.Contains(err.Error(), "status 404") {
		return true
	}
	for err != nil {
		if hasStatusCode(err, 404) {
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	return false
}

func hasStatusCode(err error, code int64) bool {
	if err == nil {
		return false
	}
	v := reflect.ValueOf(err)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		if f := v.FieldByName("StatusCode"); f.IsValid() && f.CanInt() {
			return f.Int() == code
		}
	}
	return false
}
