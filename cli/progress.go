package cli

import (
	"fmt"
	"time"
)

type Op uint

const (
	Log Op = iota
	Update
)

var (
	asciiSpinner   = []string{"-", "\\", "|", "/"}
	unicodeSpinner = []string{"⠽", "⠾", "⠷", "⠯", "⠟", "⠻"}
)

type Progress struct {
	// duration between loading char, default 100ms
	LoadingSpeed time.Duration
	// if set, show ASCII only loading symbol instead of unicode special char
	FallbackAscii bool
}

type TaskOperation struct {
	Operation Op // "log" for log message, "update" for updating task text
	Text      string
}

// startLoadingTask starts a non-blocking loading task.
// It returns a channel to send task operations and a function to stop the loading.
func (p Progress) Start(initialTaskText string) (chan<- TaskOperation, func()) {
	speed := p.LoadingSpeed
	if p.LoadingSpeed == 0 {
		speed = 100 * time.Millisecond
	}
	taskChan := make(chan TaskOperation)
	doneChan := make(chan bool)

	taskText := initialTaskText // Initial task text

	// Non-blocking loading display
	go func() {
		for {
			select {
			case <-doneChan:
				return
			case op := <-taskChan:
				if op.Operation == Log {
					fmt.Print("\r\033[K")
					fmt.Println(op.Text)
					if p.FallbackAscii {
						fmt.Printf("\x1b[32m%s\033[m %s", asciiSpinner[0], taskText)
					} else {
						fmt.Printf("\x1b[32m%s\033[m %s", unicodeSpinner[0], taskText)
					}
				} else if op.Operation == Update {
					taskText = op.Text
				}
			default:
				length := len(unicodeSpinner)
				if p.FallbackAscii {
					length = len(asciiSpinner)
				}
				for i := 0; i < length; i++ {
					if p.FallbackAscii {
						fmt.Printf("\r\x1b[32m%s\033[m %s", asciiSpinner[i], taskText)
					} else {
						fmt.Printf("\ry\x1b[32m%s\033[m %s", unicodeSpinner[i], taskText)
					}
					time.Sleep(speed)
				}
			}
		}
	}()

	// Function to stop the loading
	stopFunc := func() {
		doneChan <- true
		// Clear the loading line
		fmt.Print("\r\033[K")
	}

	return taskChan, stopFunc
}
