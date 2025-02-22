package cli

import "errors"

var (
	ERR_READLINE        = errors.New("unable to create line reader")
	ERR_SIGINT_RECEIVED = errors.New("sigint received")
	ERR_INVALID_INPUT   = errors.New("no valid input specified")
	ERR_MIN_INPUT_LEN   = errors.New("longer input required")
	ERR_READ_BYTE       = errors.New("unabel to read input byte")
)
