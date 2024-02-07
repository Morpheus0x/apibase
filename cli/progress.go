package cli

import (
	"fmt"
	"time"
)

type TaskOperation struct {
	Operation string // "log" for log message, "update" for updating task text
	Text      string
}

// startLoadingTask starts a non-blocking loading task.
// It returns a channel to send task operations and a function to stop the loading.
func StartProgress(initialTaskText string) (chan<- TaskOperation, func()) {
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
				if op.Operation == "log" {
					// Move up one line and clear line for log message
					// fmt.Print("\033[1A\033[K")
					fmt.Print("\r\033[K")
					// fmt.Printf("%s\n\n", op.Text)
					fmt.Println(op.Text)
					// Display the loading icon and current task text again
					// fmt.Printf("\r- %s", taskText)
					fmt.Printf("- %s", taskText)
				} else if op.Operation == "update" {
					// Update task text
					taskText = op.Text
				}
			default:
				// Display loading icon with current task text
				fmt.Printf("\r- %s", taskText)
				time.Sleep(100 * time.Millisecond)
				fmt.Printf("\r\\ %s", taskText)
				time.Sleep(100 * time.Millisecond)
				fmt.Printf("\r| %s", taskText)
				time.Sleep(100 * time.Millisecond)
				fmt.Printf("\r/ %s", taskText)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Function to stop the loading
	stopFunc := func() {
		doneChan <- true
		// Clear the loading line
		// fmt.Print("\033[1A\033[K")
		// fmt.Print("\n")
		fmt.Print("\r\033[K")
	}

	return taskChan, stopFunc
}
