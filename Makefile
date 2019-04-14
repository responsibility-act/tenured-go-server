version=1.0.0
build_date=`date +"%F %T"`
param=-X main.VERSION=${version} -X 'main.BUILD_TIME=${build_date}' -X 'main.GO_VERSION=`go version`'

tenured=bin/tenured

ifeq ($(P),release)
debug=-ldflags "-w -s"
else
debug=
endif

build: generate plugins main

main:
	go build ${debug} -ldflags "${param}" -o ${tenured} tenured.go

generate:
	go generate api/generator.go

plugin_registry_consul:
	go build ${debug} -buildmode=plugin -o plugins/registry/consul.so ./commons/registry/consul

plugins: plugin_registry_consul


install:
	@mkdir -p distribution
	@cp -r bin distribution
	@cp -r plugins distribution
	@cp -r conf distribution

upx:
	upx -9 -k ${tenured}


build:
	go build -ldflags "-w -s" -buildmode=plugin -o engine.so testplugin.go

install:
	@echo "do nothing"

.PHONY: clean
clean:
	@rm -rf plugins
	@rm -rf bin
	@rm -rf distribution