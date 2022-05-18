package shoset

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
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

	// 3. Elle associe le rôle 'pki' au nom logique de la shoset
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
	c.config.SetFileName(formatedIpAddress)

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
	fname := c.config.GetFileName()

	// CA private key
	CAprivateKey, _ := rsa.GenerateKey(rand.Reader, 2048) // private key rsa format
	err = EncodeFile(CAprivateKey, RSA_PRIVATE_KEY, cfgDir+fname+PATH_CA_PRIVATE_KEY)
	if err != nil {
		c.logger.Error().Msg(err.Error())
		return
	}

	// CA public key
	CApublicKey := &CAprivateKey.PublicKey                                                              // we extract the public key (CA cert) from the private key
	signedCAcert, err := x509.CreateCertificate(rand.Reader, CAcert, CAcert, CApublicKey, CAprivateKey) // we sign the certificate
	if err != nil {
		c.logger.Error().Msg("couldn't create CA : " + err.Error())
		return
	}

	CApath := cfgDir + fname + PATH_CA_CERT
	err = EncodeFile(signedCAcert, CERTIFICATE, CApath)
	if err != nil {
		c.logger.Error().Msg(err.Error())
		return
	}

	// Public key
	// Create and sign additional certificates - here the certificate of the socket from the CA
	certReq, hostPublicKey, hostPrivateKey, err := PrepareCertificate()
	if err != nil {
		c.logger.Error().Msg("prepare certificate didn't work")
		return
	}

	// Private key
	err = EncodeFile(hostPrivateKey, RSA_PRIVATE_KEY, c.config.GetBaseDir()+c.config.GetFileName()+PATH_PRIVATE_KEY)
	if err != nil {
		c.logger.Error().Msg(err.Error())
		return
	}

	signedHostCert := c.SignCertificate(certReq, hostPublicKey)
	if signedHostCert == nil {
		c.logger.Error().Msg("sign cert didn't work")
		return
	}

	err = EncodeFile(signedHostCert, CERTIFICATE, cfgDir+fname+PATH_CERT)
	if err != nil {
		c.logger.Error().Msg(err.Error())
		return
	}

	// point env variable to our CAcert so that computer does not point elsewhere
	os.Setenv(CERT_FILE_ENV, CApath)

	// tls Double way
	err = c.SetUpDoubleWay(cfgDir, fname, CApath)
	if err != nil {
		c.logger.Error().Msg(err.Error())
		return
	}

	if c.GetBindAddress() == VOID {
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
	return VOID
}

func PrepareCertificate() (*x509.Certificate, *rsa.PublicKey, *rsa.PrivateKey, error) {
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
		IPAddresses:  []net.IP{net.ParseIP(LOCALHOST)},
	}

	// Private and public keys
	hostPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}
	hostPublicKey := &hostPrivateKey.PublicKey

	return certReq, hostPublicKey, hostPrivateKey, nil
}

func (c *Shoset) SignCertificate(certReq *x509.Certificate, hostPublicKey *rsa.PublicKey) []byte {
	// check if the certificates generated are valid
	// openssl s_server -accept 8080 -www -cert yourcert.crt -key yourcert.key -CAfile CAcert.crt

	if !c.GetIsPki() {
		return nil
	}

	// Load CA
	fileDir := c.config.GetBaseDir() + c.config.GetFileName()
	catls, err := tls.LoadX509KeyPair(fileDir+PATH_CA_CERT, fileDir+PATH_CA_PRIVATE_KEY)
	if err != nil {
		c.logger.Error().Msg("couldn't load keypair : " + err.Error())
		return nil
	}

	ca, err := x509.ParseCertificate(catls.Certificate[0])
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

func (c *Shoset) SetUpDoubleWay(cfgDir, fname, CApath string) error {
	cert, err := tls.LoadX509KeyPair(cfgDir+fname+PATH_CERT, cfgDir+fname+PATH_PRIVATE_KEY)
	if err != nil {
		c.logger.Error().Msg("Unable to Load certificate : " + err.Error())
		return errors.New("Unable to Load certificate : " + err.Error())
	}
	CAcertBytes, err := ioutil.ReadFile(CApath)
	if err != nil {
		c.logger.Error().Msg("error read file cacert : " + err.Error())
		return errors.New("error read file cacert : " + err.Error())
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(CAcertBytes)
	c.tlsConfigDoubleWay = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		ClientCAs:          caCertPool,
		ClientAuth:         tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: false,
	}

	c.tlsConfigSingleWay = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: false,
	}
	return nil
}

func (c *Shoset) Certify(bindAddress, remoteAddress string) error {
	initConn, err := NewShosetConn(c, remoteAddress, OUT)
	if err != nil {
		c.logger.Error().Msg("couldn't create shoset : " + err.Error())
		return err
	}

	err = initConn.runPkiRequest(bindAddress) // I don't have my certs, I request them
	if err != nil {
		c.logger.Error().Msg("runPkiRequest didn't work" + err.Error())
		return err
	}
	c.logger.Debug().Msg("shoset certified")
	return nil
}
