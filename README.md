# lgr
Another implementation of the logger with levels.

## Install

```sh
go get github.com/va-slyusarev/lgr...
```

## Levels
In total 4 levels of logging are used:
```go
// Levels.
const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
)
```

Each level has a priority, and if the current level is higher in priority than the priority of the message, it will not be printed.

## Format
The format of the output messages can be configured using the template `text/template`.
There are 4 predefined templates:

```go
const (
	XSmallTpl = `{{ if .Prefix }}{{ printf "%s: " .Prefix }}{{ end }}{{ .Message }}`
	SmallTpl  = `{{ printf "[%-5s] " .Level }}{{ if .Prefix }}{{ printf "%s: " .Prefix }}{{ end }}{{ .Message }}`
	MediumTpl = `{{.TS.Format "2006/01/02 15:04:05" }} {{ printf "[%-5s] " .Level }}{{ if .Prefix }}{{ printf "%s: " .Prefix }}{{ end }}{{ .Message }}`
	LargeTpl  = `{{.TS.Format "2006/01/02 15:04:05.000" }} {{ printf "[%-5s] " .Level }}{{ printf "{%s} " .Caller }}{{ if .Prefix }}{{ printf "%s: " .Prefix }}{{ end }}{{ .Message }}`
)
```

Standard templates have aliases for use as flag parameters:
 - **XSmallTpl** - `XSmallTpl` or `xs`
 - **SmallTpl** - `SmallTpl` or `sm`
 - **MediumTpl** - `MediumTpl` or `md`
 - **LargeTpl** - `LargeTpl` or `lg`
 

## Use case

```go
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

	lgr.Debug(*message)
	lgr.Info(*message)
	lgr.Warn(*message)
	lgr.Error(*message)
}
```

Run

```sh
 go run main.go -l=WARN -p=ex -t=lg
```

You'll see something like this

```sh
2006/01/02 15:04:05.000 [WARN ] {_example/test.go:19} ex: hello lgr!
2006/01/02 15:04:05.000 [ERROR] {_example/test.go:20} ex: hello lgr!
```


See `_example`