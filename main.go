package main

import (
	"fmt"

	"github.com/Morpheus0x/apibase/log"
)

func main() {
	fmt.Printf("Err: %s", log.ErrEmptyString.String())
}
