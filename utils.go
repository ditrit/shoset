package shoset

import (
	"errors"
	"fmt"
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

// IP2ID :
func IP2ID(ip string) (uint64, bool) {
	parts := strings.Split(ip, ":")
	if len(parts) == 2 {
		nums := strings.Split(parts[0], ".")
		if len(nums) == 4 {
			idStr := fmt.Sprintf("%s%s%s%s%s", nums[0], nums[1], nums[2], nums[3], parts[1])
			id, err := strconv.ParseUint(idStr, 10, 64)
			if err == nil {
				return id, true
			} else {
				return 0, false
			}
		} else {
			return 0, false
		}
	} else {
		return 0, false
	}
}

// DeltaAddress return a new address with same host but with a new port (old one with an offset)
func DeltaAddress(addr string, portDelta int) (string, bool) {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 {
		port, err := strconv.Atoi(parts[1])
		if err == nil {
			return fmt.Sprintf("%s:%d", parts[0], port+portDelta), true
		}
		return "", false
	}
	return "", false
}

// GetByType : Get shoset by type.
func GetByType(m *MapSafeConn, shosetType string) []*ShosetConn {
	var result []*ShosetConn
	//m.Lock()
	for _, val := range m.GetM() {
		if val.GetShosetType() == shosetType {
			result = append(result, val)
		}
	}
	//m.Unlock()
	return result
}
