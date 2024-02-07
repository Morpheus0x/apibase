package cli

import (
	"github.com/chzyer/readline"
)

// Input Dialog
type Input struct {
	// default input to use when only pressing enter, shown in parenthesis after the prompt, if set
	Default string
	// minimum number of required characters
	RequiredChars uint
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
		if len(line) == 0 {
			if len(i.Default) >= int(i.RequiredChars) {
				return i.Default, nil
			} else {
				if i.OnlyOnce {
					return "", ERR_MIN_INPUT_LEN
				} else {
					continue
				}
			}
		}
		if len(line) < int(i.RequiredChars) {
			if i.OnlyOnce {
				return "", ERR_MIN_INPUT_LEN
			} else {
				continue
			}
		}
		return line, nil
	}
}
