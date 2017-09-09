package command

import (
	"flag"
	"fmt"
	"github.com/nickelchen/wonder/client"
	"strings"
	"time"

	"github.com/mitchellh/cli"
)

type ListCommand struct {
	Ui cli.Ui
}

func (c *ListCommand) Help() string {
	helpText := `
Usage: wonder list [options]

	List alive servers
`
	return strings.TrimSpace(helpText)
}

func (c *ListCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("list", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	config := client.Config{
		Addr:    "127.0.0.1:9898",
		Timeout: 20 * time.Second,
	}
	cl, err := client.ClientFromConfig(&config)
	if err != nil {
		c.Ui.Output(fmt.Sprintf("can not get client: %s", err))
		return 1
	}

	respCh := make(chan []string)
	if err := cl.ListServers(respCh); err != nil {
		c.Ui.Output(fmt.Sprintf("can not list: %s", err))
		return 1
	}

	select {
	case r := <-respCh:
		c.Ui.Output(fmt.Sprintf("get list response: %v\n", r))
	}

	return 0

}

func (c *ListCommand) Synopsis() string {
	return "list some trees or flowers in wonder land."
}
