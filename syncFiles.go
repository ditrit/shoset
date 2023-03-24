package shoset

import (
	fileMod "github.com/ditrit/shoset/file"
	msg "github.com/ditrit/shoset/msg"
)

type SyncFiles struct {
	fileMod.FileLibrary
}

func (s *Shoset) InitLibrary(baseDirectory string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.baseDirectory = baseDirectory
	var err error
	s.Library, err = fileMod.NewFileLibrary(baseDirectory)
	if err != nil {
		s.Logger.Error().Msg(err.Error())
	}
	s.Handlers["file"].(*FileHandler).Init(s.Library, s.Logger, s.Queue["file"], func(message *msg.FileMessage) { s.Handlers["file"].(*FileHandler).Send(s, message) })
}

func (s *Shoset) GetBaseDirectory() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.baseDirectory
}

func (s *Shoset) SyncLibrary(conn *ShosetConn) {
	publishMessage, err := s.Library.GetMessageLibrary()
	if err != nil {
		s.Logger.Error().Msg(err.Error())
		return
	}
	conn.SendMessage(publishMessage)
	conn.GetShoset().Logger.Trace().Msg("-------------SyncLibrary " + conn.GetLocalAddress() + " with " + conn.GetRemoteAddress())
}
