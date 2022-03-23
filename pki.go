package shoset

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"

	// "io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (c *Shoset) InitPKI(address string) error {
	// elle sort immédiatement si :
	if c.GetBindAddress() != "" { // il n'y a pas encore eu de bind (bindadress est vide)
		fmt.Println("shoset bound")
		return errors.New("shoset bound")
	} else if c.ConnsByName.Len() != 0 { // j'ai déjà fait un link ou un join ou j'ai un fichier de configuration (ce qui veut dire que j'ai des connsbyname)
		fmt.Println("a protocol already happened on this shoset")
		return errors.New("a protocol already happened on this shoset")
	} else if c.GetIsPki() { // il y a eu déjà un init ou j'ai déjà un certificat (mon certificat existe déjà)
		fmt.Println("shoset already initialized")
		return errors.New("shoset already initialized")
	}

	c.SetIsPki(true) // je prends le role de CA de la PKI

	// demarche d'initialisation de bind classique (shoset.go/bind)
	ipAddress, err := GetIP(address) // parse the address from function parameter to get the IP
	if err != nil {                  // check if IP is ok
		return err
	}
	_ipAddress := strings.Replace(ipAddress, ":", "_", -1)
	_ipAddress = strings.Replace(_ipAddress, ".", "-", -1)
	c.ConnsByName.SetConfigName(_ipAddress)

	dirname, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Get UserHomeDir error : ", err)
		return err
	}
	if !fileExists(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/") {
		os.Mkdir(dirname+"/.shoset/", 0700)
		os.Mkdir(dirname+"/.shoset/"+c.ConnsByName.GetConfigName()+"/", 0700)
		os.Mkdir(dirname+"/.shoset/"+c.ConnsByName.GetConfigName()+"/config/", 0700)
		os.Mkdir(dirname+"/.shoset/"+c.ConnsByName.GetConfigName()+"/cert/", 0700)
	}

	// Create Certificate Authority
	//https://fale.io/blog/2017/06/05/create-a-pki-in-golang/
	CAcert := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Organization:  []string{"Ditrit"},
			Country:       []string{"33"},
			Province:      []string{"France"},
			Locality:      []string{"Paris"},
			StreetAddress: []string{"19 Rue Bergère"},
			PostalCode:    []string{"75009"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	CAprivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)                                               // private key rsa format
	CApublicKey := &CAprivateKey.PublicKey                                                              // we extract the public key (CA cert) from the private key
	signedCAcert, err := x509.CreateCertificate(rand.Reader, CAcert, CAcert, CApublicKey, CAprivateKey) // we sign the certificate
	if err != nil {
		return errors.New("couldn't create CA")
	}

	CAcertFile, err := os.Create(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/CAcert.crt")
	if err != nil {
		return err
	}
	pem.Encode(CAcertFile, &pem.Block{Type: "CERTIFICATE", Bytes: signedCAcert})
	CAcertFile.Close()

	// Private key
	CAprivateKeyFile, err := os.OpenFile(dirname+"/.shoset/"+c.ConnsByName.GetConfigName()+"/cert/privateCAKey.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	pem.Encode(CAprivateKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(CAprivateKey)})
	CAprivateKeyFile.Close()

	// Create and sign additional certificates - here the certificate of the socket from the CA
	certReq, hostPublicKey, _ := c.PrepareCertificate()
	if certReq != nil && hostPublicKey != nil {
		signedHostCert := c.SignCertificate(certReq, hostPublicKey)
		if signedHostCert != nil {

			certFile, err := os.Create(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/cert.crt")
			if err != nil {
				return err
			}
			pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: signedHostCert})
			certFile.Close()

			// Public key
			// ioutil.WriteFile(dirname+"/.shoset/"+c.ConnsByName.GetConfigName()+"/cert/cert.crt", signedHostCert, 0644)
		} else {
			return errors.New("prepare certificate didn't work")
		}
	} else {
		return errors.New("prepare certificate didn't work")
	}

	// 3. Elle associe le rôle 'pki' au nom logique de la shoset
	c.SetIsCertified(true)
	c.Bind(address)

	return nil
}

// pour les shoset ayant le rôle 'pki' :
// 1. Service activé pour les deux fonctions
// getsecret(login, password) => { secret }
func (c *Shoset) GenerateSecret(login, password string) string {
	if c.GetIsPki() {
		// utiliser login et password
		return uuid.New().String()
	}
	return ""
}

func (c *Shoset) PrepareCertificate() (*x509.Certificate, *rsa.PublicKey, *rsa.PrivateKey) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, nil
	}

	// Prepare new certificate
	certReq := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization:  []string{"Ditrit"},
			Country:       []string{"33"},
			Province:      []string{"France"},
			Locality:      []string{"Paris"},
			StreetAddress: []string{"19 Rue Bergère"},
			PostalCode:    []string{"75009"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// Private and public keys
	hostPrivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	hostPublicKey := &hostPrivateKey.PublicKey

	// Private key
	hostPrivateKeyFile, err := os.OpenFile(dirname+"/.shoset/"+c.ConnsByName.GetConfigName()+"/cert/privateKey.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, nil, nil
	}
	pem.Encode(hostPrivateKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(hostPrivateKey)})
	hostPrivateKeyFile.Close()
	return certReq, hostPublicKey, hostPrivateKey
}

func (c *Shoset) SignCertificate(certReq *x509.Certificate, hostPublicKey *rsa.PublicKey) []byte {
	if c.GetIsPki() {
		dirname, err := os.UserHomeDir()
		if err != nil {
			return nil
		}

		// Load CA
		catls, err := tls.LoadX509KeyPair(dirname+"/.shoset/"+c.ConnsByName.GetConfigName()+"/cert/CAcert.crt", dirname+"/.shoset/"+c.ConnsByName.GetConfigName()+"/cert/privateCAKey.key")
		if err != nil {
			return nil
		}

		ca, err := x509.ParseCertificate(catls.Certificate[0]) // we parse the previous certificate
		if err != nil {
			return nil
		}

		// Sign the certificate
		signedHostCert, err := x509.CreateCertificate(rand.Reader, certReq, ca, hostPublicKey, catls.PrivateKey)
		if err != nil {
			return nil
		}
		return signedHostCert
	}
	return nil
}
