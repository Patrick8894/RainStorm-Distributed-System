package global
// Defince the global variables and structs for SWIM protocol
import (
    "time"
)

// Define the struct for the map values
type NodeInfo struct {
	ID 	string
    Address string
    State   State
}

type GossipNode struct {
    ID string
    Address string
    State State
    Incarnation int
    Time time.Time
}




var Nodes = make(map[string]NodeInfo)
var GossipNodes = make(map[string]GossipNode)


type State int

const (
    Suspected State = iota
    Alive
    Down
    Join
)

