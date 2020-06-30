package msg

import (
	"fmt"
)

// CertRequest : for signing req at startup (if needed)
type CertRequest struct {
	MessageBase
	// status string
	cn    string
	pub   []byte
	cert  []byte
	chain []string
}

func (c *CertRequest) String() string {
	if c == nil {
		fmt.Printf("\nError : *CertRequest.String : nil\n")
	}
	return fmt.Sprintf("[ cn: %s, pub: %v,\n", c.cn, c.pub)
}

// GetMsgType accessor
func (c CertRequest) GetMsgType() string { return "csr" }

// // Status :
// func (c CertRequest) Status() string { return c.status }

// // SetReqStatus :
// func (c CertRequest) SetReqStatus() { c.status = "req" }

// // SetRespStatus :
// func (c CertRequest) SetRespStatus() { c.status = "resp" }

// CN :
func (c CertRequest) CN() string { return c.cn }

// Pub :
func (c CertRequest) Pub() []byte { return c.pub }

// Cert :
func (c CertRequest) Cert() []byte { return c.cert }

// SetCert :
func (c CertRequest) SetCert(cert []byte) { c.cert = cert }

// ChainAppend :
func (c CertRequest) ChainAppend(addr string) { c.chain = append(c.chain, addr) }

// ChainGet :
func (c CertRequest) ChainGet(i int) string {
	l := len(c.chain)
	if i >= l || i < -l {
		return ""
	}
	if i < 0 {
		return c.chain[l+i]
	}
	return c.chain[i]
}

// ChainPopLast :
func (c CertRequest) ChainPopLast() (addr string) {
	i := len(c.chain)
	addr = c.chain[i-1]
	c.chain[i-1] = ""
	return
}

// NewCertRequest :
func NewCertRequest(cn, bindAddr string, pubKey []byte) (csr *CertRequest) {
	// csr.status = "req"
	csr.cn = cn
	csr.pub = pubKey
	csr.chain = []string{bindAddr}
	return
}
