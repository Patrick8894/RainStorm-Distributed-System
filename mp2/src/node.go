package main

import (
    "encoding/json"
	"flag"
    "fmt"
	"math/rand"
    "net"
    "os"
    "strconv"
	"sync"
    "time"
    pb "mp2/proto"
    "mp2/src/utils"
    "mp2/src/global"
    "google.golang.org/protobuf/proto"
)


// func Address_to_ID(address string) string {
//     NodesMutex.Lock()
//     for _, node := range Nodes {
//         if node.Address == address {
//             return node.ID
//         }
//     }
//     NodesMutex.Unlock()
// }

var INTRODUCER_ADDRESS = "fa24-cs425-6605.cs.illinois.edu:8081"
var PORT = "8081"
var PROTOCOL_PERIOD = 0.25
var TIMEOUT_PERIOD = 0.1
var SUSPECT_TIMEOUT = 30

var Introducer = false
var NodesMutex sync.Mutex
var GossipNodesMutex sync.Mutex
var Id = ""
var SelfAddress = ""

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
		global.Nodes[Id] = global.NodeInfo{ID: Id, Address: hostname + ":" + PORT, State: global.Alive}
	} else {
		dialIntroducer()
	}
	// Need additional logic to handle the case where the Introducer is down

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
    // If the node is not the introducer, then send a JOIN message to the introducer
    // TODO: implement the logic to dial the Introducer and get the list of nodes
    fmt.Println("Dialing introducer...")

    conn, err := net.Dial("udp", INTRODUCER_ADDRESS)
    if err != nil {
        fmt.Println("Error dialing introducer:", err)
        return
    }
    defer conn.Close()

    memberone := &pb.MembershipInfo{
        MemberID: Id,
    }
    // Create a SWIMMessage and send it to the introducer
    message := &pb.SWIMMessage{
        Type:   pb.SWIMMessage_JOIN,
        Sender: SelfAddress,
        Target: INTRODUCER_ADDRESS,
        Membership: []*pb.MembershipInfo{memberone},
    }
    
    fmt.Println("Sending JOIN message to Introducer from", SelfAddress)
    data, err := proto.Marshal(message)
    if err != nil {
        fmt.Printf("Failed to marshal message: %v\n", err)
        return
    }
    _, err = conn.Write(data) // Use Write method instead of WriteTo
    if err != nil {
        fmt.Printf("Failed to send message: %v\n", err)
    }
    
    // Set a read deadline for the response
    conn.SetReadDeadline(time.Now().Add(time.Duration(TIMEOUT_PERIOD) * time.Second))


    // Use Json to unmarshal when receiving the whole NodeList
    buffer := make([]byte, 4096)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Println("No response from introducer:", err)
        return
    }

    var response map[string]global.NodeInfo
    err = json.Unmarshal(buffer[:n], &response)
    if err != nil {
        fmt.Println("Failed to unmarshal message:", err)
        return
    }

    NodesMutex.Lock()
    global.Nodes = response
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
    conn.SetWriteDeadline(time.Now().Add(time.Duration(TIMEOUT_PERIOD) * time.Second))

    fmt.Println("Server started on", addr)

    buffer := make([]byte, 4096)
    for {
        n, addr, err := conn.ReadFrom(buffer)
        if err != nil {
            fmt.Println("Error reading from connection:", err)
            continue
        }

        var message pb.SWIMMessage
        err = proto.Unmarshal(buffer[:n], &message)
        if err != nil {
            fmt.Println("Failed to unmarshal message:", err)
            continue
        }

        handleGossip(message)

        //  received message
        fmt.Printf("Received message: %+v\n", message)

        if message.Type == pb.SWIMMessage_DIRECT_PING {
            // Response to ping
            // Create a SWIMMessage to send

            GossipNodesMutex.Lock()
            gossiplist := utils.GetGossiplist(global.GossipNodes)
            GossipNodesMutex.Unlock()

            response := &pb.SWIMMessage{
                Type:   pb.SWIMMessage_PONG,
                Sender: SelfAddress,
                Target: message.Sender,
                Membership : gossiplist,
            }
            
            data, err := proto.Marshal(response)
            if err != nil {
                fmt.Printf("Failed to marshal message: %v\n", err)
                return
            }
            _, err = conn.WriteTo(data, addr) // Use Write method instead of WriteTo
            if err != nil {
                fmt.Printf("Failed to send message: %v\n", err)
            }

        	fmt.Println("Message success ping to server")

        } else if message.Type == pb.SWIMMessage_INDIRECT_PING {
            // Relay the message to the target node
            targetAddr, err := net.ResolveUDPAddr("udp", message.Target)
            if err != nil {
                fmt.Println("Failed to resolve target address:", err)
                return
            }
            
            GossipNodesMutex.Lock()
            gossiplist := utils.GetGossiplist(global.GossipNodes)
            GossipNodesMutex.Unlock()

            // Create a PING message with the same sender and target
            pingMessage := &pb.SWIMMessage{
                Type: pb.SWIMMessage_INDIRECT_PING,
                Sender: message.Sender,
                Target: message.Target,
                Membership: gossiplist,
            }
            
            data, err := proto.Marshal(pingMessage)
            if err != nil {
                fmt.Printf("Failed to marshal message: %v\n", err)
                return
            }
            _, err = conn.WriteTo(data, targetAddr) // Use Write method instead of WriteTo
            if err != nil {
                fmt.Printf("Failed to send message: %v\n", err)
            }

            fmt.Println("PING message success sent to target node")

        } else if message.Type == pb.SWIMMessage_JOIN {
			if !Introducer {
                continue
            }
			NodesMutex.Lock()
            // Add the new node to the list of nodes
            global.Nodes[message.Sender] = global.NodeInfo{ID: message.Membership[0].MemberID , Address: message.Sender, State: global.Alive}
            
            jsonData, err := json.Marshal(global.Nodes)
            if err != nil {
                fmt.Println("Error serializing data:", err)
                os.Exit(1)
            }
            NodesMutex.Unlock()
            _, err = conn.WriteTo(jsonData, addr)
            if err != nil {
                fmt.Println("Failed to response to JOIN message:", err)
                return
            }
            GossipNodesMutex.Lock()
            global.GossipNodes[message.Membership[0].MemberID] = global.GossipNode{ID: message.Membership[0].MemberID, Address: message.Sender, State: global.Join, Incarnation: 0, Time: time.Now()}
            GossipNodesMutex.Unlock()
		} else if message.Type == pb.SWIMMessage_PONG {
			// This is the ack from relay message, send ack back to the sender.
            targetAddr, err := net.ResolveUDPAddr("udp", message.Target)
            if err != nil {
                fmt.Println("Failed to resolve target address:", err)
                return
            }
            GossipNodesMutex.Lock()
            gossiplist := utils.GetGossiplist(global.GossipNodes)
            GossipNodesMutex.Unlock()

            // Create a SWIMMessage to send
            pongMessage := &pb.SWIMMessage{
                Type:   pb.SWIMMessage_PONG,
                Sender: message.Sender,
                Target: message.Target,
                Membership: gossiplist,
            }
            
            data, err := proto.Marshal(pongMessage)
            if err != nil {
                fmt.Printf("Failed to marshal message: %v\n", err)
                return
            }
            _, err = conn.WriteTo(data, targetAddr) // Use Write method instead of WriteTo
            if err != nil {
                fmt.Printf("Failed to send message: %v\n", err)
            }

            fmt.Println("Relay Message sent to server")

		} else {
			fmt.Println("Unknown message:", message)
		}
    }
}

