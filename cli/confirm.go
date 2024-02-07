package cli

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
)

// Basic Yes/No Dialog
type Confirm struct {
	// Default value is false, meaning no
	Default bool
	// If kept as default false, the prompt is repeated until a valid anser is provided, otherwise if true, only asking once
	OnlyOnce bool
	// Prompt Text
	Prompt string
}

// returns the string "yes" or "no" depending on user input, on error undefined
func (c Confirm) Ask() (string, error) {
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
func (c Confirm) AskBool() (bool, error) {
	return c.askConfirm()
}

func (c Confirm) AskBoolDefaultOnErr() bool {
	out, err := c.askConfirm()
	if err != nil {
		return c.Default
	}
	return out
}

func (c Confirm) AskBoolFalseOnErr() bool {
	out, err := c.askConfirm()
	if err != nil {
		return false
	}
	return out
}

// Internal ask method, panics
func (c Confirm) askConfirm() (bool, error) {
	prompt := c.Prompt
	if c.Prompt == "" {
		prompt = "Confirm?"
	}
	question := " [y/N]: "
	if c.Default {
		question = " [Y/n]: "
	}

	rl, err := readline.New(prompt + question)
	if err != nil {
		return false, err
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			// return false, fmt.Errorf("^C received")
			panic(fmt.Errorf("^C received"))
		}
		if len(line) == 0 {
			return c.Default, nil
		}

		input := strings.ToLower(line)
		if input == "y" || input == "ye" || input == "yes" {
			return true, nil
		}
		if input == "n" || input == "no" {
			return false, nil
		}
		if c.OnlyOnce {
			return false, fmt.Errorf("no valid input specified")
		}
	}
}
