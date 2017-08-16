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
	Steps     int
}

type SpriteAdd struct {
}

type SpriteDelete struct {
}

type Human struct {
	SpriteBase
	name string
}

type Tree struct {
	SpriteBase
}

type Grass struct {
	SpriteBase
}

type Flower struct {
	SpriteBase
	color string
}
