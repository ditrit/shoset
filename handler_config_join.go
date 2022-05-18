package shoset

import (
	"errors"
	"sync"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
)

type ConfigJoinHandler struct{}

// GetConfigJoin :
func (cjh *ConfigJoinHandler) Get(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigProtocol
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfigJoin :
func (cjh *ConfigJoinHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigProtocol)
	ch := c.GetCh()
	dir := c.GetDir()
	remoteAddress := cfg.GetAddress()

	switch cfg.GetCommandName() {
	case PROTOCOL_JOIN:
		// incoming join request
		// a socket wants to join to this one
		if dir == IN {
			if ch.GetLogicalName() != cfg.GetLogicalName() || ch.GetShosetType() != cfg.GetShosetType() {
				// return early: invalid configuration (socket doesn't respect join protocol standards)
				c.SetIsValid(false)

				configNotOk := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), UNAKNOWLEDGE_JOIN)
				err := c.SendMessage(*configNotOk)
				if err != nil {
					c.logger.Warn().Msg("couldn't send configNotOk : " + err.Error())
				}
				return errors.New("error : Invalid connection for join - not the same type/name")
			}

			c.SetRemoteAddress(remoteAddress)
			c.SetRemoteLogicalName(cfg.GetLogicalName())
			c.SetRemoteShosetType(cfg.GetShosetType())
			ch.ConnsByName.Store(ch.GetLogicalName(), remoteAddress, PROTOCOL_JOIN, ch.GetShosetType(), c.ch.config.GetFileName(), c)

			configOk := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), AKNOWLEDGE_JOIN)
			err := c.SendMessage(*configOk)
			if err != nil {
				c.logger.Warn().Msg("couldn't send configOk : " + err.Error())
			}
		}

		cfgNewMember := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), MEMBER)
		conns, _ := ch.ConnsByName.smap.Load(ch.GetLogicalName())

		conns.(*sync.Map).Range(func(key, value interface{}) bool {
			SendBroNewMember(remoteAddress, cfgNewMember)(key.(string), value)
			return true
		})

	case AKNOWLEDGE_JOIN:
		// incoming aknowledge_join
		// join request validated
		c.SetRemoteLogicalName(cfg.GetLogicalName())
		c.SetRemoteShosetType(cfg.GetShosetType())
		ch.ConnsByName.Store(ch.GetLogicalName(), c.GetRemoteAddress(), PROTOCOL_JOIN, ch.GetShosetType(), c.ch.config.GetFileName(), c) // set conns in the other socket

	case UNAKNOWLEDGE_JOIN:
		// incoming unaknowledge_join
		// join request discarded
		c.SetIsValid(false)
		return errors.New("error : connection not ok")

	case MEMBER:
		// incoming member information
		// new shoset in the network, must join protocol it
		ch.Protocol(c.ch.GetBindAddress(), remoteAddress, PROTOCOL_JOIN)
	}
	return nil
}

// SendBroNewMember : inform the network that there is a new shoset connected
func SendBroNewMember(remoteAddress string, cfgNewMember *msg.ConfigProtocol) func(string, interface{}) {
	return func(address string, bro interface{}) {
		if address == remoteAddress {
			return
		}
		if err := bro.(*ShosetConn).SendMessage(*cfgNewMember); err != nil {
			bro.(*ShosetConn).logger.Warn().Msg("couldn't send cfgnewMember : " + err.Error())
		}
	}
}

// SendConfigBye :
func (cjh *ConfigJoinHandler) Send(c *Shoset, m msg.Message) {
	// no-op
	log.Warn().Msg("ConfigJoinHandler.Send not implemented")
}

// WaitConfig :
func (cjh *ConfigJoinHandler) Wait(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	// no-op
	log.Warn().Msg("ConfigJoinHandler.Wait not implemented")
	return nil
}
