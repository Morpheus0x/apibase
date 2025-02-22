package ansiesc

import (
	"fmt"
	"strconv"
	"strings"
)

// Source: https://www.perplexity.ai/search/create-a-golang-module-called-LyMrNTFATYySQ_mPM7e0Vg
// TODO: implement wysiwyg style string building where clearing styles is done individually

type AnsiColor uint

const (
	Black AnsiColor = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	Gray
	BrightRed
	BrightGreen
	BrightYellow
	BrightBlue
	BrightMagenta
	BrightCyan
	BrightWhite
)

type AnsiEsc struct {
	builder strings.Builder
}

func New() *AnsiEsc {
	return &AnsiEsc{}
}

func (a *AnsiEsc) CurUp(n int) *AnsiEsc {
	a.builder.WriteString("\x1b[" + strconv.Itoa(n) + "A")
	return a
}

func (a *AnsiEsc) CurRight(n int) *AnsiEsc {
	a.builder.WriteString("\x1b[" + strconv.Itoa(n) + "C")
	return a
}

func (a *AnsiEsc) CurDown(n int) *AnsiEsc {
	a.builder.WriteString("\x1b[" + strconv.Itoa(n) + "B")
	return a
}

func (a *AnsiEsc) CurLeft(n int) *AnsiEsc {
	a.builder.WriteString("\x1b[" + strconv.Itoa(n) + "D")
	return a
}

func (a *AnsiEsc) CurHome() *AnsiEsc {
	a.builder.WriteString("\r")
	return a
}

func (a *AnsiEsc) CurEnd() *AnsiEsc {
	a.builder.WriteString("\x1b[999C")
	return a
}

func (a *AnsiEsc) ColSet(c AnsiColor) *AnsiEsc {
	a.builder.WriteString("\x1b[" + strconv.Itoa(int(c)) + "m")
	return a
}

func (a *AnsiEsc) ColClr() *AnsiEsc {
	a.builder.WriteString("\x1b[0m") // TODO: this clears all styles, fix
	return a
}

func (a *AnsiEsc) BgSet(c AnsiColor) *AnsiEsc {
	a.builder.WriteString(fmt.Sprintf("\x1b[48;5;%dm", c))
	return a
}

func (a *AnsiEsc) BgClr() *AnsiEsc {
	a.builder.WriteString("\x1b[0m") // TODO: this clears all styles, fix
	return a
}

func (a *AnsiEsc) LineClr() *AnsiEsc {
	a.builder.WriteString("\r\x1b[2K")
	return a
}

func (a *AnsiEsc) ScrnClr() *AnsiEsc {
	a.builder.WriteString("\x1b[2J")
	return a
}

func (a *AnsiEsc) Text(text string) *AnsiEsc {
	a.builder.WriteString(text)
	return a
}

func (a *AnsiEsc) String() string {
	return a.builder.String()
}
