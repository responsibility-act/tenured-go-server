package nets

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
)

func RandPort(host string, start, end int) (int, error) {
	for i := start; i <= end; i++ {
		if lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, i)); err == nil {
			_ = lis.Close()
			return i, nil
		}
	}
	return 0, errors.New("not use any port")
}

func GetExternal() (string, error) {
	for _, v := range []string{"http://myexternalip.com/raw", "http://ip.renzhen.la/raw"} {
		resp, err := http.Get(v)
		if err != nil {
			continue
		}
		defer func() { _ = resp.Body.Close() }()

		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	return "", errors.New("Cannot get external address")
}

func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

/**
ignoreInterfaces 忽略的网卡
preferredNetworks 更倾向使用的网卡地址
*/
func GetLocalIP(ignoredInterfaces []string, preferredNetworks []string) (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	localIps := make([]string, 0)

SEARCH:
	for _, inter := range interfaces {
		if !(inter.Flags&net.FlagUp > 0 && inter.Flags&net.FlagPointToPoint == 0 &&
			inter.Flags&net.FlagLoopback == 0) {
			continue SEARCH
		}

		if ignoredInterfaces != nil && len(ignoredInterfaces) > 0 {
			for _, v := range ignoredInterfaces {
				if v == inter.Name ||
					regexp.MustCompile(v).Match([]byte(inter.Name)) {
					continue SEARCH
				}
			}
		}
		address, _ := inter.Addrs()
		for _, addr := range address {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					localIps = append(localIps, ipnet.IP.String())
				}
			}
		}
	}

	if len(localIps) == 0 {
		return "", errors.New("not found net Interface")
	}

	if preferredNetworks != nil && len(preferredNetworks) > 0 {
		for _, preferredNetwork := range preferredNetworks {
			preferredRegex := strings.ReplaceAll(preferredNetwork, ".", "\\.")
			preferredRegex = strings.ReplaceAll(preferredNetwork, "*", "(\\d+)")
			for _, localIp := range localIps {
				if regexp.MustCompile(preferredRegex).Match([]byte(localIp)) {
					return localIp, nil
				}
			}
		}
	}
	return localIps[0], nil
}
