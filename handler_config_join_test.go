package shoset

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/ditrit/shoset/msg"
)

const (
	s9001 = "127.0.0.1:9001"
	s9002 = "127.0.0.1:9002"
)

func TestHandleDoubleWay(t *testing.T) {
	shoset.InitPKI(s9001)

	shoset2 := NewShoset("cl", "cl")
	shoset2.ConnsByLname.GetConfig().SetFileName("127-0-0-1_9002")
	_, err = shoset2.ConnsByLname.GetConfig().InitFolders("127-0-0-1_9002")
	if err != nil {
		shoset2.Logger.Error().Msg("couldn't init folder: " + err.Error())
		return
	}
	err = shoset2.Certify(s9002, s9001)
	if err != nil {
		shoset2.Logger.Error().Msg("couldn't certify: " + err.Error())
		return
	}
	shoset2.Bind(s9002)

	handler, ok := shoset.Handlers["cfgjoin"]
	if !ok {
		t.Errorf("HandleDoubleWay didn't work, handler not ok")
	}

	c, _ := NewShosetConn(shoset2, s9001, OUT)

	protocolConn, err := tls.Dial(CONNECTION_TYPE, c.GetRemoteAddress(), c.GetShoset().GetTlsConfigDoubleWay())
	if err != nil {
		time.Sleep(time.Millisecond * time.Duration(5000))
		c.Logger.Error().Msg("HandleConfig err: " + err.Error())
	}
	defer func() {
		c.Logger.Debug().Msg("HandleConfig: socket closed")
		c.GetConn().Close()
	}()
	c.UpdateConn(protocolConn)

	cfg := msg.NewConfigProtocol(s9002, shoset2.GetLogicalName(), shoset2.GetShosetType(), PROTOCOL_JOIN)
	go handler.HandleDoubleWay(c, *cfg)

	time.Sleep(time.Millisecond * time.Duration(1000))
}
