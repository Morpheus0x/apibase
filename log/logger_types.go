package log

import (
	"strconv"

	"github.com/labstack/gommon/color"
)

type Armoring func(str string) string

func StrArmored(str string) string {
	return "[" + str + "]"
}

func StrNoop(str string) string {
	return str
}

type Level uint

const (
	LevelDevel Level = iota
	LevelDebug
	LevelInfo
	LevelNotice
	LevelWarning
	LevelError
	LevelCritical
	// LevelCritical is the highest level, do not add after this
)

const _Level_name = "DEVLDBUGINFONOTEWARNERORCRIT"

var col = color.New()

func (i Level) String() string {
	if i > LevelCritical {
		return "Level(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Level_name[i*4 : i*4+4]
}

func (i Level) StringColored(armoring Armoring) string {
	var levelColored string
	switch i {
	case LevelCritical:
		levelColored = col.RedBg(armoring(i.String()))
	case LevelError:
		levelColored = col.Red(armoring(i.String()))
	case LevelWarning:
		levelColored = col.Yellow(armoring(i.String()))
	case LevelNotice:
		levelColored = col.Green(armoring(i.String()))
	case LevelInfo:
		levelColored = col.Cyan(armoring(i.String()))
	case LevelDebug:
		levelColored = col.Grey(armoring(i.String()))
	case LevelDevel:
		levelColored = col.Magenta(armoring(i.String()))
	default:
		levelColored = armoring(i.String())
	}
	return levelColored
}
