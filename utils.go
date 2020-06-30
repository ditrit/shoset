package shoset

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
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
			}
			return 0, false
		}
		return 0, false
	}
	return 0, false
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
		if val.ShosetType == shosetType {
			result = append(result, val)
		}
	}
	//m.Unlock()
	return result
}

func checkCA(cert []byte) bool {
	block, _ := pem.Decode(cert)
	if block == nil || block.Type != "CERTIFICATE" {
		return false
	}
	c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}
	return c.IsCA
}

func genPrivKey() (key, pub []byte, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return
	}
	privByte, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return
	}
	pubByte, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return
	}
	var privBuffer, pubBuffer *bytes.Buffer
	pem.Encode(privBuffer, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privByte,
	})
	pem.Encode(pubBuffer, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubByte,
	})
	key = privBuffer.Bytes()
	pub = pubBuffer.Bytes()
	return
}

// signCert :
func signCert(cn string, pub, ca, key []byte) (cert []byte, err error) {
	pubBlock, _ := pem.Decode(pub)
	if pubBlock == nil || !strings.HasSuffix(pubBlock.Type, "PUBLIC KEY") {
		err = fmt.Errorf("SignCert : invalid public key")
		return
	}
	caBlock, _ := pem.Decode(ca)
	if caBlock == nil || caBlock.Type != "CERTIFICATE" {
		err = fmt.Errorf("SignCert : invalid public key")
		return
	}
	keyBlock, _ := pem.Decode(key)
	if keyBlock == nil || !strings.HasSuffix(keyBlock.Type, "PRIVATE KEY") {
		err = fmt.Errorf("SignCert : invalid public key")
		return
	}

	pubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return
	}
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return
	}
	caKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return
	}

	certTemplate := &x509.Certificate{
		Issuer:  caCert.Subject,
		Subject: pkix.Name{CommonName: cn},
	}
	certByte, err := x509.CreateCertificate(rand.Reader, certTemplate, caCert, pubKey, caKey)
	if err != nil {
		return
	}
	var certBuffer *bytes.Buffer
	pem.Encode(certBuffer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certByte,
	})
	cert = certBuffer.Bytes()
	return
}
