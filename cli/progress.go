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
	asciiSpinner      = []string{"|", "/", "-", "\\"} // #
	unicodeSpinner4x1 = []string{"⠂", "⠐", "⠠", "⠄"}  // ⠶
	unicodeSpinner4x2 = []string{"⠆", "⠒", "⠰", "⠤"}
	unicodeSpinner4x3 = []string{"⠦", "⠖", "⠲", "⠴"}
	unicodeSpinner6x1 = []string{"⠁", "⠈", "⠐", "⠠", "⠄", "⠂"} // ⠿
	unicodeSpinner6x2 = []string{"⠃", "⠉", "⠘", "⠰", "⠤", "⠆"}
	unicodeSpinner6x3 = []string{"⠇", "⠋", "⠙", "⠸", "⠴", "⠦"}
	unicodeSpinner6x4 = []string{"⠧", "⠏", "⠛", "⠹", "⠼", "⠶"}
	unicodeSpinner6x5 = []string{"⠷", "⠯", "⠟", "⠻", "⠽", "⠾"}
	unicodeSpinner8x1 = []string{"⠁", "⠈", "⠐", "⠠", "⢀", "⡀", "⠄", "⠂"} // ⣿
	unicodeSpinner8x2 = []string{"⠃", "⠉", "⠘", "⠰", "⢠", "⣀", "⡄", "⠆"}
	unicodeSpinner8x3 = []string{"⠇", "⠋", "⠙", "⠸", "⢰", "⣠", "⣄", "⡆"}
	unicodeSpinner8x4 = []string{"⡇", "⠏", "⠛", "⠹", "⢸", "⣰", "⣤", "⣆"}
	unicodeSpinner8x5 = []string{"⣇", "⡏", "⠟", "⠻", "⢹", "⣸", "⣴", "⣦"}
	unicodeSpinner8x6 = []string{"⣧", "⣏", "⡟", "⠿", "⢻", "⣹", "⣼", "⣶"}
	unicodeSpinner8x7 = []string{"⣷", "⣯", "⣟", "⡿", "⢿", "⣻", "⣽", "⣾"}
)

type Progress struct {
	// duration between loading char, default 100ms
	LoadingSpeed time.Duration
	// if set, show ASCII only loading symbol instead of unicode special char
	FallbackAscii bool
}

type TaskOperation struct {
	Log        bool
	LogText    string
	Update     bool
	UpdateText string
}

func (p Progress) Spinners() {
	fmt.Printf("%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v\n",
		asciiSpinner,
		unicodeSpinner4x1, unicodeSpinner4x2, unicodeSpinner4x3,
		unicodeSpinner6x1, unicodeSpinner6x2, unicodeSpinner6x3, unicodeSpinner6x4, unicodeSpinner6x5,
		unicodeSpinner8x1, unicodeSpinner8x2, unicodeSpinner8x3, unicodeSpinner8x4, unicodeSpinner8x5, unicodeSpinner8x6, unicodeSpinner8x7,
	)
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

	// TODO: initialize length here and use in for loop inside goroutine to improve log responsiveness
	// on every for loop iteration, check for message received or update progress text
	// Add newline at end of done to prevent override of last progress line
	// Also clear progress line before every update to prevent pollution
	// Non-blocking loading display
	spinnerCount := 0
	spinnerLenght := len(unicodeSpinner6x3)
	if p.FallbackAscii {
		spinnerLenght = len(asciiSpinner)
	}
	go func() {
		for {
			if spinnerCount >= spinnerLenght {
				spinnerCount = 0
			}
			select {
			case op := <-taskChan:
				if op.Log {
					fmt.Print("\r\033[K") // chariage return & clear line
					fmt.Printf("%s\n", op.LogText)
				}
				if op.Update {
					taskText = op.UpdateText
				}
				if p.FallbackAscii {
					fmt.Printf("\x1b[32m%s\033[m %s", asciiSpinner[spinnerCount], taskText)
				} else {
					fmt.Printf("\x1b[32m%s\033[m %s", unicodeSpinner6x3[spinnerCount], taskText)
				}
			case <-doneChan:
				return
			default:
				if p.FallbackAscii {
					fmt.Printf("\r\033[K\x1b[32m%s\033[m %s", asciiSpinner[spinnerCount], taskText)
				} else {
					fmt.Printf("\r\033[K\x1b[32m%s\033[m %s", unicodeSpinner6x3[spinnerCount], taskText)
				}
			}
			time.Sleep(speed)
			spinnerCount++
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
