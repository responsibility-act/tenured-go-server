package main

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"text/template"
)

var typePattern = regexp.MustCompile(`^(\w+) ([\[\]\w]+)?( (empty))?$`)

type FieldDef struct {
	Name   string
	Desc   string
	Type   string
	Option string
	Enums  *Enums
}

func (f *FieldDef) JsonName() string {
	return lowerName(f.Name)
}

func (f *FieldDef) OmitEmpty() string {
	if f.Option == "empty" {
		return ",omitempty"
	}
	return ""
}

func (f *FieldDef) ShowType() string {
	if isBase(f.Type) {
		return f.Type
	} else if isArray(f.Type) {
		return "[]*" + f.Type[2:]
	} else if f.Enums.HasEnum(f.Type) {
		return f.Type
	} else {
		return "*" + f.Type
	}
}

type TypeDef struct {
	Name   string
	Desc   string
	Fields []FieldDef

	Enums *Enums
}

type TypesDef struct {
	Enums *Enums
	Types map[string]TypeDef
}

func (this *TypesDef) Add(addLines []string, info *TCDInfo) error {
	desc, lines := comment(addLines)
	if !strings.HasPrefix(lines[0], "type") {
		return NotMatch
	}
	typeName := lines[0][5 : len(lines[0])-2]
	typedef := TypeDef{
		Name:   typeName,
		Desc:   desc,
		Fields: make([]FieldDef, 0),
		Enums:  this.Enums,
	}
	lines = body(lines)

	for ; len(lines) > 0; lines = lines[1:] {
		desc, lines = comment(lines)
		line := lines[0]
		if !typePattern.MatchString(line) {
			return errors.New("error at : " + lines[0])
		}
		fieldDef := typePattern.FindStringSubmatch(line)
		typedef.Fields = append(typedef.Fields, FieldDef{
			Enums: this.Enums,
			Name:  fieldDef[1], Desc: desc, Type: fieldDef[2], Option: fieldDef[4],
		})
	}
	this.Types[typeName] = typedef
	return nil
}

func (this *TypesDef) InterOuter(info *TCDInfo) []byte {
	b := new(bytes.Buffer)
	t := template.Must(template.New("letter").Parse(`
{{range .Types}}
{{.Desc}}
type {{.Name}} struct {
	{{range .Fields}}	
	{{if ne .Desc ""}}{{.Desc}}{{end}}
	{{.Name}} {{.ShowType}} !json:"{{.JsonName}}{{.OmitEmpty}}"!
	{{end}}
}
{{end}}
`))
	_ = t.Execute(b, this)
	return bytes.ReplaceAll(b.Bytes(), []byte{'!'}, []byte{'`'})
}

func NewTypes(enums *Enums) *TypesDef {
	return &TypesDef{
		Types: map[string]TypeDef{},
		Enums: enums,
	}
}