func startClient() {
	fmt.Println("Starting client...")
	curNode := 0

	var nodesArray []global.NodeInfo
	NodesMutex.Lock()
    for _, node := range global.Nodes {
        nodesArray = append(nodesArray, node)
    }
	NodesMutex.Unlock()

	rand.Shuffle(len(nodesArray), func(i, j int) {
		nodesArray[i], nodesArray[j] = nodesArray[j], nodesArray[i]
	})

    if PROTOCOL_PERIOD <= 0 {
        fmt.Println("Invalid PROTOCOL_PERIOD value")
    }
	ticker := time.NewTicker(time.Duration(PROTOCOL_PERIOD * float64(time.Second) )) // Ping every PROTOCOL_PERIOD seconds
    defer ticker.Stop()

	for {
		select {
        case <-ticker.C:
            if curNode >= len(nodesArray) {
                curNode = 0
                NodesMutex.Lock()
                nodesArray = []global.NodeInfo{} // Clear the array
                for _, node := range global.Nodes {
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

func getRandomNodes(n int) []global.NodeInfo {
    NodesMutex.Lock()

    keys := make([]string, 0, len(global.Nodes))
    
    for k := range global.Nodes {
        keys = append(keys, k)
    }

    if len(keys) < n {
        n = len(keys)
    }

    rand.Shuffle(len(keys), func(i, j int) {
        keys[i], keys[j] = keys[j], keys[i]
    })

    randomNodes := make([]global.NodeInfo, 0, n)
    for i := 0; i < n; i++ {
        randomNodes = append(randomNodes, global.Nodes[keys[i]])
    }

    NodesMutex.Unlock()
    return randomNodes
}

func pingIndirect(node global.NodeInfo) bool {
    // TODO: implement the logic to ping the indirect node
    randomNodes := getRandomNodes(3)
    resultChan := make(chan bool, len(randomNodes))
    var wg sync.WaitGroup

    for _, randomNode := range randomNodes {
        wg.Add(1)
        go func(rNode global.NodeInfo) {
            defer wg.Done()
            
            // Construct GossipMessage
            GossipNodesMutex.Lock()
            gossiplist := utils.GetGossiplist(global.GossipNodes)
            GossipNodesMutex.Unlock()

            indirectPingMessage := &pb.SWIMMessage{
                Type:   pb.SWIMMessage_INDIRECT_PING,
                Sender: SelfAddress,
                Target: node.Address, // Assuming NodeInfo has an Address field
                Membership: gossiplist,
            }

            // Serialize the INDIRECT_PING message using protobuf
            data, err := proto.Marshal(indirectPingMessage)
            if err != nil {
                fmt.Println("Failed to marshal INDIRECT_PING message:", err)
                resultChan <- false
                return
            }

            conn, err := net.DialTimeout("udp", rNode.Address, 3 * time.Duration(TIMEOUT_PERIOD) * time.Second)
            if err != nil {
                fmt.Println("Failed to dial random node:", err)
                resultChan <- false
                return
            }
            defer conn.Close()

            // Set a write deadline for the connection
            conn.SetWriteDeadline(time.Now().Add(time.Duration(TIMEOUT_PERIOD) * time.Second))
            
            // Send the INDIRECT_PING message to the random node
            _, err = conn.Write(data)
            if err != nil {
                fmt.Println("Failed to send INDIRECT_PING message:", err)
                resultChan <- false
                return
            }

            // Buffer to read the response
            buffer := make([]byte, 4096)

            // Set a read deadline
            conn.SetReadDeadline(time.Now().Add(time.Duration(TIMEOUT_PERIOD) * time.Second))

            // Read from the connection
            n, err := conn.Read(buffer)
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
            handleGossip(responseMessage)

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

func pingServer(node global.NodeInfo) {  
    conn, err := net.DialTimeout("udp", node.Address, time.Duration(TIMEOUT_PERIOD) * time.Second)
    defer conn.Close()
    if err != nil {
        fmt.Printf("Failed to ping %s: %v\n", node.Address, err)
        // TODO: Handle the case where the direct node is down

        rst := pingIndirect(node)

        if rst == false {
            // delete the node from the Nodes list
            NodesMutex.Lock()
            delete(global.Nodes, node.ID)
            NodesMutex.Unlock()

            // add the node to the GossipNodes list
            GossipNodesMutex.Lock()
            global.GossipNodes[node.ID] = global.GossipNode{ID: node.ID, Address: node.Address, State: global.Down, Incarnation: 0, Time: time.Now()}
            GossipNodesMutex.Unlock()
        }
        return
    }

    // TODO: Send a PING message to the server
    GossipNodesMutex.Lock()
    gossiping := utils.GetGossiplist(global.GossipNodes)
    GossipNodesMutex.Unlock()

    message := &pb.SWIMMessage{
        Type:   pb.SWIMMessage_INDIRECT_PING,
        Sender: SelfAddress,
        Target: node.Address,
        Membership: gossiping,
    }

    data, err := proto.Marshal(message)
    if err != nil {
        fmt.Printf("Failed to marshal message: %v\n", err)
        return
    }
    _, err = conn.Write(data) // Use Write method instead of WriteTo
    if err != nil {
        fmt.Printf("Failed to send message: %v\n", err)
    }

    // Set a read deadline for the response
    conn.SetReadDeadline(time.Now().Add(time.Duration(TIMEOUT_PERIOD) * time.Second))

    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Printf("No response from %s: %v\n", node.Address, err)
        // TODO: Handle the case where the node is down

        rst := pingIndirect(node)
        
        if rst == false {
            // delete the node from the Nodes list
            NodesMutex.Lock()
            delete(global.Nodes, node.ID)
            NodesMutex.Unlock()

            // add the node to the GossipNodes list
            GossipNodesMutex.Lock()
            global.GossipNodes[node.ID] = global.GossipNode{ID: node.ID, Address: node.Address, State: global.Down, Incarnation: 0, Time: time.Now()}
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
    handleGossip(response)
       
}

func handleGossip(message pb.SWIMMessage) {
    for _, Membership := range message.Membership {
        if Membership.MemberStatus == utils.MapState(global.Down) {
            // delete the node from the Nodes list
            _, exists := global.Nodes[Membership.MemberID];
            if exists {
                NodesMutex.Lock()
                delete(global.Nodes, Membership.MemberID)
                NodesMutex.Unlock()

                // add the node to the GossipNodes list
                GossipNodesMutex.Lock()
                global.GossipNodes[Membership.MemberID] = global.GossipNode{ID: Membership.MemberID, Address: Membership.MemberAddress, State: global.Down, Incarnation: 0, Time: time.Now()}
                GossipNodesMutex.Unlock()
            }
        } else if Membership.MemberStatus == utils.MapState(global.Join) {
            // add the node to the Nodes list
            _, exists := global.Nodes[Membership.MemberID];
            if !exists {
                NodesMutex.Lock()
                global.Nodes[Membership.MemberID] = global.NodeInfo{ID: Membership.MemberID, Address: Membership.MemberAddress, State: global.Alive}
                NodesMutex.Unlock()

                GossipNodesMutex.Lock()
                global.GossipNodes[Membership.MemberID] = global.GossipNode{ID: Membership.MemberID, Address: Membership.MemberAddress, State: global.Join, Incarnation: 0, Time: time.Now()}
                GossipNodesMutex.Unlock()
            }
        }
    }
}