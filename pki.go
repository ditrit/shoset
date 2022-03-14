package shoset

import (
	"errors"
	"fmt"

	// "io/ioutil"
	"math/big"
	"strings"

	// "io/ioutil"
	"os"

	"time"

	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"

	"github.com/google/uuid"
	// "github.com/square/certstrap/pkix"
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

	// la socket réalise les actions suivantes :
	// 1. Devient la CA
	// -> créer clef privée CA
	// CAkey, err := pkix.CreateRSAKey(4096) // on créer une clé avec le format RSA
	// if err != nil {
	// 	fmt.Println("Create CA RSA Key error : ", err)
	// 	return err
	// }

	// CAprivateKey, err := CAkey.ExportPrivate() // on extrait la clé privée à partir de la clé précédente sous la forme d'un tableau de bytes
	// if err != nil {
	// 	fmt.Println("Export CA RSA Key error : ", err)
	// 	return err
	// }

	// CAprivateKeyFile, err := os.Create(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/privateCAKey.pem") // on créer le fichier qui stocke la clé créée
	// if err != nil {
	// 	fmt.Println("Create CA RSA Key file error : ", err)
	// 	return err
	// }
	// _, err = CAprivateKeyFile.Write(CAprivateKey) // on écrit la clé dans le fichier crée
	// if err != nil {
	// 	fmt.Println("Write in CA RSA Key file error : ", err)
	// 	return err
	// }

	// création du certificat publique de la CA
	// var expires string
	// if years := 10; years != 0 {
	// 	expires = fmt.Sprintf("%s %d years", expires, years)
	// }

	// expiresTime, err := parseExpiry(expires)
	// if err != nil {
	// 	fmt.Println("Invalid expiry : ", err)
	// 	return err
	// }
	// CAcert, err := pkix.CreateCertificateAuthority(CAkey, "ditrit", expiresTime, "ditrit", "France", "", "Paris", "CA") // à voir pour des variables d'environnement plus tard
	// if err != nil {
	// 	fmt.Println("Create CA certificate error : ", err)
	// 	return err
	// }

	// CApublicCert, err := CAcert.Export()
	// if err != nil {
	// 	fmt.Println("Export CA certificate error : ", err)
	// 	return err
	// }

	// CAcertFile, err := os.Create(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/CAcert.pem")
	// if err != nil {
	// 	fmt.Println("Create CA certificate file error : ", err)
	// 	return err
	// }

	// _, err = CAcertFile.Write(CApublicCert)
	// if err != nil {
	// 	fmt.Println("Write in CA certificate file error : ", err)
	// 	return err
	// }

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
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

	// 2. Le certificat de la shoset
	// récupérer le certificat de la CA
	// génération des clefs privée, publique et request pour la shoset
	// hostKey := c.CreateKey()
	// création du certificat signé avec la clef privée de la CA
	// hostCsr := c.CreateSignRequest(hostKey) // demande de signature
	// c.SignRequest(CAcert, hostCsr, hostKey) // signature

	CAprivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)                                               // private key rsa format
	CApublicKey := &CAprivateKey.PublicKey                                                              // we extract the public key (CA cert) from the private key
	signedCAcert, err := x509.CreateCertificate(rand.Reader, CAcert, CAcert, CApublicKey, CAprivateKey) // we sign the certificate
	if err != nil {
		return errors.New("couldn't create CA")
	}

	// Public key
	CAcertFile, err := os.Create(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/CAcert.crt")
	if err != nil {
		return err
	}

	pem.Encode(CAcertFile, &pem.Block{Type: "CERTIFICATE", Bytes: signedCAcert})
	CAcertFile.Close()
	// log.Print("written cert.pem\n")

	// Private key
	CAprivateKeyFile, err := os.OpenFile(dirname+"/.shoset/"+c.ConnsByName.GetConfigName()+"/cert/privateCAKey.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	pem.Encode(CAprivateKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(CAprivateKey)})
	CAprivateKeyFile.Close()
	// log.Print("written key.pem\n")

	// Create and sign additional certificates - here the certificate of the socket from the CA
	c.InitCertificate(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/CAcert.crt", dirname+"/.shoset/"+c.ConnsByName.GetConfigName()+"/cert/privateCAKey.key")

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

// // getCAcert() => { certificat de la CA }
// func (c *Shoset) GetCAcert() *pkix.Certificate {
// 	if c.GetIsPki() {
// 		dirname, err := os.UserHomeDir()
// 		if err != nil {
// 			fmt.Println("Get UserHomeDir error : ", err)
// 		}
// 		cert, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/CAcert.pem")
// 		if err == nil {
// 			CAcert, err := pkix.NewCertificateFromPEM(cert)
// 			if err == nil {
// 				return CAcert
// 			}
// 		}
// 	}
// 	return nil
// }

// // getCert(certRequest) => { certificat }
// func (c *Shoset) GetCert() []byte {
// 	if c.GetIsPki() {
// 		dirname, err := os.UserHomeDir()
// 		if err != nil {
// 			fmt.Println("Get UserHomeDir error : ", err)
// 		}
// 		cert, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/cert.pem")
// 		if err == nil {
// 			return cert
// 		}
// 	}
// 	return nil
// }

func (c *Shoset) InitCertificate(certFile, keyFile string) error {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Load CA
	catls, err := tls.LoadX509KeyPair(certFile, keyFile)
	// catls, err := tls.LoadX509KeyPair(dirname+"/.shoset/127-0-0-1_8001/cert/CAcert.crt", dirname+"/.shoset/127-0-0-1_8001/cert/privateCAKey.key")

	if err != nil {
		fmt.Println("can't load pair :", err)
		return err
	} else {
		fmt.Println("CA loaded too")
	}
	ca, err := x509.ParseCertificate(catls.Certificate[0]) // we parse the previous certificate
	if err != nil {
		return err
	}

	// Prepare new certificate
	cert := &x509.Certificate{
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

	hostPrivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	hostPublicKey := &hostPrivateKey.PublicKey

	// Sign the certificate
	signedHostCert, err := x509.CreateCertificate(rand.Reader, cert, ca, hostPublicKey, catls.PrivateKey)
	if err != nil {
		return err
	}

	// Public key
	hostCertFile, err := os.Create(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/cert.crt")
	if err != nil {
		return err
	}
	pem.Encode(hostCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: signedHostCert})
	hostCertFile.Close()
	// log.Print("written cert.pem\n")

	// Private key
	hostPrivateKeyFile, err := os.OpenFile(dirname+"/.shoset/"+c.ConnsByName.GetConfigName()+"/cert/privateKey.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	pem.Encode(hostPrivateKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(hostPrivateKey)})
	hostPrivateKeyFile.Close()
	// log.Print("written key.pem\n")
	return nil
}

func (c *Shoset) CreateKey() (*rsa.PublicKey, *rsa.PrivateKey) {
	hostPrivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	hostPublicKey := &hostPrivateKey.PublicKey

	// key, err := pkix.CreateRSAKey(4096)
	// if err != nil {
	// 	fmt.Println("Create RSA Key error:", err)
	// 	return nil
	// }

	// privateKey, err := key.ExportPrivate()
	// if err != nil {
	// 	fmt.Println("Export RSA Key error : ", err)
	// 	return nil
	// }
	// dirname, err := os.UserHomeDir()
	// if err != nil {
	// 	fmt.Println("Get UserHomeDir error : ", err)
	// 	return nil
	// }
	// privateKeyFile, err := os.Create(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/privateKey.pem")
	// if err != nil {
	// 	fmt.Println("Create RSA Key file error : ", err)
	// 	return nil
	// }
	// _, err = privateKeyFile.Write(privateKey)
	// if err != nil {
	// 	fmt.Println("Write in RSA Key file error : ", err)
	// 	return nil
	// }
	return hostPublicKey, hostPrivateKey
}

// func (c *Shoset) CreateSignRequest(hostKey *pkix.Key) *pkix.CertificateSigningRequest {
// 	//We sign the certificate request
// 	hostCsr, err := pkix.CreateCertificateSigningRequest(hostKey, "ditrit", nil, nil, nil, "ditrit", "France", "", "Paris", "csr")
// 	// hostCsr, err := pkix.NewCertificateSigningRequestFromPEM(hostKeyBytes)
// 	if err != nil {
// 		fmt.Println("Create Sign Request error : ", err)
// 		return nil
// 	}
// 	return hostCsr
// }

// func (c *Shoset) SignRequest(CAcert *pkix.Certificate, hostCsr *pkix.CertificateSigningRequest, hostKey *pkix.Key) {
// 	expire_time, _ := time.Parse("020106 150405", "220902 050316")
// 	hostCert, err := pkix.CreateCertificateHost(CAcert, hostKey, hostCsr, expire_time)
// 	if err != nil {
// 		fmt.Println("Sign Request error : ", err)
// 	}

// 	hostCertBytes, err := hostCert.Export()
// 	if err != nil {
// 		fmt.Println("Export hostCert error : ", err)
// 	}
// 	dirname, err := os.UserHomeDir()
// 	if err != nil {
// 		fmt.Println("Get UserHomeDir error : ", err)
// 	}
// 	hostCertFile, err := os.Create(dirname + "/.shoset/" + c.ConnsByName.GetConfigName() + "/cert/cert.pem")
// 	if err != nil {
// 		fmt.Println("Create hostCert file error : ", err)
// 	}

// 	_, err = hostCertFile.Write(hostCertBytes)
// 	if err != nil {
// 		fmt.Println("Write in hostCert file error : ", err)
// 	}
// }
