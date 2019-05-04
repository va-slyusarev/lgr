// Copyright © 2019 Valentin Slyusarev <va.slyusarev@gmail.com>

package lgr

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"
)

// Levels.
const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
)

// Levels and priority.
var levels = map[string]int{DEBUG: 0, INFO: 1, WARN: 2, ERROR: 3}

// lgr. Main type.
type lgr struct {
	mu       sync.Mutex
	out      io.Writer // Where to output. By default, os.Stderr.
	Level    *level    // Current level. Not print message which priority is less.
	Prefix   *prefix   // Common info added each messages.
	Template *tpl      // Current template and text/template instance for generate message.
	skip     int       // For caller info std and custom logger.
	now      nowFn
}

// Templates for print.
const (
	XSmallTpl = `{{ if .Prefix }}{{ printf "%s: " .Prefix }}{{ end }}{{ .Message }}`
	SmallTpl  = `{{ printf "[%-5s] " .Level }}{{ if .Prefix }}{{ printf "%s: " .Prefix }}{{ end }}{{ .Message }}`
	MediumTpl = `{{.TS.Format "2006/01/02 15:04:05" }} {{ printf "[%-5s] " .Level }}{{ if .Prefix }}{{ printf "%s: " .Prefix }}{{ end }}{{ .Message }}`
	LargeTpl  = `{{.TS.Format "2006/01/02 15:04:05.000" }} {{ printf "[%-5s] " .Level }}{{ printf "{%s} " .Caller }}{{ if .Prefix }}{{ printf "%s: " .Prefix }}{{ end }}{{ .Message }}`
)

// Message. Type containing information about the logged message.
// This type use in template.
type message struct {
	TS      time.Time
	Level   string
	Caller  string
	Prefix  string
	Message string
}

// level. Type implements flag value interface.
type level struct {
	value string
}

// String.
func (l *level) String() string { return l.value }

// Set.
func (l *level) Set(value string) error {
	for level := range levels {
		if strings.EqualFold(value, level) {
			l.value = level
			return nil
		}
	}
	return fmt.Errorf("lgr: set level: value %s is unknown", value)
}

// allowed. The current priority level allows the output messages.
func (l *level) allowed(value string) bool {
	v := level{}

	if err := v.Set(value); err == nil {
		return levels[v.value] >= levels[l.value]
	}
	return false
}

// prefix. Type implements flag value interface.
type prefix struct {
	value string
}

// String.
func (p *prefix) String() string { return p.value }

// Set.
func (p *prefix) Set(value string) error {
	p.value = value
	return nil
}

// tpl. Type implements flag value interface.
type tpl struct {
	value  string
	engine *template.Template
}

// String.
func (t *tpl) String() string { return t.value }

// Set. Set template and init text/template engine.
func (t *tpl) Set(value string) error {
	// Handling standard template names
	switch value {
	case "XSmallTpl", "xs":
		value = XSmallTpl
	case "SmallTpl", "sm":
		value = SmallTpl
	case "MediumTpl", "md":
		value = MediumTpl
	case "LargeTpl", "lg":
		value = LargeTpl
	}

	engine, err := template.New("lgr").Parse(value)
	if err != nil {
		return fmt.Errorf("afs: set template: %v", err)
	}
	t.value = value
	t.engine = engine
	return nil
}

type nowFn func() time.Time // for testing.

// New lgr.
func New() *lgr {
	res := &lgr{
		Level:    &level{},
		Prefix:   &prefix{},
		Template: &tpl{},
		skip:     2,
		now:      time.Now,
	}
	return res.SetOut(os.Stderr).SetTpl(LargeTpl).SetLevel(DEBUG)
}

// SetOut. Set current out.
func (l *lgr) SetOut(w io.Writer) *lgr {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
	return l
}

// SetPrefix. Set current prefix.
func (l *lgr) SetPrefix(prefix string) *lgr {
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = l.Prefix.Set(prefix)
	return l
}

// SetLevel. Set current level.
func (l *lgr) SetLevel(level string) *lgr {
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = l.Level.Set(level)
	return l
}

// SetTpl. Set current template.
func (l *lgr) SetTpl(tpl string) *lgr {
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = l.Template.Set(tpl)
	return l
}

// Output send message in logger out.
func (l *lgr) Output(level string, s string) error {

	ci := fmt.Sprintf("%s:%s", "???", "???")
	if _, file, line, ok := runtime.Caller(l.skip); ok {
		dir, file := path.Split(file)
		dir = path.Base(dir)
		ci = fmt.Sprintf("%s/%s:%d", dir, file, line)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.Level.allowed(level) {
		return fmt.Errorf("lgr: level [%-5s] not allowed, because current level is [%-5s]", level, l.Level)
	}

	msg := message{
		TS:      l.now(),
		Level:   level,
		Caller:  ci,
		Prefix:  l.Prefix.value,
		Message: strings.TrimSuffix(s, "\n"),
	}
	buf := bytes.NewBuffer([]byte{})
	err := l.Template.engine.Execute(buf, msg)

	if err != nil {
		return fmt.Errorf("lgr: %v", err)
	}

	buf.WriteString("\n")
	_, err = buf.WriteTo(l.out)

	return err
}

// Debug. Output message with DEBUG level.
func (l *lgr) Debug(format string, v ...interface{}) { _ = l.Output(DEBUG, fmt.Sprintf(format, v...)) }

// Info. Output message with INFO level.
func (l *lgr) Info(format string, v ...interface{}) { _ = l.Output(INFO, fmt.Sprintf(format, v...)) }

// Warn. Output message with WARN level.
func (l *lgr) Warn(format string, v ...interface{}) { _ = l.Output(WARN, fmt.Sprintf(format, v...)) }

// Error. Output message with ERROR level.
func (l *lgr) Error(format string, v ...interface{}) { _ = l.Output(ERROR, fmt.Sprintf(format, v...)) }

// Std lgr.
var Std = newStd()

// newStd. Set default values.
func newStd() *lgr {
	l := New()
	l.skip = 3
	return l.SetTpl(SmallTpl).SetLevel(INFO)
}

// SetOut. Set current out for Std logger.
func SetOut(w io.Writer) { _ = Std.SetOut(w) }

// SetPrefix. Set current prefix for Std logger.
func SetPrefix(prefix string) { _ = Std.SetPrefix(prefix) }

// SetLevel. Set current level for Std logger.
func SetLevel(level string) { _ = Std.SetLevel(level) }

// SetTpl. Set current template for Std logger.
func SetTpl(tpl string) { _ = Std.SetTpl(tpl) }

// Output send message in Std logger out.
func Output(level string, s string) error { return Std.Output(level, s) }

// Debug. Output message with DEBUG level in Std logger.
func Debug(format string, v ...interface{}) { Std.Debug(format, v...) }

// Info. Output message with INFO level in Std logger.
func Info(format string, v ...interface{}) { Std.Info(format, v...) }

// Warn. Output message with WARN level in Std logger.
func Warn(format string, v ...interface{}) { Std.Warn(format, v...) }

// Error. Output message with ERROR level in Std logger.
func Error(format string, v ...interface{}) { Std.Error(format, v...) }
