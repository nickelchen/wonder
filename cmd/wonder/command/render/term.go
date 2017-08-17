package render

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/nickelchen/wonder/share"

	termbox "github.com/nsf/termbox-go"
)

var debug bool

const textColor = termbox.ColorBlack
const backgroundColor = termbox.ColorBlack

const animationSpeed = 4 * time.Second

var elemColor map[string]termbox.Attribute

const blockSize = 2

var (
	CodePointSpace  string
	CodePointTree   string
	CodePointFlower string
	CodePointGrass  string
	CodePointHuman  string
)

func readHexCodePoint(key string) string {
	valueBytes, _ := hex.DecodeString(os.Getenv(key))
	return fmt.Sprintf("%s ", valueBytes)
}
func readColorCode(key string) termbox.Attribute {
	value, _ := strconv.Atoi(os.Getenv(key))
	fmt.Printf("key : %s value : %d\n", key, value)

	return termbox.Attribute(value)
}

func readDebug() bool {
	return "true" == strings.ToLower(os.Getenv("DEBUG"))
}

func init() {
	CodePointSpace = readHexCodePoint("CodePointSpace")
	CodePointTree = readHexCodePoint("CodePointTree")
	CodePointFlower = readHexCodePoint("CodePointFlower")
	CodePointGrass = readHexCodePoint("CodePointGrass")
	CodePointHuman = readHexCodePoint("CodePointHuman")

	elemColor = map[string]termbox.Attribute{
		"ground": readColorCode("COLOR_GROUND"),
		"mud":    readColorCode("COLOR_MUD"),
		"alice":  readColorCode("COLOR_ALICE"),
		"tree":   readColorCode("COLOR_TREE"),
		"grass":  readColorCode("COLOR_GRASS"),
		"flower": readColorCode("COLOR_FLOWER"),
	}

	debug = readDebug()
}

type TermRender struct {
	renderLock sync.Mutex
	offsetX    int
	offsetY    int

	board  *share.GameBoard
	logger io.Writer
}

func (u *TermRender) Stage(board *share.GameBoard, stageWidth, stageHeight int, logger io.Writer) {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetOutputMode(termbox.Output256)

	u.board = board
	u.logger = logger

	w, h := termbox.Size()
	io.WriteString(u.logger, fmt.Sprintf("termbox.Size w: %d, h:%d\n", w, h))

	u.offsetX = (w - stageWidth) / 2
	u.offsetY = (h - stageHeight) / 2
}

func (u *TermRender) Loop() {
	defer termbox.Close()

	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	for {
		ev := <-eventQueue
		if ev.Type == termbox.EventKey {
			switch {
			case ev.Key == termbox.KeyArrowUp || ev.Ch == 'k':
				io.WriteString(u.logger, fmt.Sprintf("get keyevent: %v\n", ev))
			case ev.Key == termbox.KeyArrowDown || ev.Ch == 'j':
				io.WriteString(u.logger, fmt.Sprintf("get keyevent: %v\n", ev))
			case ev.Key == termbox.KeyArrowLeft || ev.Ch == 'h':
				io.WriteString(u.logger, fmt.Sprintf("get keyevent: %v\n", ev))
			case ev.Key == termbox.KeyArrowRight || ev.Ch == 'l':
				io.WriteString(u.logger, fmt.Sprintf("get keyevent: %v\n", ev))
			case ev.Ch == 'n':
				io.WriteString(u.logger, fmt.Sprintf("get keyevent: %v\n", ev))
			case ev.Ch == 'p':
				io.WriteString(u.logger, fmt.Sprintf("get keyevent: %v\n", ev))
			case ev.Ch == 'r':
				io.WriteString(u.logger, fmt.Sprintf("get keyevent: %v\n", ev))
			case ev.Ch == 'd':
				io.WriteString(u.logger, fmt.Sprintf("get keyevent: %v\n", ev))
			case ev.Key == termbox.KeyEsc:
				return
			}
		}
		u.Render()
		time.Sleep(animationSpeed)
	}
}

func (u *TermRender) Render() {
	termbox.Clear(backgroundColor, backgroundColor)

	tiles := u.board.Tiles
	for y := 1; y <= len(tiles); y++ {
		tilesRow := tiles[y-1]
		for x := 1; x <= len(tilesRow); x++ {
			t := tilesRow[x-1]
			if t.Gradient > 0 {
				u.RenderMud(x, y)
			} else {
				u.RenderGround(x, y)
			}
		}
	}

	for _, s := range u.board.Trees {
		p := s.GetPoint()
		u.RenderTree(p.X+1, p.Y+1)
	}
	for _, s := range u.board.Flowers {
		p := s.GetPoint()
		u.RenderFlower(p.X+1, p.Y+1)
	}
	for _, s := range u.board.Grasses {
		p := s.GetPoint()
		u.RenderGrass(p.X+1, p.Y+1)
	}

	if debug {
		for i := 1; i < 256; i++ {
			z := i / 100
			y := (i % 100) / 10
			x := (i % 100) % 10 / 1

			u.Render256(i, x, y, z)
		}

	}

	termbox.Flush()

}

func (u *TermRender) Render256(i, x, y, z int) {
	row := i % 16
	col := i / 16

	color := termbox.Attribute(z*100 + y*10 + x)
	termbox.SetCell(col*3, row, []rune(strconv.Itoa(z))[0], textColor, color)
	termbox.SetCell(col*3+1, row, []rune(strconv.Itoa(y))[0], textColor, color)
	termbox.SetCell(col*3+2, row, []rune(strconv.Itoa(x))[0], textColor, color)
}

func (u *TermRender) RenderFlower(x, y int) {
	color := elemColor["flower"]

	for k := 0; k < blockSize; k++ {
		termbox.SetCell(u.offsetX+x*blockSize+k, u.offsetY+y, 'F', textColor, color)
	}

}
func (u *TermRender) RenderTree(x, y int) {
	color := elemColor["tree"]

	for k := 0; k < blockSize; k++ {
		termbox.SetCell(u.offsetX+x*blockSize+k, u.offsetY+y, 'T', textColor, color)
	}
}
func (u *TermRender) RenderGrass(x, y int) {
	color := elemColor["grass"]

	for k := 0; k < blockSize; k++ {
		termbox.SetCell(u.offsetX+x*blockSize+k, u.offsetY+y, 'G', textColor, color)
	}
}

func (u *TermRender) RenderGround(x, y int) {
	color := elemColor["ground"]

	for k := 0; k < blockSize; k++ {
		termbox.SetCell(u.offsetX+x*blockSize+k, u.offsetY+y, ' ', textColor, color)
	}
}
func (u *TermRender) RenderMud(x, y int) {
	color := elemColor["mud"]

	for k := 0; k < blockSize; k++ {
		termbox.SetCell(u.offsetX+x*blockSize+k, u.offsetY+y, ' ', textColor, color)
	}
}
func (u *TermRender) RenderAlice(x, y int) {
	color := elemColor["alice"]

	for k := 0; k < blockSize; k++ {
		termbox.SetCell(u.offsetX+x*blockSize+k, u.offsetY+y, 'A', textColor, color)
	}
}
