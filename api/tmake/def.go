package main

import (
	"bytes"
	"errors"
)

var NotMatch = errors.New("not match")

type Module interface {
	Add(lines []string, info *TCDInfo) error
}

type Interface interface {
	InterOuter(info *TCDInfo) []byte
}

type Client interface {
	ClientOut(info *TCDInfo) []byte
}

type ServerInvoke interface {
	InvokeOut(info *TCDInfo) []byte
}

type Def struct {
	modules []Module
}

func (def *Def) Add(lines []string, info *TCDInfo) error {
	for _, module := range def.modules {
		if err := module.Add(lines, info); err != nil && err != NotMatch {
			return err
		}
	}
	return nil
}

func (def *Def) Interface(tcd *TCDInfo) []byte {
	b := new(bytes.Buffer)
	b.WriteString("//generator by tenured command defined.\n")
	b.WriteString("package ")
	b.WriteString(tcd.ApiPackageName)
	b.WriteString("\n\n\n")

	for _, module := range def.modules {
		if inter, match := module.(Interface); match {
			if bs := inter.InterOuter(tcd); bs != nil {
				b.Write(bs)
			}
		}
	}
	return b.Bytes()
}

func (def *Def) Client(tcd *TCDInfo) []byte {
	b := new(bytes.Buffer)
	b.WriteString("package ")
	b.WriteString(tcd.ClientPackageName)
	b.WriteString("\n\n\n")

	for _, module := range def.modules {
		if client, match := module.(Client); match {
			if bs := client.ClientOut(tcd); bs != nil {
				b.Write(bs)
			}
		}
	}

	return b.Bytes()
}

func (def *Def) Invoke(tcd *TCDInfo) []byte {
	b := new(bytes.Buffer)
	b.WriteString("package ")
	b.WriteString(tcd.InvokePackageName)
	b.WriteRune('\n')

	for _, module := range def.modules {
		if client, match := module.(ServerInvoke); match {
			if bs := client.InvokeOut(tcd); bs != nil {
				b.Write(bs)
			}
		}
	}

	return b.Bytes()
}

func NewDef(tcd *TCDInfo) *Def {
	imports := NewImport(tcd)
	lbs := NewLoadBalance()
	enums := NewEunms()
	typeDefs := NewTypes(enums)
	return &Def{
		modules: []Module{
			imports, NewErrors(),
			enums, typeDefs,
			lbs, NewServicesDef(imports, lbs, typeDefs),
		},
	}
}
