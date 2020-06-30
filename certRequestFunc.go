package shoset

import (
	"fmt"

	"github.com/ditrit/shoset/msg"
)

// GetCertRequest :
func GetCertRequest(c *ShosetConn) (msg.Message, error) {
	var csr msg.CertRequest
	err := c.ReadMessage(&csr)
	return csr, err
}

// HandleCertRequest :
func HandleCertRequest(c *ShosetConn, message msg.Message) error {
	csr := message.(msg.CertRequest)
	sh := c.GetCh()
	dir := c.GetDir()

	switch {
	case dir == "in":
		if c.GetBindAddr() == "" {
			c.SetBindAddr(csr.ChainGet(-1))
			sh.SetConn(c.GetBindAddr(), c.GetShosetType(), c)
		}

		if sh.canSignCert {
			cert, err := signCert(csr.CN(), csr.Pub(), sh.certs["ca"], sh.certs["cakey"])
			if err != nil {
				return fmt.Errorf("HandleCertRequest: could not sign csr : %v", err)
			}
			csr.SetCert(cert)
			c.SendMessage(csr)
		} else {
			var np1 *ShosetConn
			for _, v := range sh.ConnsByAddr.m {
				if d := v.GetDir(); d == "out" {
					np1 = v
					break
				}
			}
			if np1 == nil {
				return fmt.Errorf("HandleCertRequest: could not find higher ranking shoset")
			}
			csr.ChainAppend(c.bindAddr)
			np1.SendMessage(csr)
		}
	case dir == "out":
		if csr.ChainGet(0) == "" {
			sh.certs["cert"] = csr.Cert()
			sh.reloadTLSConfig()
			c.socket.Close()
		} else {
			addr := csr.ChainPopLast()
			conn := sh.ConnsByAddr.Get(addr)
			conn.SendMessage(csr)
		}
	}
	return nil
}
