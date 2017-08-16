package alice

import (
	"fmt"
	"nickelchen/wonder/land"
	"time"

	log "github.com/sirupsen/logrus"
)

type EventHandler interface {
	Handle(event land.Event)
}

type Alice struct {
	name             string
	land             *land.Land
	config           *Config
	landConfig       *land.Config
	shutdownCh       chan struct{}
	eventCh          chan land.Event
	eventHandlers    map[EventHandler]struct{}
	eventHandlerList []EventHandler
}

func Create(config *Config) *Alice {
	shutdownCh := make(chan struct{})
	eventCh := make(chan land.Event, 512)

	landConfig := land.DefaultConfig()
	landConfig.EventCh = eventCh

	l := land.Create(landConfig)

	alice := Alice{
		name:          config.Name,
		land:          l,
		config:        config,
		landConfig:    landConfig,
		shutdownCh:    shutdownCh,
		eventCh:       eventCh,
		eventHandlers: make(map[EventHandler]struct{}),
	}
	return &alice
}

func (a *Alice) Enter() {
	log.Info("In command/alice/alice.go Enter()")
	a.land.Spread()

	go a.eventLoop()

	// go a.autoShutdown(10 * time.Second)
}

func (a *Alice) Leave() error {
	log.Info("In command/alice/alice.go Leave()")
	a.land.Shrink()

	// simulate leaving process
	time.Sleep(2 * time.Second)
	return nil
}

func (a *Alice) ShutdownCh() <-chan struct{} {
	return a.shutdownCh
}

func (a *Alice) autoShutdown(d time.Duration) {
	select {
	case <-time.After(d):
		a.Leave()
		a.shutdownCh <- struct{}{}
	}
}

func (a *Alice) Ping() string {
	log.Info("=>")
	return "<="
}

func (a *Alice) Plant(params *land.PlantParams) (*land.PlantResult, error) {
	result, err := a.land.Plant(params)
	return result, err
}

func (a *Alice) Info(params *land.InfoParams) (*land.InfoResult, error) {
	result, err := a.land.Info(params)
	return result, err
}

func (a *Alice) Subscribe(eh EventHandler) {
	a.eventHandlers[eh] = struct{}{}
	a.eventHandlerList = nil
	for eh := range a.eventHandlers {
		a.eventHandlerList = append(a.eventHandlerList, eh)
	}

	return
}

func (a *Alice) eventLoop() {
	for {
		select {
		case event := <-a.eventCh:
			// get Event from land, fan out
			log.Debug(fmt.Sprintf("in alice eventLoop, get event :%v", event))

			for _, eh := range a.eventHandlerList {
				eh.Handle(event)
			}
		}
	}
}
