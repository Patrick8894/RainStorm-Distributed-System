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
    pb "mp2/proto"
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

func Address_to_ID(address string) string {
    NodesMutex.Lock()
    for _, node := range Nodes {
        if node.Address == address {
            return node.ID
        }
    }
    NodesMutex.Unlock()
}

func main(){
	introducerFlag := flag.Bool("introducer", false, "Set this flag to true if this node is the introducer")

    flag.Parse()

	Introducer = *introducerFlag

	hostname, err := os.Hostname()
    if err != nil {
        fmt.Println("Error getting hostname:", err)
        return
    }

    SelfAddress = hostname + ":" + PORT
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

    conn, err := net.Dial("udp", INTRODUCER_ADDRESS)
    if err != nil {
        fmt.Println("Error dialing introducer:", err)
        return
    }
    defer conn.Close()

    // Create a SWIMMessage and send it to the introducer
    message := &pb.SWIMMessage{
        Type:   pb.SWIMMessage_JOIN,
        Sender: SelfAddress,
        Target: INTRODUCER_ADDRESS
    }

    data, err := proto.Marshal(message)
    if err != nil {
        fmt.Println("Failed to marshal message:", err)
        return
    }

    _, err = conn.Write(data)
    if err != nil {
        fmt.Println("Failed to send message:", err)
        return
    }

    fmt.Println("Message sent to introducer")
    
    // Set a read deadline for the response
    conn.SetReadDeadline(time.Now().Add(TIMEOUT_PERIOD * time.Second))

    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Println("No response from introducer:", err)
        return
    }

    var message pb.SWIMMessage
    err = proto.Unmarshal(buffer[:n], &message)
    if err != nil {
        fmt.Println("Failed to unmarshal message:", err)
        return
    }

    NodesMutex.Lock()
    for _, member := range message.Membership {
        if member.Status == "Alive" {
            Nodes[member.MemberID] = NodeInfo{ID: member.MemberID, Address: member.Address, State: Alive}
        }
        else member.Status == "Suspected" {
            Nodes[member.MemberID] = NodeInfo{ID: member.MemberID, Address: member.Address, State: Suspected}
        }
    }
    NodesMutex.Unlock()

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
        
        // Unmarshal the protobuf message
        var message pb.SWIMMessage
        err = proto.Unmarshal(buf[:n], &message)
        if err != nil {
            log.Println("Failed to unmarshal message:", err)
            return
        }

        //  received message
        fmt.Printf("Received message: %+v\n", message)

        // fmt.Printf("Received %d bytes from %s: %s\n", n, addr, message)

        if message.Type == pb.SWIMMessage_PING {
            // Response to ping
            // Create a SWIMMessage to send
            response := &pb.SWIMMessage{
                Type:   pb.SWIMMessage_PONG,
                Sender: SelfAddress,
                Target: message.Sender
            }

            // Serialize the message using protobuf
            data, err := proto.Marshal(response)
            if err != nil {
                fmt.Println("Failed to marshal message:", err)
                return
            }

            // Send the message to the server
            _, err = conn.WriteTo(data, addr)
            if err != nil {
                fmt.Println("Failed to send message:", err)
                return
            }

            fmt.Println("Message sent to server")

        } else if message.Type == pb.SWIMMessage_INDIRECT_PING {
            // Relay the message to the target node
            targetAddr, err := net.ResolveUDPAddr("udp", message.Target)
            if err != nil {
                fmt.Println("Failed to resolve target address:", err)
                return
            }

            // Create a PING message with the same sender and target
            pingMessage := &pb.SWIMMessage{
                Type: pb.SWIMMessage_PING,
                Sender: message.Sender,
                Target: message.Target,
            }

            // Serialize the PING message using protobuf
            data, err := proto.Marshal(pingMessage)
            if err != nil {
                fmt.Println("Failed to marshal PING message:", err)
                return
            }

            // Send the PING message to the target node
            _, err = conn.WriteTo(data, targetAddr)
            if err != nil {
                fmt.Println("Failed to send PING message:", err)
                return
            }

            fmt.Println("PING message sent to target node")

        } else if message.Type == pb.SWIMMessage_JOIN {
			if not Introducer continue
			NodesMutex.Lock()
            // Add the new node to the list of nodes
            for _, member := range message.Membership {
                if member.Status == "Alive" {
                    Nodes[member.MemberID] = NodeInfo{ID: member.MemberID, Address: member.Address, State: Alive}
                }
                else member.Status == "Suspected" {
                    Nodes[member.MemberID] = NodeInfo{ID: member.MemberID, Address: member.Address, State: Suspected}
                }
            }

            NodesMutex.Unlock()
		} else message.Type == pb.SWIMMessage_PONG {
			// This is the ack from relay message, send ack back to the sender.
            targetAddr, err := net.ResolveUDPAddr("udp", message.Target)
            if err != nil {
                fmt.Println("Failed to resolve target address:", err)
                return
            }

            // Create a SWIMMessage to send
            response := &pb.SWIMMessage{
                Type:   pb.SWIMMessage_PONG,
                Sender: SelfAddress,
                Target: message.Sender,
            }

            // Serialize the message using protobuf
            data, err := proto.Marshal(response)
            if err != nil {
                fmt.Println("Failed to marshal message:", err)
                return
            }

            // Send the message to the server
            _, err = conn.WriteTo(data, targetAddr)
            if err != nil {
                fmt.Println("Failed to send message:", err)
                return
            }

            fmt.Println("Message sent to server")
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
            pingServer(node)
            curNode++
        }
	}
}

