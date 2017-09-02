package share

type Point struct {
	X int
	Y int
}

type Tile struct {
	Gradient int
}

type Sprite interface {
	GetPoint() Point
	MovesToPoint(dstPoint Point) []SpriteMove
	ToggleVisible()
}

type SpriteBase struct {
	P       Point
	Visible bool
}

func (s SpriteBase) GetPoint() Point {
	return s.P
}

func (s SpriteBase) MovesToPoint(dstPoint Point) []SpriteMove {
	moves := []SpriteMove(nil)
	return moves
}

func (s SpriteBase) ToggleVisible() {
	s.Visible = !s.Visible
}

func (s *SpriteBase) PutToPoint(p Point) {
	s.P = p
}

type MoveDirection int

const (
	MoveUp MoveDirection = iota
	MoveDown
	MoveLeft
	MoveRight
)

type SpriteMove struct {
	Direction MoveDirection
	Name      string
}

type SpriteJump struct {
	X    int
	Y    int
	Name string
}

type SpriteAdd struct {
}

type SpriteDelete struct {
}

type Human struct {
	SpriteBase
	Name string
}

type Tree struct {
	SpriteBase
}

type Grass struct {
	SpriteBase
}

type Flower struct {
	SpriteBase
	Color string
}

type GameBoard struct {
	Tiles   [][]Tile
	Trees   []Tree
	Flowers []Flower
	Grasses []Grass
	Humans  []Human

	moveEventsCh   chan SpriteMove
	jumpEventsCh   chan SpriteJump
	addEventsCh    chan SpriteAdd
	deleteEventsCh chan SpriteDelete
}

func NewGameBoard() *GameBoard {
	gb := GameBoard{
		moveEventsCh:   make(chan SpriteMove, 255),
		jumpEventsCh:   make(chan SpriteJump, 255),
		addEventsCh:    make(chan SpriteAdd, 255),
		deleteEventsCh: make(chan SpriteDelete, 255),
	}

	go gb.pollingEvents()

	return &gb
}

func (gb GameBoard) MoveEventsCh() chan SpriteMove {
	return gb.moveEventsCh
}
func (gb GameBoard) JumpEventsCh() chan SpriteJump {
	return gb.jumpEventsCh
}
func (gb GameBoard) AddEventsCh() chan SpriteAdd {
	return gb.addEventsCh
}
func (gb GameBoard) DeleteEventsCh() chan SpriteDelete {
	return gb.deleteEventsCh
}

func (gb *GameBoard) pollingEvents() {
	for {
		select {
		case event := <-gb.moveEventsCh:
			// for now, only human can move.
			var humans []Human
			for _, h := range gb.Humans {
				if h.Name == event.Name {
					p := h.P
					switch event.Direction {
					case MoveUp:
						p.Y -= 1
					case MoveDown:
						p.Y += 1
					case MoveRight:
						p.X += 1
					case MoveLeft:
						p.X -= 1
					}
					h.PutToPoint(p)
				}
				humans = append(humans, h)
			}

			gb.Humans = humans
		case event := <-gb.jumpEventsCh:
			// for now human can jump
			var humans []Human
			for _, h := range gb.Humans {
				if h.Name == event.Name {
					p := h.P
					p.X = event.X
					p.Y = event.Y

					h.PutToPoint(p)
				}
				humans = append(humans, h)
			}

			gb.Humans = humans

		case <-gb.addEventsCh:
			//
		case <-gb.deleteEventsCh:
			//
		}
	}
}
