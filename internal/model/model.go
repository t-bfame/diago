package model

import (
	"github.com/davecgh/go-spew/spew"
)

var storage map[string]interface{}

func InitModel() {
	storage = make(map[string]interface{})

	InitTest()
	InitTestInstance()
}

func DumpStorage() {
	spew.Dump(storage)
}
