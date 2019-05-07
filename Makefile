Version=1.0.0
BuildDate=$(shell date +"%F %T")
GitCommit=$(shell git rev-parse HEAD)

param=-X main.VERSION=${Version} -X main.GIT_VERSION=${GitCommit} -X 'main.BUILD_TIME=${BuildDate}'

tenured=bin/tenured

ifeq ($(P),release)
debug=-w -s
else
debug=
endif

build: generate plugins
	go build -ldflags "${debug} ${param}" -o ${tenured} tenured.go

build-windows: generate plugins
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "${debug} ${param}" -o ${tenured}.exe tenured.go

build-linux: generate plugins
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "${debug} ${param}" -o ${tenured} tenured.go

depends:
	GOPROXY=https://goproxy.io GO111MODULE=on go mod download

generate:
	go generate api/generator.go

plugin_registry_eureka:
	go build ${debug} -buildmode=plugin -o plugins/registry/eureka.so ./plugins/registry/eureka

plugins:
	@echo "none plugins"

install:
	@mkdir -p distribution
	@cp -r bin distribution
	@cp -r plugins/*/*.so distribution
	@cp -r conf distribution

upx:
	upx -9 -k ${tenured}

.PHONY: clean
clean:
	@rm -rf bin
	@rm -rf distribution