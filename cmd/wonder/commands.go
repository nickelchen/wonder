package main

import (
	"os"
	"os/signal"

	"github.com/mitchellh/cli"

	"github.com/nickelchen/wonder/cmd/wonder/command"
	"github.com/nickelchen/wonder/cmd/wonder/command/server"
	"github.com/nickelchen/wonder/cmd/wonder/command/stage"
)

var Commands map[string]cli.CommandFactory

func init() {
	ui := &cli.BasicUi{Writer: os.Stdout}

	Commands = map[string]cli.CommandFactory{
		"stage": func() (cli.Command, error) {
			return &stage.Command{
				Ui: ui,
			}, nil
		},
		"server": func() (cli.Command, error) {
			return &server.Command{
				Ui: ui,
			}, nil
		},
		"plant": func() (cli.Command, error) {
			return &command.PlantCommand{
				Ui: ui,
			}, nil
		},

		"list": func() (cli.Command, error) {
			return &command.ListCommand{
				Ui: ui,
			}, nil
		},

		"info": func() (cli.Command, error) {
			fh, _ := os.OpenFile("./logs/info.log",
				os.O_RDWR|os.O_APPEND|os.O_CREATE, os.FileMode(0755))
			fl := &cli.BasicUi{Writer: fh}

			return &command.InfoCommand{
				Ui: fl,
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Revision:          GitCommit,
				Version:           Version,
				VersionPrerelease: VersionPrerelease,
				Ui:                ui,
			}, nil
		},
	}
}

func makeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})

	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()

	return resultCh
}
