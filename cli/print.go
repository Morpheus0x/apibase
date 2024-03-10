package cli

import "fmt"

func PrintError(err error) {
	//TODO: colorize output, add global error check and return error id
	fmt.Printf("Error Occurred: %s", err.Error())
}

func PrintErrorMsg(msg string) {
	//TODO: colorize output, add global error check and return error id
	fmt.Printf("Error Occurred: %s", msg)
}
