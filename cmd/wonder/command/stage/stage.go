package stage

import (
	"sync"
	"time"

	"github.com/nickelchen/wonder/land"

	log "github.com/sirupsen/logrus"
)

type EventHandler interface {
	Handle(event land.Event)
}

// if the one server didnot report in 10 seconds, then mark dead.
var DefaultServerExpireTimeout = 10 * time.Second
var CheckServerExpireInterval = 10 * time.Second

type ServerState struct {
	aliveServers map[string]time.Time
	deadServers  map[string]time.Time
	l            sync.RWMutex
}

type Stage struct {
	name        string
	config      *Config
	shutdownCh  chan struct{}
	serverState *ServerState
}

func Create(config *Config) *Stage {
	shutdownCh := make(chan struct{})

	serverState := &ServerState{
		aliveServers: make(map[string]time.Time),
		deadServers:  make(map[string]time.Time),
	}
	stage := Stage{
		name:        config.Name,
		config:      config,
		shutdownCh:  shutdownCh,
		serverState: serverState,
	}
	return &stage
}

func (a *Stage) Enter() {
	log.Info("In command/stage/stage.go Enter()")
	go a.cleanDeadServers()
}

func (a *Stage) Leave() error {
	log.Info("In command/stage/stage.go Leave()")

	// simulate leaving process
	time.Sleep(2 * time.Second)
	return nil
}

func (a *Stage) ListServers() ([]string, error) {
	a.serverState.l.RLock()
	defer a.serverState.l.RUnlock()

	var servers []string
	for s, _ := range a.serverState.aliveServers {
		servers = append(servers, s)
	}
	return servers, nil
}

func (a *Stage) ServerAlive(serverAddr string) (string, error) {
	a.serverState.l.Lock()
	defer a.serverState.l.Unlock()

	a.serverState.aliveServers[serverAddr] = time.Now()
	delete(a.serverState.deadServers, serverAddr)

	// TODO for debug only.
	var ks []string
	for k, _ := range a.serverState.aliveServers {
		ks = append(ks, k)
	}
	log.Debug("Alive servers: ", ks)

	return "i know u are alive. good job! ", nil
}

func (a *Stage) cleanDeadServers() {
	for {
		select {
		case <-time.After(CheckServerExpireInterval):
			a.serverState.l.Lock()
			for s, updated := range a.serverState.aliveServers {
				// if the updated time is too old, move it to dead servers map.
				if time.Since(updated) > DefaultServerExpireTimeout {
					delete(a.serverState.aliveServers, s)
					a.serverState.deadServers[s] = updated
				}
			}
			a.serverState.l.Unlock()
		}
	}
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
