package net

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

func getIP(address string) (string, error) {
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
	ipaddr := hostIps[0] + ":" + parts[1]
	return ipaddr, nil
}
