package net

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

// GetIP :
func GetIP(address string) (string, error) {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		return "", errors.New("address '" + address + "should respect the format hots_name_or_ip:port")
	}
	hostIps, err := net.LookupHost(parts[0])
	if err != nil || len(hostIps) == 0 {
		return "", errors.New("address '" + address + "' can not be resolved")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", errors.New("'" + parts[1] + "' is not a port number")
	}
	if port < 1 || port > 65535 {
		return "", errors.New("'" + parts[1] + "' is not a valid port number")
	}
	host := getV4(hostIps)
	if host == "" {
		return "", errors.New("failed to get ipv4 address for localhost")
	}
	ipaddr := host + ":" + parts[1]
	return ipaddr, nil
}

// Grab ip4/6 string array and return an ipv4 str
func getV4(hostIps []string) string {
	for i := 0; i < len(hostIps); i++ {
		if net.ParseIP(hostIps[i]).To4() != nil {
			return hostIps[i]
		}
	}
	return ""
}
