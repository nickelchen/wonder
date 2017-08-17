package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/mitchellh/cli"

	"github.com/nickelchen/wonder/client"
	"github.com/nickelchen/wonder/cmd/wonder/command/render"
	"github.com/nickelchen/wonder/share"
)

var gRow int
var gCol int

func init() {
	gRow, _ = strconv.Atoi(os.Getenv("ROW"))
	gCol, _ = strconv.Atoi(os.Getenv("COL"))
}

type InfoCommand struct {
	Ui cli.Ui

	board *share.GameBoard
}

func (c *InfoCommand) Help() string {
	helpText := `
Usage: wonder info

	Get every information about wonder land. including tiles, sprites etc.
`
	return strings.TrimSpace(helpText)
}

func (c *InfoCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("information", flag.ContinueOnError)
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
		c.Ui.Output(fmt.Sprintf("can not get client: %s\n", err))
		return 1
	}

	board := &share.GameBoard{}
	c.board = board

	rend := render.TermRender{}
	rend.Stage(board, 80, 24, c.Ui.(*cli.BasicUi).Writer)

	respCh1 := make(chan share.InfoResponseObj, 512)
	if err := cl.Info(respCh1); err != nil {
		c.Ui.Output(fmt.Sprintf("can not get info: %s\n", err))
		return 1
	}

	c.receiveInfoItems(respCh1, &rend)

	respCh2 := make(chan share.EventResponseObj, 512)
	if err := cl.Subscribe(respCh2); err != nil {
		c.Ui.Output(fmt.Sprintf("can not subscribe: %s\n", err))
		return 1
	}

	// long run polling events from server
	go c.receiveEventItems(respCh2, &rend)

	rend.Loop()

	return 0
}

func (c *InfoCommand) receiveInfoItems(respCh chan share.InfoResponseObj, rend render.InfoRender) {

	for {
		select {
		// receive from info response
		case r := <-respCh:
			t := r.Type
			p := r.Payload

			c.Ui.Output(fmt.Sprintf("Get Info Response Payload: %v", string(p)))

			switch t {
			case share.InfoItemTypeTile:
				tiles := [][]share.Tile{}
				json.Unmarshal(p, &tiles)
				c.Ui.Output(fmt.Sprintf("receive tiles struct is: %v\n", tiles))

				c.board.Tiles = tiles

			case share.InfoItemTypeTree:
				spr := share.Tree{}
				json.Unmarshal(p, &spr)
				c.Ui.Output(fmt.Sprintf("receive tree struct is: %v\n", spr))

				c.board.Trees = append(c.board.Trees, spr)

			case share.InfoItemTypeFlower:
				spr := share.Flower{}
				json.Unmarshal(p, &spr)
				c.Ui.Output(fmt.Sprintf("receive flower struct is: %v\n", spr))

				c.board.Flowers = append(c.board.Flowers, spr)

			case share.InfoItemTypeGrass:
				spr := share.Grass{}
				json.Unmarshal(p, &spr)
				c.Ui.Output(fmt.Sprintf("receive grass struct is: %v\n", spr))

				c.board.Grasses = append(c.board.Grasses, spr)

			case share.InfoItemTypeDone:
				c.Ui.Output("received all repsonse. finish")

				return
			}
		}

		rend.Render()
	}
}

func (c *InfoCommand) receiveEventItems(respCh chan share.EventResponseObj, rend render.InfoRender) {

	for {
		select {
		// receive from subscribe response
		case r := <-respCh:
			t := r.Type
			p := r.Payload

			c.Ui.Output(fmt.Sprintf("Get Subscribed Response Payload: %v", string(p)))

			switch t {
			case share.EventTypeMove:
				event := share.SpriteMove{}
				json.Unmarshal(p, &event)
				c.Ui.Output(fmt.Sprintf("receive sprite move struct is: %v\n", event))

				c.board.MoveEvents = append(c.board.MoveEvents, event)

			case share.EventTypeAdd:
				event := share.SpriteAdd{}
				json.Unmarshal(p, &event)
				c.Ui.Output(fmt.Sprintf("receive sprite add struct is: %v\n", event))

				c.board.AddEvents = append(c.board.AddEvents, event)

			case share.EventTypeDelete:
				event := share.SpriteDelete{}
				json.Unmarshal(p, &event)
				c.Ui.Output(fmt.Sprintf("receive sprite delete struct is: %v\n", event))

				c.board.DeleteEvents = append(c.board.DeleteEvents, event)
			}

		}

		rend.Render()
	}
}

func (c *InfoCommand) Synopsis() string {
	return "The whole woner land information."
}