func getRandomNodes(n int) []NodeInfo {
    NodesMutex.Lock()

    keys := make([]string, 0, len(Nodes))
    
    for k := range Nodes {
        keys = append(keys, k)
    }

    if len(keys) < n {
        n = len(keys)
    }

    rand.Shuffle(len(keys), func(i, j int) {
        keys[i], keys[j] = keys[j], keys[i]
    })

    randomNodes := make([]NodeInfo, 0, n)
    for i := 0; i < n; i++ {
        randomNodes = append(randomNodes, Nodes[keys[i]])
    }

    NodesMutex.Unlock()
    return randomNodes
}

func pingIndirect(node NodeInfo) {
    // TODO: implement the logic to ping the indirect node
    randomNodes := getRandomNodes(3)
    resultChan := make(chan bool, len(randomNodes))
    var wg sync.WaitGroup

    for _, randomNode := range randomNodes {
        wg.Add(1)
        go func(rNode NodeInfo) {
            defer wg.Done()

            indirectPingMessage := &pb.SWIMMessage{
                Type:   pb.SWIMMessage_INDIRECT_PING,
                Sender: SelfAddress,
                Target: node.Address, // Assuming NodeInfo has an Address field
            }

            // Serialize the INDIRECT_PING message using protobuf
            data, err := proto.Marshal(indirectPingMessage)
            if err != nil {
                fmt.Println("Failed to marshal INDIRECT_PING message:", err)
                resultChan <- false
                return
            }
            
            // Resolve the address of the random node
            randomNodeAddr, err := net.ResolveUDPAddr("udp", rNode.Address)
            if err != nil {
                fmt.Println("Failed to resolve random node address:", err)
                resultChan <- false
                return
            }
            
            // Send the INDIRECT_PING message to the random node
            _, err = conn.WriteTo(data, randomNodeAddr)
            if err != nil {
                fmt.Println("Failed to send INDIRECT_PING message:", err)
                resultChan <- false
                return
            }

            // Buffer to read the response
            buffer := make([]byte, 4096)

            // Set a read deadline
            conn.SetReadDeadline(time.Now().Add(5 * time.Second))

            // Read from the connection
            n, _, err := conn.ReadFrom(buffer)
            if err != nil {
                fmt.Println("Failed to read from connection:", err)
                resultChan <- false
                return
            }

            // Unmarshal the protobuf message
            var responseMessage pb.SWIMMessage
            err = proto.Unmarshal(buffer[:n], &responseMessage)
            if err != nil {
                fmt.Println("Failed to unmarshal response message:", err)
                resultChan <- false
                return
            }

            // Check if the message type is PONG
            if responseMessage.Type == pb.SWIMMessage_PONG {
                resultChan <- true
            } else {
                resultChan <- false
            }
        }(randomNode)
    }

    // Close the result channel once all goroutines are done
    go func() {
        wg.Wait()
        close(resultChan)
    }()

    // Return true if any of the goroutines succeeded
    for result := range resultChan {
        if result {
            return true
        }
    }

    return false
}

