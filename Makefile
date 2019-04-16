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

redis:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/redis-sync  redis.go

.PHONY: clean
clean:
	@rm -rf bin
	@rm -rf distribution