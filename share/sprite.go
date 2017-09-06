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

func (s *SpriteBase) PutPoint(p Point) {
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

type Animal struct {
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
	Animals []Animal

	moveEventsCh   chan SpriteMove
	jumpEventsCh   chan SpriteJump
	addEventsCh    chan SpriteAdd
	deleteEventsCh chan SpriteDelete
}

func NewGameBoard() *GameBoard {
	board := GameBoard{
		moveEventsCh:   make(chan SpriteMove, 255),
		jumpEventsCh:   make(chan SpriteJump, 255),
		addEventsCh:    make(chan SpriteAdd, 255),
		deleteEventsCh: make(chan SpriteDelete, 255),
	}

	go board.pollingEvents()

	return &board
}

func (board GameBoard) MoveEventsCh() chan SpriteMove {
	return board.moveEventsCh
}
func (board GameBoard) JumpEventsCh() chan SpriteJump {
	return board.jumpEventsCh
}
func (board GameBoard) AddEventsCh() chan SpriteAdd {
	return board.addEventsCh
}
func (board GameBoard) DeleteEventsCh() chan SpriteDelete {
	return board.deleteEventsCh
}

func (board *GameBoard) pollingEvents() {
	for {
		select {
		case event := <-board.moveEventsCh:
			// for now, only human can move.
			var humans []Human
			for _, h := range board.Humans {
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
					h.PutPoint(p)
				}
				humans = append(humans, h)
			}

			board.Humans = humans

		case event := <-board.jumpEventsCh:
			// for now only animal can jump
			var animals []Animal
			var this Animal
			for _, a := range board.Animals {
				if a.Name == event.Name {
					this = a
					continue
				}
				animals = append(animals, a)
			}

			this.Name = event.Name
			p := this.P
			p.X = event.X
			p.Y = event.Y

			this.PutPoint(p)

			animals = append(animals, this)
			board.Animals = animals

		case <-board.addEventsCh:
			//
		case <-board.deleteEventsCh:
			//
		}
	}
}
