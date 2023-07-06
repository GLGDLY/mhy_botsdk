package utils

import (
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
)

func Try(function func(), handler func(interface{}, string)) {
	defer func() {
		err := recover()
		if err != nil {
			handler(err, "Traceback: \n"+string(debug.Stack()))
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
