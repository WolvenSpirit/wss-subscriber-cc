package main

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

func TestReturnType(t *testing.T) {
	b, e := json.Marshal(&struct {
		Name   string
		Amount float64
	}{"foo", 4.2})
	if e != nil {
		log.Println(e.Error())
	}
	var i interface{}
	json.Unmarshal(b, &i)
	if fmt.Sprintf(returnType(i)) != "map[string]interface {}" {
		t.Fail()
	}
}
