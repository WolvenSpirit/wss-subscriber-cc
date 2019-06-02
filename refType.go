package main

import (
	"reflect"
)

func returnType(data ...interface{}) string {
	for _, v := range data {
		o := reflect.TypeOf(v).String()
		return o
	}
	return ""
}
