// Copyright © 2019 Valentin Slyusarev <va.slyusarev@gmail.com>

package lgr

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

var nowTest = func() time.Time { return time.Date(2019, 3, 2, 00, 20, 15, 0, time.Local) }
var hello = fmt.Sprintf("Hello, %s!", "World")

func Example1() {
	out := bytes.NewBuffer([]byte{})
	SetOut(out)
	Warn(fmt.Sprintf("%s Current level is [%s].", hello, Std.Level))
	fmt.Println(out)

	// Output:
	// [WARN ] Hello, World! Current level is [INFO].
}

func Example2() {
	out := bytes.NewBuffer([]byte{})
	l := New().SetOut(out)
	l.now = nowTest
	l.Warn(fmt.Sprintf("%s Current level is [%s].", hello, l.Level))
	fmt.Println(out)

	// Output:
	// 2019/03/02 00:20:15.000 [WARN ] {lgr/lgr_test.go:29} Hello, World! Current level is [DEBUG].
}

func Example3() {
	out := bytes.NewBuffer([]byte{})
	l := New().SetOut(out).SetLevel(INFO)
	err := l.Output(DEBUG, hello)
	fmt.Println(err)

	// Output:
	// lgr: level [DEBUG] not allowed, because current level is [INFO ]
}

func TestLogger(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	l := New().SetOut(out)
	l.now = nowTest

	t.Run("Set level case insensitive", func(t *testing.T) {
		if err := l.Level.Set("DeBuG"); err != nil || l.Level.String() != DEBUG {
			t.Errorf("Level = %s", l.Level)
		}
	})

	t.Run("Set unknown level", func(t *testing.T) {
		if err := l.Level.Set("unknown"); err == nil {
			t.Errorf("Level = %s", l.Level)
		}
	})

	t.Run("Output unknown level", func(t *testing.T) {
		if err := l.Output("unknown", hello); err == nil {
			t.Error("Level unknown is output")
		}
	})

	t.Run("Set prefix", func(t *testing.T) {
		prefix := "testing"
		_ = l.SetPrefix(prefix)

		if l.Prefix.String() != prefix {
			t.Errorf("Prefix = %s", l.Prefix)
		}
	})

	t.Run("Set Template", func(t *testing.T) {
		tpl := "{{ .Message }}"
		if err := l.Template.Set(tpl); err != nil {
			t.Errorf("set template error: %v", err)
		}

		if l.Template.String() != tpl {
			t.Errorf("set another template: %s", l.Template)
		}

		l.Error(hello)
		if out.String() != fmt.Sprintf("%s\n", hello) {
			t.Errorf("template not apply: %s", out.String())
		}
	})

	t.Run("Set broken Template", func(t *testing.T) {
		tpl := "{{ .Message "
		if err := l.Template.Set(tpl); err == nil {
			t.Error("template not broken")
		}
	})

	t.Run("Set broken Template for message", func(t *testing.T) {
		tpl := "{{ .Message.BROKEN }}"
		if err := l.Template.Set(tpl); err != nil {
			t.Errorf("template not valid: %v", err)
		}
		out.Reset()
		if err := l.Output(ERROR, hello); err == nil {
			t.Errorf("execute: %s", out.String())
		}
	})

	t.Run("Set Template alias", func(t *testing.T) {

		// Extra Small
		if err := l.Template.Set("XSmallTpl"); err != nil || l.Template.value != XSmallTpl {
			t.Error("template not broken")
		}
		if err := l.Template.Set("xs"); err != nil || l.Template.value != XSmallTpl {
			t.Error("template not broken")
		}

		// Small
		if err := l.Template.Set("SmallTpl"); err != nil || l.Template.value != SmallTpl {
			t.Error("template not broken")
		}
		if err := l.Template.Set("sm"); err != nil || l.Template.value != SmallTpl {
			t.Error("template not broken")
		}

		// Medium
		if err := l.Template.Set("MediumTpl"); err != nil || l.Template.value != MediumTpl {
			t.Error("template not broken")
		}
		if err := l.Template.Set("md"); err != nil || l.Template.value != MediumTpl {
			t.Error("template not broken")
		}

		// Large
		if err := l.Template.Set("LargeTpl"); err != nil || l.Template.value != LargeTpl {
			t.Error("template not broken")
		}
		if err := l.Template.Set("lg"); err != nil || l.Template.value != LargeTpl {
			t.Error("template not broken")
		}
	})

	t.Run("Levels", func(t *testing.T) {
		_ = l.SetTpl(MediumTpl).SetPrefix("").SetLevel(DEBUG)
		out.Reset()
		l.Debug(hello)
		l.Info(hello)
		l.Warn(hello)
		l.Error(hello)
		if out.String() != `2019/03/02 00:20:15 [DEBUG] Hello, World!
2019/03/02 00:20:15 [INFO ] Hello, World!
2019/03/02 00:20:15 [WARN ] Hello, World!
2019/03/02 00:20:15 [ERROR] Hello, World!
` {
			t.Errorf("levels not work:\n%s", out.String())
		}
	})
}

func TestStdLogger(t *testing.T) {
	out, level, prefix, tpl := Std.out, Std.Level.String(), Std.Prefix.String(), Std.Template.String()
	defer func() {
		SetOut(out)
		SetLevel(level)
		SetPrefix(prefix)
		SetTpl(tpl)
	}()

	out1 := bytes.NewBuffer([]byte{})
	Std.now = nowTest
	SetOut(out1)
	SetLevel(DEBUG)
	SetPrefix("")
	SetTpl(MediumTpl)

	t.Run("Output", func(t *testing.T) {
		if err := Output(DEBUG, hello); err != nil {
			t.Errorf("Std output: %v", err)
		}
	})

	t.Run("Levels", func(t *testing.T) {
		out1.Reset()
		Debug(hello)
		Info(hello)
		Warn(hello)
		Error(hello)
		if out1.String() != `2019/03/02 00:20:15 [DEBUG] Hello, World!
2019/03/02 00:20:15 [INFO ] Hello, World!
2019/03/02 00:20:15 [WARN ] Hello, World!
2019/03/02 00:20:15 [ERROR] Hello, World!
` {
			t.Errorf("levels not work:\n%s", out1.String())
		}
	})
}

func BenchmarkStd(b *testing.B) {
	out := bytes.NewBuffer([]byte{})
	SetOut(out)
	for n := 0; n < b.N; n++ {
		Info(hello)
	}
}

func BenchmarkCustom(b *testing.B) {
	out := bytes.NewBuffer([]byte{})
	l := New().SetOut(out).SetPrefix("testing").SetLevel(DEBUG).SetTpl(LargeTpl)
	for n := 0; n < b.N; n++ {
		l.Info(hello)
	}
}
