package land

import (
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
	l.tiles = initTiles()
	l.sprites = initSprites()

	l.aliceEnter()

	go l.spawnFakeEvents()

	log.Info("land/land.go Spread()")
	return 0
}

func (l *Land) Shrink() {
	log.Info("land/land.go Shrink()")
}

func (l *Land) Plant(params *PlantParams) (*PlantResult, error) {
	log.Info("land/land.go Plant()")

	var s share.Sprite
	for i := 0; i < params.Number; i++ {
		point := l.randPoint()
		switch params.What {
		case share.PlantTree:
			o := share.Tree{}
			o.PutPoint(point)
			s = o

		case share.PlantFlower:
			o := share.Flower{}
			o.PutPoint(point)
			s = o

		case share.PlantGrass:
			o := share.Grass{}
			o.PutPoint(point)
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
func (l *Land) spawnFakeEvents() {
	tick := time.Tick(200 * time.Millisecond)

	loop := 0

	for _ = range tick {
		loop++
		// every 25 loops, rabbit jump once.
		if loop%25 == 0 {
			newPoint := l.rabbitJump()
			l.sendEvent(
				share.EventTypeJump,
				share.SpriteJump{Name: "Rabbit", X: newPoint.X, Y: newPoint.Y})

		}

		// for now, only spwan 0 type event: EventTypeMove
		choice := rand.Int() % 1

		// type of event
		var t string
		// the item associated with the event
		var i interface{}

		switch choice {
		case 0:
			t = share.EventTypeMove

			a, err := l.aliceInfo()
			log.Debug(fmt.Sprintf("l.aliceInfo: %v", a))
			if err != nil {
				continue
			}

			r, err := l.rabbitInfo()
			log.Debug(fmt.Sprintf("l.rabbitInfo: %v", r))
			if err != nil {
				continue
			}

			dir, err := moveDirection(a.P.X, a.P.Y, r.P.X, r.P.Y)
			if err != nil {
				continue
			}

			l.aliceMove(dir)

			i = share.SpriteMove{Name: "Alice", Direction: dir}

		case 1:
			t = share.EventTypeAdd
			i = share.SpriteAdd{}
		case 2:
			t = share.EventTypeDelete
			i = share.SpriteDelete{}
		}

		l.sendEvent(t, i)
	}
}

func (l *Land) sendEvent(eventType string, eventItem interface{}) {
	event := Event{Type: eventType, Item: eventItem}
	// send to channel, on the other end, alice eventLoop is waiting.
	if l.config.EventCh != nil {
		l.config.EventCh <- event
	}
}
