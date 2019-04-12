package main

import (
	"bytes"
	"strings"
)

type Imports struct {
	Imports map[string]string

	InterfacePackage map[string]string
	ClientPackage    map[string]string
	InvokePackage    map[string]string
}

func NewImport(tcd *TCDInfo) *Imports {
	return &Imports{
		Imports: map[string]string{
			TenuredHome + "/commons/protocol": "",
		},
		InterfacePackage: map[string]string{},
		ClientPackage: map[string]string{
			tcd.ApiPackageUrl:                 "",
			TenuredHome + "/commons/registry": "",
			TenuredHome + "/commons":          "",
			"time":                            "",
		},
		InvokePackage: map[string]string{
			tcd.ApiPackageUrl:                  "",
			TenuredHome + "/commons/executors": "",
			TenuredHome + "/commons/remoting":  "",
			TenuredHome + "/commons/logs":      "",
			"time":                             "",
		},
	}
}

func (this *Imports) AddInterface(pkg, name string) {
	this.InterfacePackage[pkg] = name
}
func (this *Imports) AddClient(pkg, name string) {
	this.ClientPackage[pkg] = name
}
func (this *Imports) AddInvoke(pkg, name string) {
	this.InvokePackage[pkg] = name
}

func (this *Imports) Is(head string) bool {
	return "imports {" == head
}

func (this *Imports) Add(addLines []string, info *TCDInfo) error {
	_, lines := comment(addLines)
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
		(this.Imports)[importUrl] = alias
	}
	return nil
}

func (this *Imports) InterOuter(info *TCDInfo) []byte {
	b := new(bytes.Buffer)
	ftl(`
import (
	{{range $k,$v :=.Imports}}
	{{$v}} "{{$k}}"{{end}}
	{{range $k,$v :=.InterfacePackage}}
	{{$v}} "{{$k}}"{{end}}
)
`, this, b)
	return b.Bytes()
}

func (this *Imports) ClientOut(info *TCDInfo) []byte {
	b := new(bytes.Buffer)
	ftl(`
import (
	{{range $k,$v :=.Imports}}
	{{$v}} "{{$k}}"{{end}}
	{{range $k,$v :=.ClientPackage}}
	{{$v}} "{{$k}}"{{end}}
)
`, this, b)
	return b.Bytes()
}

func (this *Imports) InvokeOut(info *TCDInfo) []byte {
	b := new(bytes.Buffer)
	ftl(`
import (
	{{range $k,$v :=.Imports}}
	{{$v}} "{{$k}}"{{end}}
	{{range $k,$v :=.InvokePackage}}
	{{$v}} "{{$k}}"{{end}}
)
`, this, b)
	return b.Bytes()
}
