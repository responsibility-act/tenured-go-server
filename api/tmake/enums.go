package main

import (
	"bytes"
	"fmt"
	"regexp"
)

var enumPattern = regexp.MustCompile("^(enum )(\\w+)(\\(([\\w]+)\\))? {$")
var eunmValuePattern = regexp.MustCompile("^(\\w+)( = (\\w+))?( .*)?$")

type EnumDef struct {
	Name  string
	Desc  string
	Value [][]string
	Type  string
}

type Enums struct {
	enums map[string]*EnumDef
}

func (e *Enums) HasEnum(name string) bool {
	_, has := e.enums[name]
	return has
}

func (this Enums) Add(addLines []string, info *TCDInfo) error {
	desc, lines := comment(addLines)

	if !enumPattern.MatchString(lines[0]) {
		return NotMatch
	}

	enumDef := &EnumDef{
		Desc: desc,
	}

	heads := enumPattern.FindStringSubmatch(lines[0])
	enumDef.Name = heads[2]
	enumDef.Type = heads[4]
	enumDef.Value = make([][]string, 0)
	if enumDef.Type == "" {
		enumDef.Type = "string"
	}
	lines = body(lines)

	for ; len(lines) > 0; lines = lines[1:] {
		desc, lines = comment(lines)
		evs := eunmValuePattern.FindStringSubmatch(lines[0])
		enumKey := evs[1]
		enumValue := evs[3]
		if enumValue == "" {
			enumValue = enumKey
		}
		enumDesc := evs[4]
		if enumDesc == "" {
			enumDesc = desc
		} else {
			enumDesc = enumDesc[1:]
		}
		enumDef.Value = append(enumDef.Value, []string{enumKey, enumValue, enumDesc})
	}

	this.enums[heads[2]] = enumDef
	return nil
}

func (this Enums) InterOuter(info *TCDInfo) []byte {
	b := new(bytes.Buffer)
	if len(this.enums) == 0 {
		return b.Bytes()
	}
	for _, e := range this.enums {
		b.WriteString("\n")
		b.WriteString(e.Desc)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("type %s %s\n", e.Name, e.Type))
		b.WriteString("const (\n")
		for _, enumValue := range e.Value {
			if enumValue[2] != "" {
				b.WriteString(fmt.Sprintf("\t%s\n", enumValue[2]))
			}
			switch e.Type {
			case "string":
				b.WriteString(fmt.Sprintf("	%s%s = \"%s\" \n", e.Name, enumValue[0], enumValue[1]))
			default:
				b.WriteString(fmt.Sprintf("	%s%s = %s(%s) \n", e.Name, enumValue[0], e.Type, enumValue[1]))
			}
		}
		b.WriteString(")\n")
	}
	return b.Bytes()
}

func NewEunms() *Enums {
	return &Enums{
		enums: map[string]*EnumDef{},
	}
}
