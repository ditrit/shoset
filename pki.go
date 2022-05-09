package shoset

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"net"

	// "io/ioutil"
	"math/big"
	"os"

	"strings"
	"time"

	"github.com/google/uuid"
)

func (c *Shoset) InitPKI(address string) {
	if c.GetIsPki() { // il y a eu déjà un init ou j'ai déjà un certificat (mon certificat existe déjà)
		c.logger.Error().Msg("shoset is already pki")
		return
	}
	c.SetIsPki(true) // je prends le role de CA de la PKI

	// demarche d'initialisation de bind classique (shoset.go/bind)
	ipAddress, err := GetIP(address) // parse the address from function parameter to get the IP
	if err != nil {
		// IP nok -> return early
		c.logger.Error().Msg("wrong IP format : " + err.Error())
		return
	}

	formatedIpAddress := strings.Replace(ipAddress, ":", "_", -1)
	formatedIpAddress = strings.Replace(formatedIpAddress, ".", "-", -1)
	c.SetFileName(formatedIpAddress)

	cfgDir, err := c.config.InitFolders(formatedIpAddress)
	if err != nil { // initialization of folders did not work
		c.logger.Error().Msg("couldn't get dirname : " + err.Error())
		return
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
		c.logger.Error().Msg("couldn't create CA : " + err.Error())
		return
	}

	fname := c.GetFileName()
	CAcertFile, err := os.Create(cfgDir + fname + "/cert/CAcert.crt")
	if err != nil {
		c.logger.Error().Msg("couldn't create CAcertFile : " + err.Error())
		return
	}
	err = pem.Encode(CAcertFile, &pem.Block{Type: "CERTIFICATE", Bytes: signedCAcert})
	if err != nil {
		c.logger.Error().Msg("couldn't encode in file : " + err.Error())
		return
	}
	CAcertFile.Close()

	// Private key
	CAprivateKeyFile, err := os.OpenFile(cfgDir+fname+"/cert/privateCAKey.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		c.logger.Error().Msg("couldn't open CAprivateKeyfile : " + err.Error())
		return
	}
	err = pem.Encode(CAprivateKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(CAprivateKey)})
	if err != nil {
		c.logger.Error().Msg("couldn't encode in CAprivateKeyfile : " + err.Error())
		return
	}
	CAprivateKeyFile.Close()

	// Public key
	// Create and sign additional certificates - here the certificate of the socket from the CA
	certReq, hostPublicKey, _ := c.PrepareCertificate()
	if certReq == nil || hostPublicKey == nil {
		c.logger.Error().Msg("prepare certificate didn't work")
		return
	}

	signedHostCert := c.SignCertificate(certReq, hostPublicKey)
	if signedHostCert == nil {
		c.logger.Error().Msg("dign cert didn't work")
		return
	}

	certFile, err := os.Create(cfgDir + fname + "/cert/cert.crt")
	if err != nil {
		c.logger.Error().Msg("couldn't create certFile : " + err.Error())
		return
	}
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: signedHostCert})
	if err != nil {
		c.logger.Error().Msg("couldn't encode in file : " + err.Error())
		return
	}
	certFile.Close()

	// 3. Elle associe le rôle 'pki' au nom logique de la shoset
	c.SetIsCertified(true)

	// point env variable to our CAcert so that computer does not point elsewhere
	os.Setenv("SSL_CERT_FILE", cfgDir+fname+"/cert/CAcert.crt")

	// tls Double way
	cert, err := tls.LoadX509KeyPair(cfgDir+fname+"/cert/cert.crt", cfgDir+fname+"/cert/privateKey.key")
	if err != nil {
		c.logger.Error().Msg("Unable to Load certificate : " + err.Error())
		return
	}
	CAcertBytes, err := ioutil.ReadFile(cfgDir + fname + "/cert/CAcert.crt")
	if err != nil {
		c.logger.Error().Msg("error read file cacert : " + err.Error())
		return
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(CAcertBytes)
	c.tlsConfigDoubleWay = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		ClientCAs:          caCertPool,
		ClientAuth:         tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: false,
	}
	// c.tlsConfigDoubleWay.BuildNameToCertificate()

	// tls config single way
	c.tlsConfigSingleWay = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: false, // peut etre true
	}

	if c.GetBindAddress() == "" {
		err := c.Bind(address) // I have my certs, I can bind
		if err != nil {
			c.logger.Error().Msg("couldn't set bindAddress : " + err.Error())
			return
		}
	}
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
	// Prepare new certificate
	// voir infos du certificat généré
	// openssl x509 -in ./cert.crt -text -noout
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
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Private and public keys
	hostPrivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	hostPublicKey := &hostPrivateKey.PublicKey

	// Private key
	hostPrivateKeyFile, err := os.OpenFile(c.GetConfigDir()+c.GetFileName()+"/cert/privateKey.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		c.logger.Error().Msg("error open hostPrivateKeyFile : " + err.Error())
		return nil, nil, nil
	}
	err = pem.Encode(hostPrivateKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(hostPrivateKey)})
	if err != nil {
		c.logger.Error().Msg("couldn't encode in file : " + err.Error())
		return nil, nil, nil
	}
	hostPrivateKeyFile.Close()
	return certReq, hostPublicKey, hostPrivateKey
}

func (c *Shoset) SignCertificate(certReq *x509.Certificate, hostPublicKey *rsa.PublicKey) []byte {
	// check if the certificates generated are valid
	// openssl s_server -accept 8080 -www -cert yourcert.crt -key yourcert.key -CAfile CAcert.crt

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.GetIsPki() {
		return nil
	}

	// Load CA
	fileDir := c.GetConfigDir() + c.GetFileName() + "/cert"
	catls, err := tls.LoadX509KeyPair(fileDir+"/CAcert.crt", fileDir+"/privateCAKey.key")
	if err != nil {
		c.logger.Error().Msg("couldn't load keypair : " + err.Error())
		return nil
	}

	ca, err := x509.ParseCertificate(catls.Certificate[0]) // we parse the previous certificate
	if err != nil {
		c.logger.Error().Msg("couldn't parse cert : " + err.Error())
		return nil
	}

	// Sign the certificate
	signedHostCert, err := x509.CreateCertificate(rand.Reader, certReq, ca, hostPublicKey, catls.PrivateKey)
	if err != nil {
		c.logger.Error().Msg("couldn't sign certreq : " + err.Error())
		return nil
	}
	return signedHostCert
}
