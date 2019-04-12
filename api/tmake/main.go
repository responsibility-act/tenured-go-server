package main

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const TenuredHome = "github.com/ihaiker/tenured-go-server"

type TCDInfo struct {
	TCDFile string

	ApiFileName    string
	ApiPackageName string
	ApiPackageUrl  string

	ClientFileName    string
	ClientPackageName string
	ClientPackageUrl  string

	InvokeFileName    string
	InvokePackageName string
	InvokePackageUrl  string
}

func NewTCD(tcdFile string) *TCDInfo {
	tcd := &TCDInfo{}
	goPath := filepath.Join(os.Getenv("GOPATH"), "src") + "/"

	dir := filepath.Dir(tcdFile)
	name := strings.Replace(filepath.Base(tcdFile), ".tcd", "", 1)

	tcd.ApiFileName = dir + "/" + name + "_tcd.go"
	tcd.ApiPackageName = filepath.Base(dir)
	tcd.ApiPackageUrl = strings.Replace(dir, goPath, "", 1)

	tcd.ClientPackageName = "client"
	tcd.InvokePackageName = "invoke"

	tcd.ClientPackageUrl = tcd.ApiPackageUrl + "/" + tcd.ClientPackageName
	tcd.ClientFileName = dir + "/" + tcd.ClientPackageName + "/" + name + "_tcd.go"

	tcd.InvokePackageUrl = tcd.ApiPackageUrl + "/" + tcd.InvokePackageName
	tcd.InvokeFileName = dir + "/" + tcd.InvokePackageName + "/" + name + "_tcd.go"
	return tcd
}

func main() {
	if len(os.Args) == 1 {
		log.Fatal("至少制定一个接口配置文件")
	}

	apis := os.Args[1:]
	for k, v := range apis {
		absPath, _ := filepath.Abs(v)
		apis[k] = absPath
	}

	for index, api := range apis {
		log.Printf("%d: %s", index, api)
		read, err := NewReader(api)
		if err != nil {
			log.Panic(err)
		}

		tcd := NewTCD(api)

		def := NewDef(tcd)
		for {
			if lines, err := read.Read(); err == io.EOF {
				break
			} else if err = def.Add(lines, tcd); err != nil {
				log.Panic(err)
			}
		}

		//interface
		{
			f := commons.NewFile(tcd.ApiFileName)
			if f.Exist() {
				_ = os.Remove(f.GetPath())
			}

			if w, err := f.GetWriter(false); err != nil {
				log.Panic(err)
			} else {
				defer w.Close()
				if _, err = w.Write(def.Interface(tcd)); err != nil {
					log.Panic(err)
				}
			}
		}

		//client
		{
			f := commons.NewFile(tcd.ClientFileName)
			if f.Exist() {
				_ = os.Remove(f.GetPath())
			}
			if err := f.Parent().Mkdir(); err != nil {
				log.Panic(err)
			}

			if w, err := f.GetWriter(false); err != nil {
				log.Panic(err)
			} else {
				defer func() {
					_ = w.Close()
				}()
				if _, err = w.Write(def.Client(tcd)); err != nil {
					log.Panic(err)
				}
			}
		}

		//invoke
		//client
		{
			f := commons.NewFile(tcd.InvokeFileName)
			if f.Exist() {
				_ = os.Remove(f.GetPath())
			}
			if err := f.Parent().Mkdir(); err != nil {
				log.Panic(err)
			}

			if w, err := f.GetWriter(false); err != nil {
				log.Panic(err)
			} else {
				defer func() {
					_ = w.Close()
				}()
				if _, err = w.Write(def.Invoke(tcd)); err != nil {
					log.Panic(err)
				}
			}
		}
	}
}
