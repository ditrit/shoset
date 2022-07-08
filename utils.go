package shoset

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/howeyc/gopass"
	"github.com/square/certstrap/depot"
	"github.com/square/certstrap/pkix"
	"github.com/urfave/cli"
)

// fileExists returns true if the path indicated corresponds to an existing file
func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

// mkdir creates a repertory if it doesn't already exist
func mkdir(path string) error {
	if !fileExists(path) {
		return os.Mkdir(path, 0700)
	}
	return nil
}

// contains range through a slice to search for a particular string
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// removeDuplicateStr range through a slice to delete string duplicates
func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// Keys returns a []string corresponding to the keys from the map[string]*sync.Map object.
// direction set the specific keys depending on the desired direction.
func Keys(mapSync *sync.Map, direction string) []string {
	var keys []string
	if direction == ALL {
		// all keys whatever the direction from the ShosetConn
		mapSync.Range(func(key, _ interface{}) bool {
			keys = append(keys, key.(string))
			return true
		})
	} else {
		// keys with specific ShosetConn direction
		mapSync.Range(func(key, value interface{}) bool {
			if value.(*ShosetConn).GetDirection() == direction {
				keys = append(keys, key.(string))
			}
			return true
		})
	}
	return removeDuplicateStr(keys)
}

// LenSyncMap returns the length of an sync.Map object.
func LenSyncMap(m *sync.Map) int {
	var i int
	m.Range(func(k, v interface{}) bool {
		i++
		return true
	})
	return i
}

// EncodeFile encodes bytes to a specific encoding type to a specific path
func EncodeFile(object interface{}, encodeType, path string) error {
	switch encodeType {
	case CERTIFICATE:
		file, err := os.Create(path)
		if err != nil {
			return errors.New("couldn't create cert file : " + err.Error())
		}
		defer file.Close()

		err = pem.Encode(file, &pem.Block{Type: encodeType, Bytes: object.([]byte)})
		if err != nil {
			return errors.New("couldn't encode in cert file : " + err.Error())
		}
	case RSA_PRIVATE_KEY:
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return errors.New("couldn't open private key file : " + err.Error())
		}
		defer file.Close()

		err = pem.Encode(file, &pem.Block{Type: encodeType, Bytes: x509.MarshalPKCS1PrivateKey(object.(*rsa.PrivateKey))})
		if err != nil {
			return errors.New("couldn't encode in private key file : " + err.Error())
		}
	}
	return nil
}

// GetPrivateKey returns the private key contained in a file
func GetPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	CAprivateKeyBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(CAprivateKeyBytes)
	b := block.Bytes

	CAprivateKey, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		return nil, err
	}
	return CAprivateKey, nil
}

// GetIP returns an IPaddress format from an address
func GetIP(address string) (string, error) {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		return VOID, errors.New("address '" + address + "should respect the format hots_name_or_ip:port")
	}
	hostIps, err := net.LookupHost(parts[0])
	if err != nil || len(hostIps) == 0 {
		return VOID, errors.New("address '" + address + "' can not be resolved")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return VOID, errors.New("'" + parts[1] + "' is not a port number")
	}
	if port < 1 || port > 65535 {
		return VOID, errors.New("'" + parts[1] + "' is not a valid port number")
	}
	host := getV4(hostIps)
	if host == VOID {
		return VOID, errors.New("failed to get ipv4 address for localhost")
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
	return VOID
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
		return VOID, false
	}
	return VOID, false
}

// Functions for PKI secret - will be used and documented later
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
