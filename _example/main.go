// Copyright © 2019 Valentin Slyusarev <va.slyusarev@gmail.com>

package main

import (
	"flag"

	"github.com/va-slyusarev/lgr"
)

var message = flag.String("msg", "hello lgr!", "Logger message.")

func main() {
	flag.Var(lgr.Std.Level, "l", "Logger level.")
	flag.Var(lgr.Std.Prefix, "p", "Logger prefix.")
	flag.Var(lgr.Std.Template, "t", "Logger template.")
	flag.Parse()

	// template from flag value
	out()

	lgr.SetTpl(lgr.SmallTpl)
	out()

	lgr.SetTpl(lgr.MediumTpl)
	out()

	lgr.SetTpl(lgr.LargeTpl)
	out()
}

func out() {
	lgr.Debug(*message)
	lgr.Info(*message)
	lgr.Warn(*message)
	lgr.Error(*message)
}
