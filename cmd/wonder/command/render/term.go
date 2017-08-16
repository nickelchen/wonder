package render

import (
	"encoding/hex"
	"fmt"
	"os"
	"sync"

	_ "github.com/joho/godotenv/autoload"
	"github.com/mitchellh/cli"

	tm "github.com/buger/goterm"
)

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

func init() {
	CodePointSpace = readHexCodePoint("CodePointSpace")
	CodePointTree = readHexCodePoint("CodePointTree")
	CodePointFlower = readHexCodePoint("CodePointFlower")
	CodePointGrass = readHexCodePoint("CodePointGrass")
	CodePointHuman = readHexCodePoint("CodePointHuman")
}

type TermRender struct {
	renderLock sync.Mutex
	offsetRow  int
	offsetCol  int

	Ui cli.Ui
}

func (u *TermRender) Reset(stageRow, stageCol int) {
	u.renderLock.Lock()
	defer u.renderLock.Unlock()

	u.offsetRow = (tm.Height() - stageRow) / 2
	u.offsetCol = (tm.Width() - stageCol) / 2

	tm.Clear()
}

func (u *TermRender) RenderRow(row int, str string) {
	u.renderLock.Lock()
	defer u.renderLock.Unlock()
	tm.MoveCursor(row+u.offsetRow, 1+u.offsetCol)
	tm.Print(str)
	tm.Flush()

}

func (u *TermRender) RenderChar(row, col int, char string, color int) {
	u.renderLock.Lock()
	defer u.renderLock.Unlock()

	u.Ui.Output(fmt.Sprintf("char is :%v", char))

	tm.MoveCursor(row+u.offsetRow, col+u.offsetCol)
	tm.Print(tm.Color(tm.Background(char, tm.WHITE), color))
	tm.Flush()

}

func (u *TermRender) RenderCharBg(row, col int, char string, color int) {
	u.renderLock.Lock()
	defer u.renderLock.Unlock()

	tm.MoveCursor(row+u.offsetRow, col+u.offsetCol)
	tm.Print(tm.Background(char, color))
	tm.Flush()

}

func (u *TermRender) RenderFlower(row, col int) {
	u.RenderChar(row, col, CodePointFlower, tm.RED)
}
func (u *TermRender) RenderTree(row, col int) {
	u.RenderChar(row, col, CodePointTree, tm.GREEN)
}
func (u *TermRender) RenderGrass(row, col int) {
	u.RenderChar(row, col, CodePointGrass, tm.GREEN)
}
func (u *TermRender) RenderGround(row, col int) {
	u.RenderCharBg(row, col, CodePointSpace, tm.YELLOW)
}
func (u *TermRender) RenderMud(row, col int) {
	u.RenderCharBg(row, col, CodePointSpace, tm.WHITE)
}
func (u *TermRender) RenderAlice(row, col int) {
	u.RenderCharBg(row, col, CodePointHuman, tm.WHITE)
}
