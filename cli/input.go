package cli

import (
	"github.com/chzyer/readline"
)

type ValidateFunc func(string) error

// Input Dialog
type Input struct {
	// default input to use when only pressing enter, shown in parenthesis after the prompt, if set
	Default string
	// input validation function
	Validate ValidateFunc
	// if kept as default false, the prompt is repeated until a valid anser is provided, otherwise if true, only asking once
	OnlyOnce bool
	// Prompt Text, colon with single trailing space will be appended as separator to user input
	Prompt string
}

func (i Input) Get() (string, error) {
	return i.getInputInternal()
}

func (i Input) GetEmptyOnErr() string {
	out, err := i.getInputInternal()
	if err != nil {
		return ""
	}
	return out
}

func (i Input) getInputInternal() (string, error) {
	defaultValue := ""
	if i.Default != "" {
		defaultValue = " (" + i.Default + ")"
	}
	rl, err := readline.New(i.Prompt + defaultValue + ": ")
	if err != nil {
		return "", ERR_READLINE
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF, e.g. SIGINT
			return "", ERR_SIGINT_RECEIVED
		}
		input := i.Default
		if len(line) > 0 {
			input = line
		}
		if i.Validate != nil {
			err = i.Validate(input)
			if err != nil {
				if i.OnlyOnce {
					return "", err
				}
				continue
			}
		}
		return input, nil
	}
}
