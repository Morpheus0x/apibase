package cli

import (
	"fmt"

	h "github.com/KiloHost/apibase/helper"
)

// Basic Yes/No Dialog
type Confirm struct {
	// Default value is false, meaning no
	Default h.ConfirmDefault
	// How strict the input parsing should be
	Sensitivity h.ConfirmSensitivity
	// Prompt Text
	Prompt string
}

// returns the string "yes" or "no" depending on user input, on error undefined
func (c *Confirm) Ask() (string, error) {
	out, err := c.askConfirm()
	if err != nil {
		return "", err
	}
	if out {
		return "yes", nil
	}
	return "no", nil
}

// The returned bool is always true for yes and false for no, independent of the set default response
func (c *Confirm) AskBool() (bool, error) {
	return c.askConfirm()
}

// Internal ask method
func (c *Confirm) askConfirm() (bool, error) {
	q := c.Prompt
	if c.Prompt == "" {
		q = "Confirm?"
	}
	// TODO: cont
	return false, fmt.Errorf(q)
}

type Input struct {
	Default       string
	RequiredChars uint
	Prompt        string
}

func (i *Input) Get() (string, error) {
	// TODO: cont
	return "", fmt.Errorf("")
}
