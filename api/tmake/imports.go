package main

import (
	"bytes"
	"strings"
)

type Imports struct {
	imports map[string]string
}

func NewImport() *Imports {
	return &Imports{
		imports: map[string]string{
			"github.com/ihaiker/tenured-go-server/commons/protocol": "protocol",
		},
	}
}

func (this *Imports) Is(head string) bool {
	return "imports {" == head
}

func (this *Imports) Add(lines []string, info *TCDInfo) error {
	if !this.Is(lines[0]) {
		return NotMatch
	}
	for i := 1; i < len(lines)-1; i++ {
		spt := strings.SplitN(lines[i], " ", 2)
		alias := ""
		importUrl := ""
		if len(spt) == 1 {
			importUrl = spt[0]
		} else {
			alias = spt[0]
			importUrl = spt[1]
		}
		(this.imports)[importUrl] = alias
	}
	return nil
}

func (this *Imports) InterOuter(info *TCDInfo) []byte {
	b := new(bytes.Buffer)
	b.WriteString("import (\n")
	for k, v := range this.imports {
		b.WriteRune('\t')
		if v != "" {
			b.WriteString(v)
			b.WriteString(" ")
		}
		b.WriteString("\"")
		b.WriteString(k)
		b.WriteString("\"")
		b.WriteRune('\n')
	}
	b.WriteString(")\n\n")
	return b.Bytes()
}
