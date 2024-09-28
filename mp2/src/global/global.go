package global
// Defince the global variables and structs for SWIM protocol
import (
    "sync"
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

var INTRODUCER_ADDRESS = "fa24-cs425-6605.cs.illinois.edu:8081"
var PORT = "8081"
var PROTOCOL_PERIOD = 5
var TIMEOUT_PERIOD = 1
var SUSPECT_TIMEOUT = 30
var DEAD_TIMEOUT = 60
var ALIVE_TIMEOUT = 10

var Introducer = false
var Nodes = make(map[string]NodeInfo)
var GossipNodes = make(map[string]GossipNode)
var NodesMutex sync.Mutex
var GossipNodesMutex sync.Mutex
var Id = ""
var SelfAddress = ""


type State int

const (
    Suspected State = iota
    Alive
    Down
)

