package land

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"

	"nickelchen/wonder/share"
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
	tiles   [][]share.Tile
	sprites []share.Sprite
	config  *Config
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

func (l *Land) findEmptyPoint() share.Point {
	return share.Point{X: rand.Int() % gRow, Y: rand.Int() % gCol}
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
	tick := time.Tick(1 * time.Second)
	for _ = range tick {
		choice := rand.Int() % 3

		var t string
		var i interface{}
		switch choice {
		case 0:
			t = share.EventTypeMove
			i = share.SpriteMove{}
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