func pingServer(node NodeInfo) {  
    conn, err := net.DialTimeout("udp", node.address, TIMEOUT_PERIOD * time.Second)
    if err != nil {
        fmt.Printf("Failed to ping %s: %v\n", node.address, err)
        // TODO: Handle the case where the direct node is down

        rst := pingIndirect(node)

        if rst == false {
            // delete the node from the Nodes list
            NodesMutex.Lock()
            delete(Nodes, node.ID)
            NodesMutex.Unlock()

            // add the node to the GossipNodes list
            GossipNodesMutex.Lock()
            GossipNodes[node.ID] = GossipNode{ID: node.ID, Address: node.address, State: Down, Incarnation: 0}
            GossipNodesMutex.Unlock()
        }
        return
    }
    defer conn.Close()

    // TODO: Send a PING message to the server
    message := &pb.SWIMMessage{
        Type:   pb.SWIMMessage_PING,
        Sender: hoostname + ":" + PORT,
        Target: node.address,
    }

    data, err := proto.Marshal(message)
    if err != nil {
        fmt.Printf("Failed to marshal message: %v\n", err)
        return
    }

    _, err = conn.Write(data)
    if err != nil {
        fmt.Printf("Failed to send message: %v\n", err)
        return
    }


    // Set a read deadline for the response
    conn.SetReadDeadline(time.Now().Add(TIMEOUT_PERIOD * time.Second))

    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Printf("No response from %s: %v\n", node.address, err)
        // TODO: Handle the case where the node is down

        rst := pingIndirect(node)
        
        if rst == false {
            // delete the node from the Nodes list
            NodesMutex.Lock()
            delete(Nodes, node.ID)
            NodesMutex.Unlock()

            // add the node to the GossipNodes list
            GossipNodesMutex.Lock()
            GossipNodes[node.ID] = GossipNode{ID: node.ID, Address: node.address, State: Down, Incarnation: 0}
            GossipNodesMutex.Unlock()
        }
        return
    }


    // TODO: Parse the response and update the state of the node
    var response pb.SWIMMessage
    err = proto.Unmarshal(buffer[:n], &response)
    if err != nil {
        fmt.Printf("Failed to unmarshal message: %v\n", err)
        return
    }

    // Update the state of the node
    for _, Membership := range response.MembershipInfo {
        if Membership.Status == "Down" {
            // delete the node from the Nodes list
            NodesMutex.Lock()
            delete(Nodes, node.ID)
            NodesMutex.Unlock()

            // add the node to the GossipNodes list
            GossipNodesMutex.Lock()
            _, exists := GossipNodes[node.ID]; 
            if exists {
                // if gossipNodes is not empty, check if the node is already in the list
                if GossipNodes[node.ID].Time < time.Now().Add(-DEAD_TIMEOUT * time.Second) {
                    delete(GossipNodes, node.ID)
                }
            }
            else {
                // not exist
                GossipNodes[node.ID] = GossipNode{ID: node.ID, Address: node.address, State: Down, Incarnation: 0, Time: time.Now()}
            }
            GossipNodesMutex.Unlock()
            return
        }
    }   


}