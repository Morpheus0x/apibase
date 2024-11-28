package log

import "strconv"

type Level uint

const (
	LevelDebug Level = iota
	LevelInfo
	LevelNotice
	LevelWarning
	LevelError
	LevelCritical
)

const _Level_name = "DEBUGINFONOTICEWARNINGERRORCRITICAL"

var _Level_index = [...]uint8{0, 5, 9, 15, 22, 27, 35}

func (i Level) String() string {
	if i >= Level(len(_Level_index)-1) {
		return "Level(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Level_name[_Level_index[i]:_Level_index[i+1]]
}
