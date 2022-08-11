package logger

import (
	"CTng/config"
	"CTng/gossip"
	"CTng/revocator"
	"net/http"
)

type LoggerContext struct {
	Config *config.Logger_config
	SRHs   []gossip.Gossip_object
	// FakeSRHs []gossip.Gossip_object
	// STHS           []gossip.Gossip_object
	// FakeSTHs       []gossip.Gossip_object
	Revocators     []*revocator.Revocator // array of all the CRV's
	Request_Count  int
	Current_Period int
	Client         *http.Client
}
