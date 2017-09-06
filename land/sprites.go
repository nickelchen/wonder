package land

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/nickelchen/wonder/share"
	log "github.com/sirupsen/logrus"
)

func (l *Land) randPoint() share.Point {
	return share.Point{X: rand.Int() % gCol, Y: rand.Int() % gRow}
}

func initTiles() [][]share.Tile {
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

func initSprites() []share.Sprite {
	var sprites []share.Sprite

	return sprites
}

func isAlice(sprite share.Sprite) bool {
	h, ok := sprite.(share.Human)
	return ok && h.Name == "Alice"
}

func isRabbit(sprite share.Sprite) bool {
	h, ok := sprite.(share.Animal)
	return ok && h.Name == "Rabbit"
}

func (l *Land) aliceEnter() {
	point := l.randPoint()
	alice := share.Human{Name: "Alice"}
	alice.PutPoint(point)
	l.sprites = append(l.sprites, alice)
}

func (l *Land) aliceInfo() (share.Human, error) {
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

func (l *Land) aliceMove(dir share.MoveDirection) {
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
			a.PutPoint(p)
			sprites = append(sprites, a)

		} else {
			sprites = append(sprites, s)

		}
	}
	l.sprites = sprites
}

func (l *Land) rabbitInfo() (share.Animal, error) {
	l.spritesLock.Lock()
	defer l.spritesLock.Unlock()

	for _, s := range l.sprites {
		if isRabbit(s) {
			a, _ := s.(share.Animal)
			return a, nil
		}
	}
	return share.Animal{}, errors.New("can not find rabbit")
}

func (l *Land) rabbitJump() share.Point {
	l.spritesLock.Lock()
	defer l.spritesLock.Unlock()

	log.Debug(fmt.Sprintf("rabbitJump"))

	var sprites []share.Sprite
	point := l.randPoint()
	rabbit := share.Animal{Name: "Rabbit"}

	for _, s := range l.sprites {
		if isRabbit(s) {
			point.X = rand.Int() % gCol
			point.Y = rand.Int() % gRow
		} else {
			sprites = append(sprites, s)
		}
	}

	rabbit.PutPoint(point)
	sprites = append(sprites, rabbit)

	l.sprites = sprites
	return point
}

func moveDirection(srcX, srcY, dstX, dstY int) (dir share.MoveDirection, err error) {

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
