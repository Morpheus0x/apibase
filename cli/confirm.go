package cli

import (
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
	out, err := c.askConfirmInternal()
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
	return c.askConfirmInternal()
}

func (c Confirm) AskBoolDefaultOnErr() bool {
	out, err := c.askConfirmInternal()
	if err != nil {
		return c.Default
	}
	return out
}

func (c Confirm) AskBoolFalseOnErr() bool {
	out, err := c.askConfirmInternal()
	if err != nil {
		return false
	}
	return out
}

// Internal ask method
func (c Confirm) askConfirmInternal() (bool, error) {
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
		return false, ERR_READLINE
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF, e.g. SIGINT
			return false, ERR_SIGINT_RECEIVED
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
			return false, ERR_INVALID_INPUT
		}
	}
}
