package helper

import (
	"fmt"
	"time"

	"github.com/xhit/go-str2duration/v2"
)

func StringToDuration(toParse string) (time.Duration, error) {
	if toParse == "" {
		return 0, fmt.Errorf("unable to parse empty string to duration")
	}
	duration, err := str2duration.ParseDuration(toParse)
	if err != nil {
		return 0, fmt.Errorf("unable to parse string to duration: %v", err)
	}
	return duration, nil
}
