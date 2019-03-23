package mixins

import (
	"os"
	"strings"
)

//服务前缀
const KeyServerPrefix = "tenured.prefix"
const ServerPrefix = "tenured"

const KeyRegistry = "tenured.registry"
const Registry = "consul://127.0.0.1:8500"

func Get(key, value string) string {
	if val, has := os.LookupEnv(key); has {
		return val
	}
	envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	if val, has := os.LookupEnv(envKey); has {
		return val
	}
	return value
}
