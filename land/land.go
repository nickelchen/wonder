package land

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"

	"github.com/nickelchen/wonder/share"
)

var gRow int
var gCol int

func init() {
	gRow, _ = strconv.Atoi(os.Getenv("ROW"))
	gCol, _ = strconv.Atoi(os.Getenv("COL"))
}

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
}

type Land struct {
	tiles       [][]share.Tile
	sprites     []share.Sprite
	spritesLock sync.RWMutex
	config      *Config
}

type Config struct {
	EventCh chan Event
}

type Event struct {
	Type string
	Item interface{}
}

type PlantParams struct {
	What   share.PlantType
	Color  string
	Number int
}

type PlantResult struct {
	Succ int
	Fail int
}

type InfoParams struct {
}

type InfoResult struct {
	resultCh chan InfoResultItem
}

type InfoResultItem struct {
	Type string
	Item interface{}
}

func (i InfoResult) ResultCh() <-chan InfoResultItem {
	return i.resultCh
}

func DefaultConfig() *Config {
	return &Config{}
}

func Create(config *Config) *Land {
	var land Land = Land{
		config: config,
	}
	return &land
}

func (l *Land) Spread() int {
	l.tiles = randTiles()
	l.sprites = randSprites()

	l.putAlice()

	go l.spawnEvents()

	log.Info("land/land.go Spread()")
	return 0
}

func randTiles() [][]share.Tile {
	var tiles [][]share.Tile

	for i := 0; i < gRow; i++ {
		var row []share.Tile
		for j := 0; j < gCol; j++ {
			row = append(row, share.Tile{Gradient: rand.Int() % 2})
		}
		tiles = append(tiles, row)
	}

	return tiles
}

func randSprites() []share.Sprite {
	var sprites []share.Sprite

	return sprites
}

func (l *Land) putAlice() {
	// put alice.
	point := l.findEmptyPoint()
	alice := share.Human{Name: "Alice"}
	alice.PutToPoint(point)
	l.sprites = append(l.sprites, alice)
}

func (l *Land) moveAlice(dir share.MoveDirection) {
	var sprites []share.Sprite

	l.spritesLock.Lock()
	defer l.spritesLock.Unlock()

	for _, s := range l.sprites {
		if isAlice(s) {
			a, _ := s.(share.Human)
			// found alice. move its place
			p := a.P
			switch dir {
			case share.MoveUp:
				p.Y -= 1
			case share.MoveDown:
				p.Y += 1
			case share.MoveLeft:
				p.X -= 1
			case share.MoveRight:
				p.X += 1
			}
			// s and a is the same object. put s is equivlent.
			a.PutToPoint(p)
			sprites = append(sprites, a)

		} else {
			sprites = append(sprites, s)

		}
	}

	l.sprites = sprites
}

func (l *Land) getAlice() (share.Human, error) {
	l.spritesLock.Lock()
	defer l.spritesLock.Unlock()

	for _, s := range l.sprites {
		if isAlice(s) {
			h, _ := s.(share.Human)
			return h, nil
		}
	}
	return share.Human{}, errors.New("can not find alice")
}

func (l *Land) getGhost() (share.Human, error) {
	l.spritesLock.Lock()
	defer l.spritesLock.Unlock()

	for _, s := range l.sprites {
		if isGhost(s) {
			h, _ := s.(share.Human)
			return h, nil
		}
	}
	return share.Human{}, errors.New("can not find ghost")
}

func isAlice(sprite share.Sprite) bool {
	h, ok := sprite.(share.Human)
	return ok && h.Name == "Alice"
}

func isGhost(sprite share.Sprite) bool {
	h, ok := sprite.(share.Human)
	return ok && h.Name == "Ghost"
}

func (l *Land) findEmptyPoint() share.Point {
	return share.Point{X: rand.Int() % gCol, Y: rand.Int() % gRow}
}

func (l *Land) Shrink() {
	log.Info("land/land.go Shrink()")
}

func (l *Land) Plant(params *PlantParams) (*PlantResult, error) {
	log.Info("land/land.go Plant()")

	var s share.Sprite
	for i := 0; i < params.Number; i++ {
		point := l.findEmptyPoint()
		switch params.What {
		case share.PlantTree:
			o := share.Tree{}
			o.PutToPoint(point)
			s = o

		case share.PlantFlower:
			o := share.Flower{}
			o.PutToPoint(point)
			s = o

		case share.PlantGrass:
			o := share.Grass{}
			o.PutToPoint(point)
			s = o
		}

		l.sprites = append(l.sprites, s)
	}

	result := PlantResult{
		Succ: params.Number,
		Fail: 0,
	}

	return &result, nil
}

