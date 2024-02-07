package cli

import "fmt"

type Input struct {
	Default       string
	RequiredChars uint
	Prompt        string
}

func (i Input) Get() (string, error) {
	// TODO: cont
	return "", fmt.Errorf("")
}
