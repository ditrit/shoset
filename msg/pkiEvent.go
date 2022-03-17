package msg

import (
	"crypto/rsa"
	"crypto/x509"
)

// Event : Gandalf event internal
type PkiEvent struct {
	MessageBase
	RequestAddress string
	Command string
	Secret string
	LogicalName string
	CertReq *x509.Certificate
	SignedCert []byte
	HostPublicKey *rsa.PublicKey
	CAprivateKey []byte
	CAcert []byte
}

func NewPkiEventInit(command, requestAddress, logicalName string, certReq *x509.Certificate, hostPublicKey *rsa.PublicKey) *PkiEvent{
	e := new(PkiEvent)
	e.InitMessageBase()

	e.RequestAddress = requestAddress
	e.Command = command
	e.CertReq = certReq
	e.HostPublicKey = hostPublicKey
	e.LogicalName = logicalName
	return e
}

func NewPkiEventReturn(requestAddress string, signedCert, CAcert []byte, caPrivateKey []byte) *PkiEvent{
	e := new(PkiEvent)
	e.InitMessageBase()

	e.RequestAddress = requestAddress
	e.SignedCert = signedCert
	e.CAcert = CAcert
	e.CAprivateKey = caPrivateKey
	return e
}

func (e PkiEvent) GetMsgType() string { return "pkievt" }

func (e PkiEvent) GetSecret() string { return e.Secret }

func (e PkiEvent) GetCommand() string { return e.Command }

func (e PkiEvent) GetRequestAddress() string { return e.RequestAddress }

func (e PkiEvent) GetLogicalName() string { return e.LogicalName }

func (e PkiEvent) GetCertReq() *x509.Certificate {return e.CertReq}

func (e PkiEvent) GetSignedCert() []byte {return e.SignedCert}

func (e PkiEvent) GetHostPublicKey() *rsa.PublicKey { return e.HostPublicKey }

func (e PkiEvent) GetCAprivateKey() []byte { return e.CAprivateKey }

func (e PkiEvent) GetCAcert() []byte {return e.CAcert}


