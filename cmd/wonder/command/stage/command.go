package stage

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/mitchellh/cli"
)

type Command struct {
	Ui cli.Ui
}

type Config struct {
	Name    string
	RPCAddr string
}

func (c *Command) readConfig(args []string) *Config {
	cmdFlags := flag.NewFlagSet("stage", flag.ContinueOnError)

	var rpcAddr string
	var nodeName string
	var debug bool
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.StringVar(&rpcAddr, "rpc-addr", "127.0.0.1:9898", "rpc ip:port to listen")
	cmdFlags.StringVar(&nodeName, "name", "beauty", "name of node")
	cmdFlags.BoolVar(&debug, "debug", true, "debug mode")

	if err := cmdFlags.Parse(args); err != nil {
		log.Fatalf("can not parse args: %s", err.Error())
	}

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	config := Config{
		Name:    nodeName,
		RPCAddr: rpcAddr,
	}

	return &config
}

func (c *Command) createStage(config *Config) *Stage {
	stage := Create(config)
	return stage
}

func (c *Command) startStage(config *Config, stage *Stage) *StageIPC {
	stage.Enter()

	ipcListener, err := net.Listen("tcp", config.RPCAddr)
	if err != nil {
		log.Error(fmt.Sprintf("can not listen. %s", err.Error()))
		return nil
	}

	ipc := NewStageIPC(stage, ipcListener)
	return ipc
}

func (c *Command) Run(args []string) int {
	log.Info("Stage is running.")
	config := c.readConfig(args)

	stage := c.createStage(config)
	if stage == nil {
		log.Error("can not create stage.")
		return 1
	}
	defer stage.Leave()

	ipc := c.startStage(config, stage)
	if ipc == nil {
		log.Error("can not start stage.")
		return 1
	}
	defer ipc.Shutdown()

	return c.exitSignals(stage)
}

func (c *Command) exitSignals(stage *Stage) int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	var sig os.Signal
	select {
	case s := <-signalCh:
		sig = s
	case <-stage.ShutdownCh():
		log.Info("stage shutdown himself.")
		return 0
	}

	graceful := false
	if sig == os.Interrupt || sig == syscall.SIGTERM {
		graceful = true
	}

	if !graceful {
		return 1
	}

	gracefulCh := make(chan struct{})
	gracefulTimeout := 3 * time.Second

	go func() {
		if err := stage.Leave(); err != nil {
			return
		}
		// gracefulCh <- struct{}{}
		close(gracefulCh)
	}()

	select {
	case <-signalCh:
		return 1
	case <-time.After(gracefulTimeout):
		return 1
	case <-gracefulCh:
		return 0
	}
}

func (c *Command) Help() string {
	return "stage command"
}

func (c *Command) Synopsis() string {
	return "stage command"
}
