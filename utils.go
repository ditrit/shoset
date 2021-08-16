package shoset

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/howeyc/gopass"
	"github.com/square/certstrap/depot"
	"github.com/square/certstrap/pkix"
	"github.com/urfave/cli"
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
		if val.GetRemoteShosetType() == shosetType {
			result = append(result, val)
		}
	}
	//m.Unlock()
	return result
}



// pki
// https://github.com/square/certstrap/tree/master/cmd
var (
	d        *depot.FileDepot
	depotDir string
)

// InitDepot creates the depot directory, which stores key/csr/crt files
func InitDepot(path string) error {
	depotDir = path
	if d == nil {
		var err error
		if d, err = depot.NewFileDepot(path); err != nil {
			return err
		}
	}
	return nil
}

func createPassPhrase() ([]byte, error) {
	pass1, err := gopass.GetPasswdPrompt("Enter passphrase (empty for no passphrase): ", false, os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}
	pass2, err := gopass.GetPasswdPrompt("Enter same passphrase again: ", false, os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(pass1, pass2) {
		return nil, errors.New("Passphrases do not match.")
	}
	return pass1, nil
}

func askPassPhrase(name string) ([]byte, error) {
	pass, err := gopass.GetPasswdPrompt(fmt.Sprintf("Enter passphrase for %v (empty for no passphrase): ", name), false, os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}
	return pass, nil
}

func getPassPhrase(c *cli.Context, name string) ([]byte, error) {
	if c.IsSet("passphrase") {
		return []byte(c.String("passphrase")), nil
	}
	return askPassPhrase(name)
}

func putCertificate(c *cli.Context, d *depot.FileDepot, name string, crt *pkix.Certificate) error {
	if c.IsSet("cert") {
		bytes, err := crt.Export()
		if err != nil {
			return err
		}
		return ioutil.WriteFile(c.String("cert"), bytes, depot.LeafPerm)
	}
	return depot.PutCertificate(d, name, crt)
}

func putCertificateSigningRequest(c *cli.Context, d *depot.FileDepot, name string, csr *pkix.CertificateSigningRequest) error {
	if c.IsSet("csr") {
		bytes, err := csr.Export()
		if err != nil {
			return err
		}
		return ioutil.WriteFile(c.String("csr"), bytes, depot.LeafPerm)
	}
	return depot.PutCertificateSigningRequest(d, name, csr)
}

func getCertificateSigningRequest(c *cli.Context, d *depot.FileDepot, name string) (*pkix.CertificateSigningRequest, error) {
	if c.IsSet("csr") {
		bytes, err := ioutil.ReadFile(c.String("csr"))
		if err != nil {
			return nil, err
		}
		return pkix.NewCertificateSigningRequestFromPEM(bytes)
	}
	return depot.GetCertificateSigningRequest(d, name)
}

func putEncryptedPrivateKey(c *cli.Context, d *depot.FileDepot, name string, key *pkix.Key, passphrase []byte) error {
	if c.IsSet("key") {
		if fileExists(c.String("key")) {
			return nil
		}

		bytes, err := key.ExportEncryptedPrivate(passphrase)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(c.String("key"), bytes, depot.BranchPerm)
	}
	return depot.PutEncryptedPrivateKey(d, name, key, passphrase)
}

func putPrivateKey(c *cli.Context, d *depot.FileDepot, name string, key *pkix.Key) error {
	if c.IsSet("key") {
		if fileExists(c.String("key")) {
			return nil
		}

		bytes, err := key.ExportPrivate()
		if err != nil {
			return err
		}
		return ioutil.WriteFile(c.String("key"), bytes, depot.BranchPerm)
	}
	return depot.PutPrivateKey(d, name, key)
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

var nowFunc = time.Now

func parseExpiry(fromNow string) (time.Time, error) {
	now := nowFunc().UTC()
	re := regexp.MustCompile(`\s*(\d+)\s*(day|month|year|hour|minute|second)s?`)
	matches := re.FindAllStringSubmatch(fromNow, -1)
	addDate := map[string]int{
		"day":    0,
		"month":  0,
		"year":   0,
		"hour":   0,
		"minute": 0,
		"second": 0,
	}
	for _, r := range matches {
		number, err := strconv.ParseInt(r[1], 10, 32)
		if err != nil {
			return now, err
		}
		addDate[r[2]] = int(number)
	}

	// Ensure that we do not overflow time.Duration.
	// Doing so is silent and causes signed integer overflow like issues.
	if _, err := time.ParseDuration(fmt.Sprintf("%dh", addDate["hour"])); err != nil {
		return now, fmt.Errorf("hour unit too large to process")
	} else if _, err = time.ParseDuration(fmt.Sprintf("%dm", addDate["minute"])); err != nil {
		return now, fmt.Errorf("minute unit too large to process")
	} else if _, err = time.ParseDuration(fmt.Sprintf("%ds", addDate["second"])); err != nil {
		return now, fmt.Errorf("second unit too large to process")
	}

	result := now.
		AddDate(addDate["year"], addDate["month"], addDate["day"]).
		Add(time.Duration(addDate["hour"]) * time.Hour).
		Add(time.Duration(addDate["minute"]) * time.Minute).
		Add(time.Duration(addDate["second"]) * time.Second)

	if now == result {
		return now, fmt.Errorf("invalid or empty format")
	}

	// ASN.1 (encoding format used by SSL) only supports up to year 9999
	// https://www.openssl.org/docs/man1.1.0/crypto/ASN1_TIME_check.html
	if result.Year() > 9999 {
		return now, fmt.Errorf("proposed date too far in to the future: %s. Expiry year must be less than or equal to 9999", result)
	}

	return result, nil
}
