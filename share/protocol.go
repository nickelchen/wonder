package share

type RequestHeader struct {
	Seq     uint64
	Command string
}

type ResponseHeader struct {
	Seq   uint64
	Error string
}

//
// Plant command
//
type PlantType string

const (
	PlantTree   = "tree"
	PlantFlower = "flower"
	PlantGrass  = "grass"
)

type PlantRequest struct {
	What   PlantType
	Color  string
	Number int
}

type PlantResponse struct {
	Succ int
	Fail int
}

type ListServersRequest struct {
}

type ListServersResponse struct {
	Servers []string
}

//
// Info command
//
type InfoRequest struct {
}

type InfoResponse struct {
}

const (
	InfoItemTypeTile   = "tiles"
	InfoItemTypeTree   = "trees"
	InfoItemTypeFlower = "flowers"
	InfoItemTypeGrass  = "grass"
	InfoItemTypeHuman  = "human"
	InfoItemTypeAnimal = "animal"
	InfoItemTypeDone   = "done"
)

type InfoResponseObj struct {
	Type    string
	Payload []byte
}

//
// Subscribe command
//
type SubscribeRequest struct {
}
type SubscribeResponse struct {
}

const (
	EventTypeMove   = "move"
	EventTypeJump   = "jump"
	EventTypeAdd    = "add"
	EventTypeDelete = "delete"
)

type EventResponseObj struct {
	Type    string
	Payload []byte
}

//
// command list
//
const (
	PlantCommand     = "PlantCommand"
	InfoCommand      = "InfoCommand"
	SubscribeCommand = "SubscribeCommand"
)
