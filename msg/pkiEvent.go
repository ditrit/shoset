package msg

import (
	"crypto/rsa"
	"crypto/x509"
)

// PkiEvent : Event dedicated to PKI initialization.
type PkiEvent struct {
	MessageBase
	RequestAddress     string            // requesting address for certification
	Command            string            // type of event
	Secret             string            // secret shown to CA to authenticate
	LogicalName        string            // logical name of the Shoset
	CertificateRequest *x509.Certificate // certificate to be signed by the CA
	SignedCertificate  []byte            // certificate signed by the CA
	HostPublicKey      *rsa.PublicKey    // public key of the requesting Shoset
	CAprivateKey       *rsa.PrivateKey   // private key of the responding CA
	CAcertificate      []byte            // certificate of the CA
}

// NewPkiEventInit creates a new init PkiEvent object.
// Initializes each fields and message base.
func NewPkiEventInit(command, requestAddress, logicalName string, certificateRequest *x509.Certificate, hostPublicKey *rsa.PublicKey) *PkiEvent {
	e := new(PkiEvent)
	e.InitMessageBase()

	e.RequestAddress = requestAddress
	e.Command = command
	e.CertificateRequest = certificateRequest
	e.HostPublicKey = hostPublicKey
	e.LogicalName = logicalName
	return e
}

// NewPkiEventReturn creates a new return PkiEvent object.
// Initializes each fields and message base.
func NewPkiEventReturn(command, requestAddress string, signedCertificate, CAcertificate []byte, caPrivateKey *rsa.PrivateKey) *PkiEvent {
	e := new(PkiEvent)
	e.InitMessageBase()

	e.Command = command
	e.RequestAddress = requestAddress
	e.SignedCertificate = signedCertificate
	e.CAcertificate = CAcertificate
	e.CAprivateKey = caPrivateKey
	return e
}

// GetMessageType returns MessageType from PkiEvent.
func (e PkiEvent) GetMessageType() string {
	switch e.GetCommand() {
	case "pkievt_TLSsingleWay":
		return "pkievt_TLSsingleWay"
	case "return_pkievt_TLSsingleWay":
		return "pkievt_TLSsingleWay"
	case "pkievt_TLSdoubleWay":
		return "pkievt_TLSdoubleWay"
	case "return_pkievt_TLSdoubleWay":
		return "pkievt_TLSdoubleWay"
	}
	return "Wrong input protocolType"
}

// GetSecret returns Secret from PkiEvent.
func (e PkiEvent) GetSecret() string { return e.Secret }

// GetCommand returns Command from PkiEvent.
func (e PkiEvent) GetCommand() string { return e.Command }

// SetCommand sets Command for PkiEvent.
func (e *PkiEvent) SetCommand(command string) { e.Command = command }

// GetRequestAddress returns RequestAddress from PkiEvent.
func (e PkiEvent) GetRequestAddress() string { return e.RequestAddress }

// GetLogicalName returns LogicalName from PkiEvent.
func (e PkiEvent) GetLogicalName() string { return e.LogicalName }

// GetCertRequest returns CertificateRequest from PkiEvent.
func (e PkiEvent) GetCertificateRequest() *x509.Certificate { return e.CertificateRequest }

// GetSignedCert returns SignedCertificate from PkiEvent.
func (e PkiEvent) GetSignedCert() []byte { return e.SignedCertificate }

// GetHostPublicKey returns HostPublicKey from PkiEvent.
func (e PkiEvent) GetHostPublicKey() *rsa.PublicKey { return e.HostPublicKey }

// GetCAprivateKey returns CAprivateKey from PkiEvent.
func (e PkiEvent) GetCAprivateKey() *rsa.PrivateKey { return e.CAprivateKey }

// GetCAcert returns CAcertificate from PkiEvent.
func (e PkiEvent) GetCAcert() []byte { return e.CAcertificate }