func (l *Land) Info(params *InfoParams) (*InfoResult, error) {
	l.spritesLock.Lock()
	defer l.spritesLock.Unlock()

	resultCh := make(chan InfoResultItem, 1+len(l.sprites))

	defer func() {
		go l.sendResultItem(resultCh)
	}()

	result := InfoResult{
		resultCh: resultCh,
	}

	return &result, nil
}

func (l *Land) sendResultItem(resultCh chan InfoResultItem) {
	resultCh <- InfoResultItem{
		Type: share.InfoItemTypeTile,
		Item: l.tiles,
	}
	for _, sprite := range l.sprites {
		var st string
		switch sprite.(type) {
		case share.Tree:
			st = share.InfoItemTypeTree
		case share.Flower:
			st = share.InfoItemTypeFlower
		case share.Grass:
			st = share.InfoItemTypeGrass
		case share.Human:
			st = share.InfoItemTypeHuman
		}
		log.Debug(fmt.Sprintf("sendResultItem: %v", sprite))
		resultCh <- InfoResultItem{
			Type: st,
			Item: sprite,
		}
	}
	resultCh <- InfoResultItem{
		Type: share.InfoItemTypeDone,
		Item: struct{}{},
	}
}

// random spawn some events
func (l *Land) spawnEvents() {
	tick := time.Tick(200 * time.Millisecond)

	loop := 0
	l.jumpGhost()

	for _ = range tick {
		loop++
		if loop%25 == 0 {
			l.jumpGhost()
		}

		choice := rand.Int() % 1

		var t string
		var i interface{}
		switch choice {
		case 0:
			t = share.EventTypeMove

			a, err := l.getAlice()
			log.Debug(fmt.Sprintf("l.getAlice: %v", a))
			if err != nil {
				continue
			}

			g, err := l.getGhost()
			log.Debug(fmt.Sprintf("l.getGhost: %v", g))
			if err != nil {
				continue
			}

			dir, err := computeAliceMove(a.P.X, a.P.Y, g.P.X, g.P.Y)
			if err != nil {
				continue
			}

			l.moveAlice(dir)

			i = share.SpriteMove{Name: "Alice", Direction: dir}

		case 1:
			t = share.EventTypeAdd
			i = share.SpriteAdd{}
		case 2:
			t = share.EventTypeDelete
			i = share.SpriteDelete{}
		}
		event := Event{Type: t, Item: i}

		// send to channel, the other end is alice eventLoop waiting.
		if l.config.EventCh != nil {
			l.config.EventCh <- event
		}
	}

}

func (l *Land) jumpGhost() {
	log.Debug(fmt.Sprintf("jumpGhost"))

	var sprites []share.Sprite

	l.spritesLock.Lock()
	defer l.spritesLock.Unlock()

	point := l.findEmptyPoint()
	ghost := share.Human{Name: "Ghost"}

	for _, s := range l.sprites {
		if isGhost(s) {
			point.X = rand.Int() % gCol
			point.Y = rand.Int() % gRow

		} else {
			sprites = append(sprites, s)

		}
	}

	// send event
	if l.config.EventCh != nil {
		t := share.EventTypeJump
		i := share.SpriteJump{Name: "Ghost", X: point.X, Y: point.Y}

		event := Event{Type: t, Item: i}
		l.config.EventCh <- event
	}

	ghost.PutToPoint(point)
	sprites = append(sprites, ghost)

	l.sprites = sprites
}

func computeAliceMove(srcX, srcY, dstX, dstY int) (dir share.MoveDirection, err error) {

	var dirs []share.MoveDirection
	if srcX > dstX {
		dirs = append(dirs, share.MoveLeft)
	} else if srcX < dstX {
		dirs = append(dirs, share.MoveRight)
	}

	if srcY > dstY {
		dirs = append(dirs, share.MoveUp)
	} else if srcY < dstY {
		dirs = append(dirs, share.MoveDown)
	}

	if len(dirs) == 0 {
		err = errors.New("no need to move")
	} else {
		dir = dirs[rand.Int()%len(dirs)]
	}

	return dir, err
}
