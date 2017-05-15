package service

import (
	"log"
)

type PingerServer struct {
}

func (s *PingerServer) Ping(p Pinger_ping) error {
	msg, err := p.Params.Msg()
	if err != nil {
		return err
	}

	log.Printf("Ping: %s\n", msg)
	if err := p.Results.SetMsg("Ping: " + msg); err != nil {
		return err
	}

	return nil
}
