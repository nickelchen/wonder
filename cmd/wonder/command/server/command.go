package server

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

var DefaultStageTimeout = 10 * time.Second
var DefaultReportInterval = 2 * time.Second

type Config struct {
	bindIP     string
	ServerAddr string

	StageAddr      string
	StageTimeout   time.Duration
	ReportInterval time.Duration
}

func (c *Command) readConfig(args []string) *Config {
	cmdFlags := flag.NewFlagSet("server", flag.ContinueOnError)

	var stageAddr string
	var stageTimeout int
	var reportInterval int

	var bindIP string
	var debug bool

	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.StringVar(&stageAddr, "stage-addr", "127.0.0.1:9898", "which stage doest the server to report")
	cmdFlags.IntVar(&stageTimeout, "stage-timeout", 0, "timeout when connect to stage")
	cmdFlags.IntVar(&reportInterval, "stage-report-interval", 0, "time interval to report to stage")

	cmdFlags.StringVar(&bindIP, "bind-ip", "127.0.0.1", "this server bind ip address")

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
		bindIP:         bindIP,
		StageAddr:      stageAddr,
		StageTimeout:   time.Duration(stageTimeout) * time.Second,
		ReportInterval: time.Duration(reportInterval) * time.Second,
	}

	return &config
}

func (c *Command) createServer(config *Config) *Server {
	server := Create(config)
	return server
}

func (c *Command) startServer(config *Config, server *Server) *ServerIPC {
	server.Enter()

	// let the os to determine which port to listen
	tcpAddr := &net.TCPAddr{IP: net.ParseIP(config.bindIP), Port: 0}
	ipcLn, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		log.Error(fmt.Sprintf("can not listen. %s", err.Error()))
		return nil
	}

	config.ServerAddr = ipcLn.Addr().(*net.TCPAddr).String()

	ipc := NewServerIPC(server, ipcLn)
	return ipc
}

func (c *Command) Run(args []string) int {
	log.Info("Server is running.")
	config := c.readConfig(args)

	server := c.createServer(config)
	if server == nil {
		log.Error("can not create server.")
		return 1
	}
	defer server.Leave()

	ipc := c.startServer(config, server)
	if ipc == nil {
		log.Error("can not start server.")
		return 1
	}
	defer ipc.Shutdown()

	if ok := c.connectStage(config, server); !ok {
		log.Error("Cannot connect to stage, shutdown current server")
		return 1
	}

	go c.reportStage(config, server)
	go c.pingLoop(server)

	// block until we get a signal
	return c.exitSignals(server)
}

func (c *Command) pingLoop(server *Server) {
	for {
		pong := server.Ping()
		c.Ui.Output(pong)

		time.Sleep(10 * time.Second)
	}
}

func (c *Command) connectStage(config *Config, server *Server) bool {
	if config.StageTimeout == 0 {
		config.StageTimeout = DefaultStageTimeout
	}
	log.Debug("config.StageTimeout: ", config.StageTimeout)

	return server.ConnectStage()
}

func (c *Command) reportStage(config *Config, server *Server) {
	if config.ReportInterval == 0 {
		config.ReportInterval = DefaultReportInterval
	}

	for {
		select {
		case <-time.After(config.ReportInterval):
			server.ReportStage()
		}
	}
}

func (c *Command) exitSignals(server *Server) int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	var sig os.Signal
	select {
	case s := <-signalCh:
		sig = s
	case <-server.ShutdownCh():
		log.Info("server shutdown himself.")
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
		if err := server.Leave(); err != nil {
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
	return "server command"
}

func (c *Command) Synopsis() string {
	return "server command"
}
