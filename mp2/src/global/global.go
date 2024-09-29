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


const (
    SWIM_PROROCOL = 0
    SWIM_SUSPIECT_PROROCOL = 1
)


var Nodes = make(map[string]NodeInfo)
var GossipNodes = make(map[string]GossipNode)
var Protocol = SWIM_PROROCOL
var SuspectedNodes = make(map[string]time.Time)
var Incarnation = 0

type State int

const (
    Suspected State = iota
    Alive
    Down
    Join
)

var COMMAND_PORT = "8082"
var Cluster = []string{
    "fa24-cs425-6601.cs.illinois.edu:" + COMMAND_PORT,
    "fa24-cs425-6602.cs.illinois.edu:" + COMMAND_PORT,
    "fa24-cs425-6603.cs.illinois.edu:" + COMMAND_PORT,
    "fa24-cs425-6604.cs.illinois.edu:" + COMMAND_PORT,
    "fa24-cs425-6605.cs.illinois.edu:" + COMMAND_PORT,
    "fa24-cs425-6606.cs.illinois.edu:" + COMMAND_PORT,
    "fa24-cs425-6607.cs.illinois.edu:" + COMMAND_PORT,
    "fa24-cs425-6608.cs.illinois.edu:" + COMMAND_PORT,
    "fa24-cs425-6609.cs.illinois.edu:" + COMMAND_PORT,
    "fa24-cs425-6610.cs.illinois.edu:" + COMMAND_PORT,
}
