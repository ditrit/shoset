package shoset

import (
	"sync"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
)

// ConfigJoinHandler implements MessageHandlers interface.
type ConfigJoinHandler struct{}

// Get returns the message for a given ShosetConn.
func (cjh *ConfigJoinHandler) Get(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigProtocol
	err := c.GetReader().ReadMessage(&cfg)
	return cfg, err
}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (cjh *ConfigJoinHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigProtocol)

	switch cfg.GetCommandName() {
	case PROTOCOL_JOIN:
		// incoming join request, a socket wants to join to this one.
		// save info and retrieve brothers to inform network.

		c.Logger.Trace().Str("lname", cfg.GetLogicalName()).Str("IP", cfg.GetAddress()).Msg("Incoming join request : " + PROTOCOL_JOIN)

		c.GetShoset().DeleteConn(cfg.GetLogicalName(), cfg.GetAddress())

		c.SetRemoteAddress(cfg.GetAddress())
		c.Store(PROTOCOL_JOIN, c.GetShoset().GetLogicalName(), cfg.GetAddress(), c.GetShoset().GetShosetType())

		// Send ACKNOWLEDGE_JOIN
		configOk := msg.NewConfigProtocol(cfg.GetAddress(), c.GetShoset().GetLogicalName(), c.GetShoset().GetShosetType(), ACKNOWLEDGE_JOIN)
		if err := c.GetWriter().SendMessage(*configOk); err != nil {
			c.Logger.Warn().Msg("couldn't send configOk : " + err.Error())
		}

		// Notify all the "members" = localBrothers (the ones with the same Lname) that a new member has joined so that they can all initiate a connection with the new member.
		cfgNewMember := msg.NewConfigProtocol(cfg.GetAddress(), c.GetShoset().GetLogicalName(), c.GetShoset().GetShosetType(), MEMBER)
		mapConns, _ := c.GetShoset().ConnsByLname.Load(c.GetShoset().GetLogicalName())
		mapConns.(*sync.Map).Range(func(key, value interface{}) bool {
			func(address string, bro interface{}) {
				if address == cfg.GetAddress() {
					return
				}
				if err := bro.(*ShosetConn).GetWriter().SendMessage(*cfgNewMember); err != nil {
					bro.(*ShosetConn).Logger.Warn().Msg("couldn't send cfgNewMember : " + err.Error())
				}
			}(key.(string), value)
			return true
		})

		// Notify all the remoteBrothers (the ones with a different Lname) that a new member has joined so that they can all initiate a connection with the new member.
		localBrothersArray := []string{} // localBrothers = members = socket with the same Lname
		if localBrothers, _ := c.GetShoset().ConnsByLname.Load(c.GetShoset().GetLogicalName()); localBrothers != nil {
			localBrothersArray = Keys(localBrothers.(*sync.Map), ALL)
		}

		c.GetShoset().ConnsByLname.Iterate(
			func(lname string, ipAddress string, remoteBro interface{}) {
				if lname != c.GetShoset().GetLogicalName() { // if remoteBrother
					c.Logger.Debug().Msg("sending brother info to " + ipAddress)
					remoteBrothers, _ := c.GetShoset().ConnsByLname.Load(lname) // remoteBrothers = sockets with the remote Lname = sockets linked to this one
					remoteBrothersArray := []string{}
					if remoteBrothers != nil {
						remoteBrothersArray = Keys(remoteBrothers.(*sync.Map), ALL) // get the IP addresses of the remoteBrothers
					}
					cfgBrothers := msg.NewConfigBrothersProtocol(localBrothersArray, remoteBrothersArray, BROTHERS, c.GetShoset().GetLogicalName(), c.GetShoset().GetShosetType())
					if err := remoteBro.(*ShosetConn).GetWriter().SendMessage(*cfgBrothers); err != nil {
						remoteBro.(*ShosetConn).Logger.Warn().Msg("couldn't send brothers : " + err.Error())
					}
				}
			},
		)

	case ACKNOWLEDGE_JOIN:
		// incoming acknowledge_join, join request validated.
		// save info.

		c.Logger.Trace().Str("lname", cfg.GetLogicalName()).Str("IP", cfg.GetAddress()).Msg("Incoming acknowledge join : " + ACKNOWLEDGE_JOIN)

		// c.GetShoset().DeleteConn(cfg.GetLogicalName(), cfg.GetAddress()) // we never delete a conn

		c.Store(PROTOCOL_JOIN, c.GetShoset().GetLogicalName(), c.GetRemoteAddress(), c.GetShoset().GetShosetType())

		// Deletes the IP from the list of started but not yet ready.
		c.GetShoset().LaunchedProtocol.DeleteFromConcurentSlice(c.GetRemoteAddress())

	case MEMBER:
		// incoming member information.
		// need to link protocol on it if not already in the map of known conn.

		c.Logger.Trace().Str("lname", cfg.GetLogicalName()).Str("IP", cfg.GetAddress()).Msg("Incoming brother member : " + MEMBER)

		mapConns, _ := c.GetShoset().ConnsByLname.Load(c.GetShoset().GetLogicalName())
		if mapConns == nil {
			return nil
		}
		if exists, _ := mapConns.(*sync.Map).Load(cfg.GetAddress()); exists == nil {
			c.GetShoset().Protocol(c.GetShoset().GetBindAddress(), cfg.GetAddress(), PROTOCOL_JOIN)
		}
	}
	return nil
}

// Send sends the message through the given Shoset network.
func (cjh *ConfigJoinHandler) Send(s *Shoset, m msg.Message) {
	// no-op
	log.Warn().Msg("ConfigJoinHandler.Send not implemented")
}

// Wait returns the message received for a given Shoset.
func (cjh *ConfigJoinHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	// no-op
	log.Warn().Msg("ConfigJoinHandler.Wait not implemented")
	return nil
}
