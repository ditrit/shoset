package shoset

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// InitPKI inits the concerned Shoset to get the role PKI which is basically admin of network.
// First inits CA role and then inits Shoset itself.
// Overload of shoset.Bind() function
func (s *Shoset) InitPKI(bindAddress string) {
	s.InitCA(bindAddress)
	s.InitShoset(bindAddress)
}

// InitCA inits CA role for the concerned Shoset, who then become admin of network.
func (s *Shoset) InitCA(bindAddress string) {
	if s.GetIsPki() {
		s.Logger.Error().Msg("shoset is already pki")
		return
	}
	s.SetIsPki(true)

	ipAddress, err := GetIP(bindAddress)
	if err != nil {
		s.Logger.Error().Msg("wrong IP format : " + err.Error())
		return
	}
	formattedIpAddress := strings.Replace(ipAddress, ":", "_", -1)
	formattedIpAddress = strings.Replace(formattedIpAddress, ".", "-", -1) // formats for filesystem to 127-0-0-1_8001 instead of 127.0.0.1:8001
	s.ConnsByLname.GetConfig().SetFileName(formattedIpAddress)

	cfgDir, err := s.ConnsByLname.GetConfig().InitFolders(formattedIpAddress)
	if err != nil {
		s.Logger.Error().Msg("couldn't get dirname : " + err.Error())
		return
	}
	fileName := s.ConnsByLname.GetConfig().GetFileName()
	CApath := cfgDir + fileName + PATH_CA_CERT

	CAcertificate := &x509.Certificate{
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

	CAprivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	err = EncodeFile(CAprivateKey, RSA_PRIVATE_KEY, cfgDir+fileName+PATH_CA_PRIVATE_KEY)
	if err != nil {
		s.Logger.Error().Msg(err.Error())
		return
	}

	CApublicKey := &CAprivateKey.PublicKey

	signedCAcert, err := x509.CreateCertificate(rand.Reader, CAcertificate, CAcertificate, CApublicKey, CAprivateKey)
	if err != nil {
		s.Logger.Error().Msg("couldn't create CA : " + err.Error())
		return
	}

	err = EncodeFile(signedCAcert, CERTIFICATE, CApath)
	if err != nil {
		s.Logger.Error().Msg(err.Error())
		return
	}
}

// InitShoset inits certs for the Shoset with the previous created CA.
func (s *Shoset) InitShoset(bindAddress string) {
	fileName := s.ConnsByLname.GetConfig().GetFileName()
	cfgDir := s.ConnsByLname.GetConfig().GetBaseDir()
	CApath := cfgDir + fileName + PATH_CA_CERT

	certificateRequest, hostPublicKey, hostPrivateKey, err := PrepareCertificate()
	if err != nil {
		s.Logger.Error().Msg("prepare certificate didn't work")
		return
	}
	err = EncodeFile(hostPrivateKey, RSA_PRIVATE_KEY, s.ConnsByLname.GetConfig().GetBaseDir()+s.ConnsByLname.GetConfig().GetFileName()+PATH_PRIVATE_KEY)
	if err != nil {
		s.Logger.Error().Msg(err.Error())
		return
	}

	signedHostCert := s.SignCertificate(certificateRequest, hostPublicKey)
	if signedHostCert == nil {
		s.Logger.Error().Msg("sign cert didn't work")
		return
	}
	err = EncodeFile(signedHostCert, CERTIFICATE, cfgDir+fileName+PATH_CERT)
	if err != nil {
		s.Logger.Error().Msg(err.Error())
		return
	}

	os.Setenv(CERT_FILE_ENVIRONMENT, CApath)

	err = s.SetUpDoubleWay()
	if err != nil {
		s.Logger.Error().Msg(err.Error())
		return
	}

	err = s.Bind(bindAddress)
	if err != nil {
		s.Logger.Error().Msg("couldn't set bindAddress : " + err.Error())
		return
	}
}

// GenerateSecret generates a secret based on a login and a password that the admin of the system will have created.
// Must be PKI to generate a secret.
func (s *Shoset) GenerateSecret(login, password string) string {
	if s.GetIsPki() {
		// use login and password.
		return uuid.New().String()
	}
	return VOID
}

// PrepareCertificate prepares certificates for a Shoset by returning certificateRequest, hostPublicKey and hostPrivateKey.
// To get more info about the generated cert, use : openssl x509 -in ./cert.crt -text -noout
func PrepareCertificate() (*x509.Certificate, *rsa.PublicKey, *rsa.PrivateKey, error) {
	certificateRequest := &x509.Certificate{
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

	hostPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}
	hostPublicKey := &hostPrivateKey.PublicKey

	return certificateRequest, hostPublicKey, hostPrivateKey, nil
}

// SignCertificate signs a certificate request certificate with a public key and the the CA information.
// Must be PKI to sign a certificate request.
// To check the validity of the signed certificate request, use : openssl s_server -accept 8080 -www -cert yourcert.crt -key yourcert.key -CAfile CAcert.crt
func (s *Shoset) SignCertificate(certificateRequest *x509.Certificate, hostPublicKey *rsa.PublicKey) []byte {
	if !s.GetIsPki() {
		return nil
	}

	filePath := s.ConnsByLname.GetConfig().GetBaseDir() + s.ConnsByLname.GetConfig().GetFileName()
	loadedCAkeys, err := tls.LoadX509KeyPair(filePath+PATH_CA_CERT, filePath+PATH_CA_PRIVATE_KEY)
	if err != nil {
		s.Logger.Error().Msg("couldn't load keyPair : " + err.Error())
		return nil
	}

	parsedCAcert, err := x509.ParseCertificate(loadedCAkeys.Certificate[0])
	if err != nil {
		s.Logger.Error().Msg("couldn't parse cert : " + err.Error())
		return nil
	}

	signedHostCert, err := x509.CreateCertificate(rand.Reader, certificateRequest, parsedCAcert, hostPublicKey, loadedCAkeys.PrivateKey)
	if err != nil {
		s.Logger.Error().Msg("couldn't sign certificateRequest : " + err.Error())
		return nil
	}
	return signedHostCert
}

// SetUpDoubleWay sets up the tls config once the certificate is signed.
// Sets up Double Way for future secured exchanges.
// Updates Single Way for future exchanges with non-certified Shoset.
func (s *Shoset) SetUpDoubleWay() error {
	fileName := s.ConnsByLname.GetConfig().GetFileName()
	cfgDir := s.ConnsByLname.GetConfig().GetBaseDir()
	CApath := cfgDir + fileName + PATH_CA_CERT

	loadedCAkeys, err := tls.LoadX509KeyPair(cfgDir+fileName+PATH_CERT, cfgDir+fileName+PATH_PRIVATE_KEY)
	if err != nil {
		s.Logger.Error().Msg("Unable to Load certificate : " + err.Error())
		return errors.New("Unable to Load certificate : " + err.Error())
	}

	CAcertBytes, err := ioutil.ReadFile(CApath)
	if err != nil {
		s.Logger.Error().Msg("error read file cacert : " + err.Error())
		return errors.New("error read file cacert : " + err.Error())
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(CAcertBytes)

	s.tlsConfigDoubleWay = &tls.Config{
		Certificates:       []tls.Certificate{loadedCAkeys},
		ClientCAs:          caCertPool,
		ClientAuth:         tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: true, // true for reconnection to work (but not as secure)
		//Renegotiation: tls.RenegotiateOnceAsClient,
	}

	s.tlsConfigSingleWay = &tls.Config{
		Certificates:       []tls.Certificate{loadedCAkeys},
		InsecureSkipVerify: false,
	}
	return nil
}

// Certify runs a certification request with the CA.
func (s *Shoset) Certify(bindAddress, remoteAddress string) error {
	certRequestConn, err := NewShosetConn(s, remoteAddress, OUT)
	if err != nil {
		s.Logger.Error().Msg("couldn't create shoset : " + err.Error())
		return err
	}

	err = certRequestConn.RunPkiRequest(bindAddress)
	if err != nil {
		s.Logger.Error().Msg("RunPkiRequest didn't work" + err.Error())
		return err
	}
	s.Logger.Debug().Msg("shoset certified")
	return nil
}
