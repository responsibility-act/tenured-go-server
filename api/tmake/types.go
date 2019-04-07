package main

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"text/template"
)

var typePattern = regexp.MustCompile(`^(\w+) ([\[\]\w]+)?( (empty|zone))?$`)

type FieldDef struct {
	Name   string
	Desc   string
	Type   string
	Option string
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

type TypeDef struct {
	Name   string
	Desc   string
	Fields []FieldDef
}

func (this *TypeDef) ZoneField() *FieldDef {
	for _, v := range this.Fields {
		if v.Option == "zone" {
			return &v
		}
	}
	return nil
}

type TypesDef struct {
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
			Name: fieldDef[1], Desc: desc, Type: fieldDef[2], Option: fieldDef[4],
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
	{{.Name}} {{.Type}} !json:"{{.JsonName}}{{.OmitEmpty}}"!
	{{end}}
}
{{end}}
`))

	_ = t.Execute(b, this)
	return bytes.ReplaceAll(b.Bytes(), []byte{'!'}, []byte{'`'})
}

func NewTypes() *TypesDef {
	return &TypesDef{
		Types: map[string]TypeDef{},
	}
}
