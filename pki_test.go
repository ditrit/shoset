package shoset

import (
	"crypto/rsa"
	"crypto/x509"
	"testing"
)

var certReq *x509.Certificate
var hostPublicKey *rsa.PublicKey
var hostPrivateKey *rsa.PrivateKey
var err error

// TestPrepareCertificate verifies if PrepareCertificate() function returns expected certReq, hostPublicKey and hostPrivateKey.
func TestPrepareCertificate(t *testing.T) {
	certReq, hostPublicKey, hostPrivateKey, err = PrepareCertificate()
	if certReq == nil {
		t.Errorf("certReq is not valid")
	}
	if hostPublicKey == nil {
		t.Errorf("hostPublicKey is not valid")
	}
	if hostPrivateKey == nil {
		t.Errorf("hostPrivateKey is not valid")
	}
	if err != nil {
		t.Errorf("unexepected error : %s", err)
	}
}

// TestSignCertificate verifies if SignCertificate() function returns a correct signedCert.
func TestSignCertificate(t *testing.T) {
	shoset := NewShoset("cl", "cl") // cluster
	shoset.InitPKI("localhost:8001")

	TestPrepareCertificate(t)

	signedCert := shoset.SignCertificate(certReq, hostPublicKey)
	if signedCert == nil {
		t.Errorf("TestSignCertificate didn't work")
	}
}

// TestGenerateSecret verifies if GenerateSecret() returns a correct secret.
func TestGenerateSecret(t *testing.T) {
	shoset := NewShoset("cl", "cl") // cluster
	shoset.InitPKI("localhost:8001")

	secret := shoset.GenerateSecret(VOID, VOID)
	if secret == VOID {
		t.Errorf("TestGenerateSecret didn't work")
	}
}
