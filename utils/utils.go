package utils

import (
	"fmt"
	"reflect"
	"runtime"
)

func Try(function func(), handler func(interface{})) {
	defer func() {
		err := recover()
		if err != nil {
			handler(err)
		}
	}()
	function()
}

func String(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
