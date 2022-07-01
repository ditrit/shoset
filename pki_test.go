package shoset

import (
	"crypto/rsa"
	"crypto/x509"
	"testing"
)

var certificateRequest *x509.Certificate
var hostPublicKey *rsa.PublicKey
var hostPrivateKey *rsa.PrivateKey
var err error

// TestPrepareCertificate verifies if PrepareCertificate() function returns expected certificateRequest, hostPublicKey and hostPrivateKey.
func TestPrepareCertificate(t *testing.T) {
	certificateRequest, hostPublicKey, hostPrivateKey, err = PrepareCertificate()
	if certificateRequest == nil {
		t.Errorf("certificateRequest is not valid")
	}
	if hostPublicKey == nil {
		t.Errorf("hostPublicKey is not valid")
	}
	if hostPrivateKey == nil {
		t.Errorf("hostPrivateKey is not valid")
	}
	if err != nil {
		t.Errorf("unexpected error : %s", err)
	}
}

// TestSignCertificate verifies if SignCertificate() function returns a correct signedCertificate.
func TestSignCertificate(t *testing.T) {
	shoset := NewShoset("cl", "cl") // cluster
	shoset.InitPKI("localhost:8001")

	TestPrepareCertificate(t)

	signedCertificate := shoset.SignCertificate(certificateRequest, hostPublicKey)
	if signedCertificate == nil {
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
