package utils

import (
	"errors"
	"net"
)

// GetLocalIP 获取本机内网 IPv4 地址
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		ip := ipNet.IP

		// 跳过回环地址，例如 127.0.0.1
		// 跳过非内网地址
		if ip.IsLoopback() || !ip.IsPrivate() {
			continue
		}

		// 只要 IPv4
		ipv4 := ip.To4()
		if ipv4 == nil {
			continue
		}

		return ipv4.String(), nil
	}

	return "", errors.New("ERR_NO_LOCAL_IP_FOUND")
}
