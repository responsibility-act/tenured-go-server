package mixins

import (
	"os"
	"strconv"
	"strings"
)

//服务前缀
const KeyServerPrefix = "tenured.prefix"
const ServerPrefix = "tenured"

const KeyRegistry = "tenured.registry"
const Registry = "consul://127.0.0.1:8500"

const KeyDataPath = "tenured.dataPath"
const DataPath = "/data/tenured"

const PortStore = 6072
const PortLinker = 6073
const PortConsole = 6074
const PortTenant = 6075

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

func GetInt(key string, value int) int {
	if val, has := os.LookupEnv(key); has {
		if i, er := strconv.Atoi(val); er == nil {
			return i
		}
	}
	envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	if val, has := os.LookupEnv(envKey); has {
		if i, er := strconv.Atoi(val); er == nil {
			return i
		}
	}
	return value
}

func serverName(prefix, server string) string {
	if prefix == "" {
		return server
	} else {
		return prefix + "_" + server
	}
}

func Store(prefix string) string {
	return serverName(prefix, "store")
}

func Linker(prefix string) string {
	return serverName(prefix, "linker")
}

func Console(prefix string) string {
	return serverName(prefix, "console")
}

func Tenant(prefix string) string {
	return serverName(prefix, "tenant")
}
