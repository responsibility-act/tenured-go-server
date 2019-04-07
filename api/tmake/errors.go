package main

import (
	"bytes"
	"errors"
	"regexp"
	"text/template"
)

var errorPattern = regexp.MustCompile(`^(\w+)\((\w+),(\S+)\)$`)

type ErrorDef struct {
	Code    string
	Message string
}

type ErrorsDef struct {
	Errors map[string]ErrorDef
}

func (this *ErrorsDef) Add(addlines []string, info *TCDInfo) error {
	_, lines := comment(addlines)
	if "errors {" != lines[0] {
		return NotMatch
	}
	lines = body(lines)
	for _, line := range lines {
		if !errorPattern.MatchString(line) {
			return errors.New("error at: " + line)
		}
		errorDef := errorPattern.FindStringSubmatch(line)
		this.Errors[errorDef[1]] = ErrorDef{
			Code: errorDef[2], Message: errorDef[3],
		}
	}
	return nil
}

func (this *ErrorsDef) InterOuter(info *TCDInfo) []byte {
	b := new(bytes.Buffer)
	if len(this.Errors) > 0 {
		t := template.Must(template.New("letter").Parse(`
var ({{range $k,$v := .Errors}}
	Err{{$k}} =  protocol.NewError("{{$v.Code}}","{{$v.Message}}"){{end}}
)`))
		_ = t.Execute(b, this)
	}
	return b.Bytes()
}

func NewErrors() *ErrorsDef {
	return &ErrorsDef{
		Errors: map[string]ErrorDef{},
	}
}
