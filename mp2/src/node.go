package main

import (
	"flag"
    "fmt"
	"math/rand"
    "net"
    "os"
    "strconv"
	"sync"
    "time"
)

type State int

const (
    Suspected State = iota
    Alive
    Down
)

// Define the struct for the map values
type NodeInfo struct {
	ID 	string
    Address string
    State   State
}

var INTRODUCER_ADDRESS = "fa24-cs425-6605.cs.illinois.edu:8081"
var PORT = "8081"
var PROTOCOL_PERIOD = 5
var TIMEOUT_PERIOD = 1
var SUSPECT_TIMEOUT = 30

var Introducer = false
var Nodes = make(map[string]NodeInfo)
var NodesMutex sync.Mutex
var Id = ""

func main(){
	introducerFlag := flag.Bool("introducer", false, "Set this flag to true if this node is the introducer")

    flag.Parse()

	Introducer = *introducerFlag

	hostname, err := os.Hostname()
    if err != nil {
        fmt.Println("Error getting hostname:", err)
        return
    }

	Id = hostname + ":" + PORT + ":" + strconv.FormatInt(time.Now().UnixNano(), 10)
	fmt.Println("Node ID:", Id)

	if Introducer {
		Nodes[Id] = NodeInfo{ID: Id, Address: hostname + ":" + PORT, State: Alive}
	} else {
		dialIntroducer()
	}
	// Need additional logic to handle the case where the introducer is down

	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup
    wg.Add(2)

	go func() {
        defer wg.Done()
        startServer()
    }()
	
	go func() {
        defer wg.Done()
        startClient()
    }()

	wg.Wait()

}

func dialIntroducer() {
    // TODO: implement the logic to dial the introducer and get the list of nodes
    fmt.Println("Dialing introducer...")
}

func startServer() {
	addr := ":" + PORT
    conn, err := net.ListenPacket("udp", addr)
    if err != nil {
        fmt.Println("Error starting server:", err)
        return
    }
    defer conn.Close()

    fmt.Println("Server started on", addr)

    buffer := make([]byte, 4096)
    for {
        n, addr, err := conn.ReadFrom(buffer)
        if err != nil {
            fmt.Println("Error reading from connection:", err)
            continue
        }
        message := string(buffer[:n])
        // fmt.Printf("Received %d bytes from %s: %s\n", n, addr, message)

        if strings.HasPrefix(message, "PING") {
            // TODO: response to ping
        } else if strings.HasPrefix(message, "RELAY") {
            // TODO: Relay the message to the target node
        } else if strings.HasPrefix(message, "JOIN") {
			if not Introducer continue
			NodesMutex.Lock()
            // TODO: Add the new node to the list of nodes
            NodesMutex.Unlock()
		} else if strings.HasPrefix(message, "ACK") {
			// TODO: This is the ack from relay message, send ack back to the sender.
		} else {
			fmt.Println("Unknown message:", message)
		}
    }
}

func startClient() {
	fmt.Println("Starting client...")
	curNode := 0

	var nodesArray []NodeInfo
	NodesMutex.Lock()
    for _, node := range Nodes {
        nodesArray = append(nodesArray, node)
    }
	NodesMutex.Unlock()

	rand.Shuffle(len(nodesArray), func(i, j int) {
		nodesArray[i], nodesArray[j] = nodesArray[j], nodesArray[i]
	})

	ticker := time.NewTicker(PROTOCOL_PERIOD * time.Second) // Ping every PROTOCOL_PERIOD seconds
    defer ticker.Stop()

	for {
		select {
        case <-ticker.C:
            if curNode >= len(nodesArray) {
                curNode = 0
                NodesMutex.Lock()
                nodesArray = []NodeInfo{} // Clear the array
                for _, node := range Nodes {
                    nodesArray = append(nodesArray, node)
                }
                NodesMutex.Unlock()

                rand.Shuffle(len(nodesArray), func(i, j int) {
                    nodesArray[i], nodesArray[j] = nodesArray[j], nodesArray[i]
                })
            }
            node := nodesArray[curNode]
            pingServer(curNode)
            curNode++
        }
	}
}

func pingServer(address string) {
    conn, err := net.DialTimeout("udp", address, TIMEOUT_PERIOD * time.Second)
    if err != nil {
        fmt.Printf("Failed to ping %s: %v\n", address, err)
        // TODO: Handle the case where the node is down
    }
    defer conn.Close()

    // TODO: Send a PING message to the server

    // Set a read deadline for the response
    conn.SetReadDeadline(time.Now().Add(TIMEOUT_PERIOD * time.Second))

    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Printf("No response from %s: %v\n", address, err)
        // TODO: Handle the case where the node is down
    }

    response := string(buffer[:n])
    // TODO: Parse the response and update the state of the node
}