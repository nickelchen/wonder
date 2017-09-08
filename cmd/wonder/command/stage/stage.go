package stage

import (
	"fmt"
	"sync"
	"time"

	"github.com/nickelchen/wonder/land"

	log "github.com/sirupsen/logrus"
)

type EventHandler interface {
	Handle(event land.Event)
}

type Stage struct {
	name        string
	config      *Config
	servers     []string
	serversLock sync.RWMutex
	shutdownCh  chan struct{}
}

func Create(config *Config) *Stage {
	shutdownCh := make(chan struct{})

	stage := Stage{
		name:       config.Name,
		config:     config,
		shutdownCh: shutdownCh,
	}
	return &stage
}

func (a *Stage) Enter() {
	log.Info("In command/stage/stage.go Enter()")
}

func (a *Stage) Leave() error {
	log.Info("In command/stage/stage.go Leave()")

	// simulate leaving process
	time.Sleep(2 * time.Second)
	return nil
}

func (a *Stage) ListServers() []string {
	a.serversLock.RLock()
	defer a.serversLock.Unlock()

	var servers []string
	for _, s := range a.servers {
		servers = append(servers, s)
	}
	return servers
}

func (a *Stage) ShutdownCh() <-chan struct{} {
	return a.shutdownCh
}

func (a *Stage) autoShutdown(d time.Duration) {
	select {
	case <-time.After(d):
		a.Leave()
		a.shutdownCh <- struct{}{}
	}
}
