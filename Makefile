

build:
	go build -ldflags "-w -s" -buildmode=plugin -o engine.so testplugin.go

install:
	@echo "do nothing"

.PHONY: clean
clean:
	@rm -rf plugins
	@rm -rf bin