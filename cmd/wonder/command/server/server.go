package server

import (
	"fmt"
	"time"

	"github.com/nickelchen/wonder/client"
	"github.com/nickelchen/wonder/land"

	log "github.com/sirupsen/logrus"
)

type EventHandler interface {
	Handle(event land.Event)
}

type Server struct {
	name             string
	land             *land.Land
	config           *Config
	landConfig       *land.Config
	shutdownCh       chan struct{}
	eventCh          chan land.Event
	eventHandlers    map[EventHandler]struct{}
	eventHandlerList []EventHandler

	stageClient *client.RPCClient
	reportTimes int
}

func Create(config *Config) *Server {
	shutdownCh := make(chan struct{})
	eventCh := make(chan land.Event, 512)

	landConfig := land.DefaultConfig()
	landConfig.EventCh = eventCh

	l := land.Create(landConfig)

	server := Server{
		land:          l,
		config:        config,
		landConfig:    landConfig,
		shutdownCh:    shutdownCh,
		eventCh:       eventCh,
		eventHandlers: make(map[EventHandler]struct{}),
	}
	return &server
}

func (a *Server) Enter() {
	log.Info("In command/server/server.go Enter()")
	a.land.Spread()

	go a.eventLoop()

	// go a.autoShutdown(10 * time.Second)
}

func (a *Server) Leave() error {
	log.Info("In command/server/server.go Leave()")
	a.land.Shrink()

	// simulate leaving process
	time.Sleep(2 * time.Second)
	return nil
}

func (a *Server) ShutdownCh() <-chan struct{} {
	return a.shutdownCh
}

func (a *Server) autoShutdown(d time.Duration) {
	select {
	case <-time.After(d):
		a.Leave()
		a.shutdownCh <- struct{}{}
	}
}

func (a *Server) Ping() string {
	log.Info("=>")
	return "<="
}

func (a *Server) Plant(params *land.PlantParams) (*land.PlantResult, error) {
	result, err := a.land.Plant(params)
	return result, err
}

func (a *Server) Info(params *land.InfoParams) (*land.InfoResult, error) {
	result, err := a.land.Info(params)
	return result, err
}

func (a *Server) Subscribe(eh EventHandler) {
	a.eventHandlers[eh] = struct{}{}
	a.eventHandlerList = nil
	for eh := range a.eventHandlers {
		a.eventHandlerList = append(a.eventHandlerList, eh)
	}

	return
}

func (a *Server) ConnectStage() bool {
	stageConfig := client.Config{
		Addr:    a.config.StageAddr,
		Timeout: a.config.StageTimeout,
	}

	stageClient, err := client.ClientFromConfig(&stageConfig)
	a.stageClient = stageClient

	return err == nil
}

func (a *Server) ReportStage() {
	respCh := make(chan string)
	log.Debug("Report To Stage, times: ", a.reportTimes)

	a.stageClient.ServerAlive(a.config.ServerAddr, respCh)
	<-respCh

	a.reportTimes++
}

func (a *Server) eventLoop() {
	for {
		select {
		case event := <-a.eventCh:
			// get Event from land, fan out
			log.Debug(fmt.Sprintf("in server eventLoop, get event :%v", event))

			for _, eh := range a.eventHandlerList {
				eh.Handle(event)
			}
		}
	}
}
