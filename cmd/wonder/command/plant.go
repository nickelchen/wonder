package command

import (
	"flag"
	"fmt"
	"nickelchen/wonder/client"
	"strings"
	"time"

	"github.com/mitchellh/cli"
)

type PlantCommand struct {
	Ui cli.Ui
}

func (c *PlantCommand) Help() string {
	helpText := `
Usage: wonder plant [options]

	Plant some trees or flowers or grass in alice wonder land.

Options:
	--what choose from [tree, flower, grass]
	--color color of this plant, hex
	--number plant how many instances
`
	return strings.TrimSpace(helpText)
}

func (c *PlantCommand) Run(args []string) int {
	var what, color string
	var number int

	cmdFlags := flag.NewFlagSet("plant", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.StringVar(&what, "what", "flower", "plant what ?")
	cmdFlags.StringVar(&color, "color", "red", "what color is it?")
	cmdFlags.IntVar(&number, "number", 1, "how many")

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

	respCh := make(chan string)
	if err := cl.Plant(what, color, number, respCh); err != nil {
		c.Ui.Output(fmt.Sprintf("can not plant: %s", err))
		return 1
	}

	select {
	case r := <-respCh:
		c.Ui.Output(fmt.Sprintf("get plant response: %v\n", r))
	}

	return 0

}

func (c *PlantCommand) Synopsis() string {
	return "plant some trees or flowers in wonder land."
}
