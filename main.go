package main

import (
	"gopkg.cc/apibase/log"
)

func main() {
	ErrEmptyString := log.RegisterErrType("ErrEmptyString")
	log.NewErrorWithType(ErrEmptyString, "no data").Log()
}
