package render

import (
	"io"

	"github.com/nickelchen/wonder/share"
)

type InfoRender interface {
	Stage(*share.GameBoard, int, int, io.Writer)
	Render()
	Loop()
}
