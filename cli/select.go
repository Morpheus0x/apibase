package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

type Select struct {
	Options  []string
	Defaults []bool
	Validate func(sel []bool) error
	Multiple bool
	OnlyOnce bool
	Prompt   string
}

func (s Select) GetOne() (string, error) {
	if s.Multiple {
		return "", errors.New("must only be used if Multiple is false")
	}
	res, err := s.Get()
	if err != nil {
		return "", err
	}
	out := ""
	selCnt := 0
	for idx, yes := range res {
		if yes {
			selCnt++
			out = s.Options[idx]
		}
	}
	if selCnt == 1 {
		return out, nil
	}
	return "", errors.New("more than one was selected")
}

func (s Select) Get() ([]bool, error) {
	if s.Prompt == "" {
		return []bool{}, errors.New("prompt is required")
	}
	fmt.Printf("%s\n", s.Prompt)
	// Force terminal into raw mode to disable echoing of control sequences.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}
	// Reset terminal back to default
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Hide Cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	optionsLen := len(s.Options)
	defaultsLen := len(s.Defaults)
	if optionsLen < 1 {
		return nil, errors.New("no options provided")
	}
	if defaultsLen > 0 && optionsLen != defaultsLen {
		return nil, errors.New("provided defaults array must be of same length as options")
	}
	selected := s.Defaults
	if defaultsLen < 1 {
		selected = make([]bool, optionsLen)
	}

	cursor := 0
	infoLine := ""
	drawOptions(true, s, selected, cursor, infoLine)
	defer fmt.Print("\n\r") // prevent last option from being overwritten on return
	for {
		switch readInput() {
		case inputExit:
			return selected, ERR_SIGINT_RECEIVED
		case inputErr:
			return selected, ERR_READ_BYTE
		case inputSelect:
			if s.Multiple {
				selected[cursor] = !selected[cursor]
			} else {
				// In single-select mode, clear all selections and select the current option.
				for i := range selected {
					selected[i] = false
				}
				selected[cursor] = true
			}
		case inputUp:
			if cursor > 0 {
				cursor -= 1
			}
		case inputDown:
			if cursor < (optionsLen - 1) {
				cursor += 1
			}
		case inputConfirm:
			if s.Validate == nil {
				return selected, nil
			}
			if err := s.Validate(selected); err != nil {
				if s.OnlyOnce {
					return selected, ERR_INVALID_INPUT
				}
				infoLine = err.Error()
				break
			}
			return selected, nil
		case inputUnknown:
			continue
		}
		drawOptions(false, s, selected, cursor, infoLine)
		infoLine = ""
	}
}

func drawOptions(first bool, sel Select, selected []bool, current int, infoLine string) {
	if first {
		fmt.Print(strings.Repeat("\n", len(sel.Options)-1))
	}
	fmt.Printf("\033[%dA\r", len(sel.Options)-1)
	if infoLine == "" {
		fmt.Printf("\033[1A\r\033[2K%s\033[1B\r", sel.Prompt)
	} else {
		fmt.Printf("\033[1A\r\033[2K\033[31m%s\033[0m\033[1B\r", infoLine)
	}
	for i, opt := range sel.Options {
		if i != 0 {
			fmt.Print("\033[1B\r")
		}
		ptr := " "
		if i == current {
			ptr = ">"
		}
		mark := "[ ]"
		if selected[i] {
			mark = "[x]"
		}
		fmt.Printf("%s %s %s", ptr, mark, opt)
	}
}

type selectInput uint

const (
	inputErr selectInput = iota
	inputExit
	inputSelect
	inputConfirm
	inputUp
	inputDown
	inputUnknown
)

func readInput() selectInput {
	readState := 0
	var buf [1]byte
	for {
		n, err := os.Stdin.Read(buf[:])
		if err != nil || n == 0 {
			return inputErr
		}
		switch rune(buf[0]) {
		case '\003':
			return inputExit
		case '\004':
			return inputExit
		case ' ':
			return inputSelect
		case '\r', '\n':
			return inputConfirm
		case '\033':
			readState = 1
		case '[':
			if readState == 1 {
				readState = 2
			} else {
				return inputUnknown
			}
		case 'A':
			if readState == 2 {
				return inputUp
			}
		case 'B':
			if readState == 2 {
				return inputDown
			}
		default:
			return inputUnknown
		}
	}
}
