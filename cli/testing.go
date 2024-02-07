package cli

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/chzyer/readline"
	"golang.org/x/term"
)

func testing() (bool, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return false, err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	var conn io.ReadWriter

	term := term.NewTerminal(conn, "")
	line, err := term.ReadLine()
	if err != nil {
		return false, err
	}
	fmt.Printf("Out: %s", line)
	// func (t *Terminal) ReadLine() (line string, err error)

	l, err := readline.NewEx(&readline.Config{
		Prompt:      "\033[31m»\033[0m ",
		HistoryFile: "/tmp/readline.tmp",
		// AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold: true,
		// FuncFilterInputRune: filterInput,
	})
	if err != nil {
		return false, err
	}
	defer l.Close()
	l.CaptureExitSignal()

	setPasswordCfg := l.GenPasswordConfig()
	setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		l.SetPrompt(fmt.Sprintf("Enter password(%v): ", len(line)))
		l.Refresh()
		return nil, 0, false
	})

	log.SetOutput(l.Stderr())
	// for {
	// 	line, err := l.Readline()
	// }
	return false, nil
}
