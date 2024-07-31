package main

import (
	"fmt"

	"gopkg.cc/apibase/log"
)

func main() {
	fmt.Printf("Err: %s", log.ErrEmptyString.String())
}
