package main

import (
	"bytes"
	"fmt"
	"github.com/kataras/iris/core/errors"
	"log"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var servicePattern = regexp.MustCompile(`^service (\w+)\(([0-9]{4,5})\)( loadBalance\((\w+)\))?[ ]?\{$`)
var funcPattern = regexp.MustCompile(`^(\w+)\(([ ,\[\]\w]*)\) \(([ ,\[\]\w]*)\)( error\(([,\w]+)\))?( loadBalance\((\w+)\))?( timeout\((\w+)\))?$`)

type FunParam struct {
	Name string
	Type string
	tcd  *TCDInfo
}

func (this *FunParam) IsBody() bool {
	return this.Type == "[]byte"
}

func (this *FunParam) UpperName() string {
	return UpperName(this.Name)
}

func isBase(t string) bool {
	switch t {
	case "bool",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"map[string]string",
		"[]byte",
		"string", "[]string":
		return true
	default:
		return false
	}
}

func isArray(t string) bool {
	return strings.HasPrefix(t, "[]")
}

func (this *FunParam) UseShowType() string {
	if isBase(this.Type) {
		return this.Type
	} else if isArray(this.Type) {
		if isBase(this.Type[2:]) {
			return this.Type
		} else {
			return "[]*" + this.tcd.ApiPackageName + "." + this.Type[2:]
		}
	} else {
		return "*" + this.tcd.ApiPackageName + "." + this.Type
	}
}

func (this *FunParam) ShowType() string {
	if isBase(this.Type) {
		return this.Type
	} else if isArray(this.Type) {
		if isBase(this.Type[2:]) {
			return this.Type
		} else {
			return "[]*" + this.Type[2:]
		}
	} else {
		return "*" + this.Type
	}
}

type FuncDef struct {
	Name        string
	Desc        string
	Errors      []string
	LoadBalance string
	Ins         []FunParam
	Outs        []FunParam

	Types *TypesDef

	RequestCode uint16

	serviceDef ServiceDef
	tcd        *TCDInfo

	Timeout string
}

func (this *FuncDef) TimeoutDuration() string {
	timeout, err := time.ParseDuration(this.Timeout)
	if err != nil {
		log.Panic("方法" + this.serviceDef.Name + "." + this.Name + " timeout定义错误,查阅 time.ParseDuration")
	}
	timeoutMillisecond := timeout.Nanoseconds() / int64(time.Millisecond)
	return fmt.Sprintf("time.Millisecond*%d", timeoutMillisecond)
}

func (this *FuncDef) ClientBody() string {
	b := new(bytes.Buffer)

	loadBanlanceParam := this.tcd.ApiPackageName + "." + this.serviceDef.Name + this.Name
	if this.LoadBalance == "none" {
		loadBanlanceParam += ",gl"
	}
	if len(this.Ins) > 0 {
		for _, v := range this.Ins {
			loadBanlanceParam += "," + v.Name
		}
	}

	b.WriteString(fmt.Sprintf(`
			serverInstance,regKey, err := this.%sLB.Select(%s)
			if err != nil || len(serverInstance) == 0 || registry.AllNotOK(serverInstance...) {
				return %s protocol.ErrorRouter()
			}
			defer this.%sLB.Return(regKey)
		`, this.LoadBalance, loadBanlanceParam, strings.Repeat("nil,", len(this.Outs)), this.LoadBalance,
	))

	//header
	if len(this.Ins) == 0 {
		b.WriteString(`
			requestHeader := (interface{})(nil)
		`)
	} else if !isBase(this.Ins[0].Type) {
		b.WriteString("	requestHeader := " + this.Ins[0].Name + "\n")
	} else {
		bs := ftlc(`
			requestHeader := &struct{ {{range .}} {{if not .IsBody}}
				{{.UpperName}} {{.Type}} !json:"{{.Name}}"!{{end}}{{end}}
			}{	{{range .}} {{if not .IsBody}}
				{{.UpperName}}: {{.Name}},{{end}}{{end}}
			}
		`, this.Ins)
		b.WriteString(strings.ReplaceAll(string(bs), "!", "`"))
	}
	//body
	hasBody := false
	for _, v := range this.Ins {
		if v.IsBody() {
			hasBody = true
			b.WriteString(fmt.Sprintf(`
				requestBody := %s
			`, v.Name))
			break
		}
	}
	if !hasBody {
		b.WriteString(`
				requestBody :=  []byte(nil)
		`)
	}

	//invoke
	requestCode := fmt.Sprintf("%s.%s%s", this.tcd.ApiPackageName, this.serviceDef.Name, this.Name)
	timeoutMillisecond := this.TimeoutDuration()
	outLength := len(this.Outs)
	if outLength == 0 {
		b.WriteString(fmt.Sprintf(`
			if _, err = this.Invoke(serverInstance[0], %s, requestHeader,requestBody, %s, nil); !commons.IsNil(err) {
				return protocol.ConvertError(err)
			}
			return nil
		`, requestCode, timeoutMillisecond))
	} else if outLength == 1 {
		if "[]byte" == this.Outs[0].Type { //body
			b.WriteString(fmt.Sprintf(`
					var respBody []byte
					if respBody, err = this.Invoke(serverInstance[0], %s, requestHeader,requestBody, %s, nil); !commons.IsNil(err) {
						return nil,protocol.ConvertError(err)
					}else{
						return respBody,nil
					}
				`, requestCode, timeoutMillisecond))
		} else if isBase(this.Outs[0].Type) {
			log.Panic("方法" + this.serviceDef.Name + "." + this.Name + "返回值定义错误，只能为 struct,[]byte两种类型。")
		} else { //from header
			b.WriteString(fmt.Sprintf(`
				respHeader := &%s{}
				if _, err = this.Invoke(serverInstance[0], %s, requestHeader,requestBody, %s, respHeader); !commons.IsNil(err) {
					return nil,protocol.ConvertError(err)
				}else{
					return respHeader,nil
				}
			`, (this.tcd.ApiPackageName + "." + this.Outs[0].Type), requestCode, timeoutMillisecond))
		}
	} else {
		b.WriteString(fmt.Sprintf(`
			respHeader := &%s{}
			var respBody []byte
			if respBody, err = this.Invoke(serverInstance[0], %s, requestHeader,requestBody, %s, respHeader); !commons.IsNil(err) {
				return nil, nil, protocol.ConvertError(err)
			}else{
				return respHeader, respBody, nil
			}
		`, (this.tcd.ApiPackageName + "." + this.Outs[0].Type), requestCode, timeoutMillisecond))
	}
	return string(b.Bytes())
}

func (this *FuncDef) InvokeBody() string {
	b := new(bytes.Buffer)
	st := struct {
		Header  bool
		Bodyer  bool
		Method  string
		Request string
	}{Header: false, Bodyer: false, Method: this.Name}

	st.Request = ""
	if this.LoadBalance == "none" {
		st.Request += "nil,"
	}

	if len(this.Ins) == 0 {

	} else if len(this.Ins) == 1 {
		paramName := this.Ins[0].Name
		if this.Ins[0].IsBody() {
			b.WriteString(fmt.Sprintf(`%s := request.Body`, paramName))
			st.Request += paramName
		} else if !isBase(this.Ins[0].Type) {
			b.WriteString(fmt.Sprintf(`%s := &%s.%s{}`, paramName, this.tcd.ApiPackageName, this.Ins[0].Type))
			st.Request += paramName
			b.WriteString(fmt.Sprintf(`
				if err := request.GetHeader(%s); err != nil {
					logger.Error("InvokeGetHeaderError",err)
				}
			`, paramName))
		} else {
			b.WriteString(fmt.Sprintf("requestHeader := &struct{%s %s}{}", this.Ins[0].UpperName(), this.Ins[0].Type))
			b.WriteString(`
				if err := request.GetHeader(requestHeader); err != nil {
					logger.Error("InvokeGetHeaderError",err)
				}
			`)
			st.Request += "requestHeader." + UpperName(paramName)
		}
	} else {
		b.WriteString(" requestHeader := &struct{")
		hasBody := ""
		for _, v := range this.Ins {
			if v.IsBody() {
				hasBody = fmt.Sprintf(" %s := request.Body \n", v.Name)
				st.Request += v.Name + ","
			} else {
				b.WriteString("\n" + v.Name + " " + v.Type)
				st.Request += "requestHeader." + v.Name + ","
			}
		}
		b.WriteString("}{}\n")
		b.WriteString(hasBody)
		b.WriteString(`
			if err := request.GetHeader(requestHeader); err != nil {
				logger.Error("InvokeGetHeaderError",err)
			}
		`)
	}

	if len(this.Outs) == 0 {

	} else if len(this.Outs) == 1 {
		if "[]byte" == this.Outs[0].Type { //body
			st.Bodyer = true
		} else { //header
			st.Header = true
		}
	} else {
		st.Header = true
		st.Bodyer = true
	}

	ftl(`
		if {{if .Header}}respHeader,{{end}}{{if .Bodyer}}respBody,{{end}} err := service.{{.Method}}({{.Request}}); err != nil {
			response.RemotingError(err)
		} else {
			{{if .Header}} _ = response.SetHeader(respHeader) {{end}}
			{{if .Bodyer}} response.Body = respBody {{end}}
		}
	`, st, b)

	return string(b.Bytes())
}

func NameAndType(p string) (string, string) {
	nt := strings.SplitN(p, " ", 2)
	if len(nt) == 2 {
		return nt[0], nt[1]
	}
	return lowerName(p), p
}

type ServiceDef struct {
	Name               string
	Desc               string
	StartCode          uint16
	Funcs              []FuncDef
	DefaultLoadBalance string
	DefinedLoadBalance map[string]string
}

type ServicesDef struct {
	Imports      *Imports
	LoadBalances *LoadBalancesDef
	Types        *TypesDef
	TCD          *TCDInfo
	Services     []ServiceDef
}

func (this *ServicesDef) Add(addLines []string, info *TCDInfo) error {
	desc, lines := comment(addLines)
	gs, err := match(servicePattern, lines[0])
	if err != nil {
		return err
	}
	startCode, _ := strconv.ParseUint(gs[2], 10, 16)
	serviceDef := ServiceDef{
		Name: gs[1], Desc: desc,
		StartCode: uint16(startCode), Funcs: make([]FuncDef, 0),
		DefaultLoadBalance: gs[4], DefinedLoadBalance: map[string]string{},
	}

	if serviceDef.DefaultLoadBalance != "" {
		serviceDef.DefinedLoadBalance[serviceDef.DefaultLoadBalance] = this.LoadBalances.New(serviceDef.DefaultLoadBalance)
	}

	lines = body(lines)

	startRequestCode := uint16(startCode)

	for ; len(lines) > 0; lines = lines[1:] {
		desc, lines = comment(lines)
		gs, err = match(funcPattern, lines[0])
		if err != nil {
			return errors.New("error at : " + lines[0])
		}
		startRequestCode = startRequestCode + 1
		funDef := FuncDef{
			Name:        gs[1],
			Desc:        desc,
			Types:       this.Types,
			RequestCode: startRequestCode,
			serviceDef:  serviceDef,
			tcd:         info,
			Timeout:     "3s",
		}

		funDef.LoadBalance = gs[7]
		if funDef.LoadBalance == "" {
			funDef.LoadBalance = serviceDef.DefaultLoadBalance
		}
		if funDef.LoadBalance == "" {
			return errors.New(fmt.Sprintf("服务 %s.%s() 方法未定义负载器", serviceDef.Name, funDef.Name))
		}
		if gs[9] != "" {
			funDef.Timeout = gs[9]
		}

		if funDef.LoadBalance == "none" {
			this.Imports.AddInterface(TenuredHome+"/commons/registry", "")
		}
		serviceDef.DefinedLoadBalance[funDef.LoadBalance] = this.LoadBalances.New(funDef.LoadBalance)

		if trim(gs[2]) != "" {
			ins := strings.Split(gs[2], ",")
			funDef.Ins = make([]FunParam, len(ins))
			for i, in := range ins {
				pName, pType := NameAndType(in)
				funDef.Ins[i] = FunParam{Name: pName, Type: pType, tcd: info}
			}
		}
		if trim(gs[3]) != "" {
			outs := strings.Split(gs[3], ",")
			funDef.Outs = make([]FunParam, len(outs))
			for i, out := range outs {
				pName, pType := NameAndType(out)
				funDef.Outs[i] = FunParam{Name: pName, Type: pType, tcd: info}
			}
		}
		//errorss := gs[5]

		serviceDef.Funcs = append(serviceDef.Funcs, funDef)
	}
	this.Services = append(this.Services, serviceDef)
	return nil
}

func (this *ServicesDef) InterOuter(info *TCDInfo) []byte {
	b := new(bytes.Buffer)
	ftl(`
//RequestCode
var (
{{range $i,$s := .Services}}{{range .Funcs}}
		{{$s.Name}}{{.Name}} = uint16({{.RequestCode}})
{{end}}{{end}}
)

{{range .Services}}
{{.Desc}}
type {{.Name}} interface {
	{{range .Funcs}}
	{{.Desc}}
	{{.Name}}({{if eq .LoadBalance "none" }} gl *registry.GlobalLoading,{{end}} {{range $i,$in := .Ins}}{{if gt $i 0}},{{end}} {{.Name}} {{.ShowType}}{{end}} ) ( {{range .Outs}}{{.ShowType}}, {{end}}*protocol.TenuredError )
	{{end}}
}{{end}}`, this, b)
	return b.Bytes()
}

func (this *ServicesDef) ClientOut(tcd *TCDInfo) []byte {
	b := new(bytes.Buffer)
	t := template.Must(template.New("letter").Parse(`
{{range $i,$s  := .Services}}
{{.Desc}}
type {{.Name}}Client struct {
	*protocol.TenuredClientInvoke
	serverName string
	reg     registry.ServiceRegistry
	{{range $name,$v := .DefinedLoadBalance}}
	{{$name}}LB registry.LoadBalance{{end}}
	
	serviceManager *commons.ServiceManager
}

func (this *{{.Name}}Client) Start() error {
	return this.serviceManager.Start()
}

func (this *{{.Name}}Client) Shutdown(interrupt bool) {
	this.serviceManager.Shutdown(interrupt)
}

{{range .Funcs}}
	{{.Desc}}
func (this *{{$s.Name}}Client) {{.Name}}({{if eq .LoadBalance "none" }} gl *registry.GlobalLoading,{{end}}{{range $i,$in := .Ins}}{{if gt $i 0}},{{end}} {{.Name}} {{.UseShowType}}{{end}} ) ( {{range .Outs}}{{.UseShowType}}, {{end}}*protocol.TenuredError ) {
	{{.ClientBody}}
}
{{end}}

func New{{.Name}}Client(serverName string, reg registry.ServiceRegistry) (*{{.Name}}Client, error){
	client := &{{.Name}}Client{
		TenuredClientInvoke: &protocol.TenuredClientInvoke{},
	}
	client.serverName = serverName
	client.reg = reg
	client.serviceManager = commons.NewServiceManager()
	client.serviceManager.Add(client.TenuredClientInvoke)

	{{range $name,$v := .DefinedLoadBalance}}
	client.{{$name}}LB = {{$v}}(serverName,reg)
	client.serviceManager.Add(client.{{$name}}LB)
	{{end}}

	return client, nil
}

{{end}}`))
	_ = t.Execute(b, this)
	return b.Bytes()
}

func (this *ServicesDef) InvokeOut(tcd *TCDInfo) []byte {
	this.TCD = tcd
	b := new(bytes.Buffer)
	ftl(`

{{range $i,$s  := .Services}}
func New{{.Name}}Invoke(tenuredServer *protocol.TenuredServer, service {{$.TCD.ApiPackageName}}.{{.Name}}, manager executors.ExecutorManager) error {
	var logger = logs.GetLogger("invoke")

	{{range .Funcs}}
	{{.Desc}}
	{
		executor := manager.Get("{{$s.Name}}.{{.Name}}")
		tenuredServer.RegisterCommandProcesser({{$.TCD.ApiPackageName}}.{{$s.Name}}{{.Name}}, func(channel remoting.RemotingChannel, request *protocol.TenuredCommand) {
			response := protocol.NewACK(request.ID())
			{{.InvokeBody}}
			if err := channel.Write(response, {{.TimeoutDuration}}); err != nil {
				logger.Error("{{$s.Name}}.{{.Name}} write error: ", err)
			}
		}, executor)
	}
	{{end}}
	return nil
}
{{end}}
	`, this, b)
	return b.Bytes()
}

func NewServicesDef(importDef *Imports, lbs *LoadBalancesDef, typeDefs *TypesDef) *ServicesDef {
	return &ServicesDef{
		Imports: importDef, LoadBalances: lbs, Types: typeDefs,
		Services: make([]ServiceDef, 0),
	}
}
