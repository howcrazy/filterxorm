package filterxorm

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"
)

func paramsToSlice(items ...interface{}) []interface{} {
	if len(items) == 1 {
		val := reflect.Indirect(reflect.ValueOf(items[0]))
		if val.Kind() == reflect.Slice {
			items = make([]interface{}, val.Len())
			for i := 0; i < val.Len(); i++ {
				items[i] = val.Index(i).Interface()
			}
		}
	}
	return items
}

func funcName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "<error>"
	}
	name := runtime.FuncForPC(pc).Name()
	return name[strings.LastIndex(name, ".")+1:]
}

func _log(prefix string, args ...interface{}) {
	var msg string
	if len(args) > 0 {
		var format []string
		if len(args) == 1 {
			format = []string{"%s"}
		} else if val := reflect.Indirect(reflect.ValueOf(args[0])); val.Kind() == reflect.String {
			format, args = []string{val.String()}, args[1:len(args)]
		} else {
			format = []string{""}
			for i, _ := range args {
				format = append(format, fmt.Sprintf("%d:\t", i)+"%s")
			}
		}
		msg = fmt.Sprintf(strings.Join(format, "\n"), args...)
	}
	if prefix != "" {
		prefix = fmt.Sprintf("%s:", prefix)
	}
	log.Println(fmt.Sprintf("GoConv:%s%s", prefix, msg))
}

func debug(args ...interface{}) {
	_log("DEBUG", args...)
}

func warning(args ...interface{}) {
	_log("WARNING", args...)
}
