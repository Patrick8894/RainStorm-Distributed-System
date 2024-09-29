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
// listen the command from other machine
var COMMAND_PORT = "8082"
var PROTOCOL_PERIOD = 2.0
var TIMEOUT_PERIOD = 1.0
var SUSPECT_TIMEOUT = 5.0

var Introducer = false
var GossipNodesMutex sync.Mutex
var Id = ""
var SelfAddress = ""
var Hostname = ""

func main(){
	introducerFlag := flag.Bool("introducer", false, "Set this flag to true if this node is the introducer")

    flag.Parse()

	Introducer = *introducerFlag

	hostname, err := os.Hostname()
    if err != nil {
        fmt.Println("Error getting hostname:", err)
        return
    }
    Hostname = hostname

    SelfAddress = Hostname + ":" + PORT
	fmt.Println("Node ID:", Id)

	if Introducer {
        Id = Hostname + ":" + PORT + ":" + strconv.FormatInt(time.Now().UnixNano(), 10)
		global.Nodes[Id] = global.NodeInfo{ID: Id, Address: Hostname + ":" + PORT, State: global.Alive}
	} else {
		dialIntroducer()
	}
	// Need additional logic to handle the case where the Introducer is down

	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup
    wg.Add(4)

	go func() {
        defer wg.Done()
        startServer()
    }()
	
	go func() {
        defer wg.Done()
        startClient()
    }()

    go func () {
        defer wg.Done()
        startHandlecommand()
    }()

    go func () {
        defer wg.Done()
        checkSuspected()
    }()

	wg.Wait()

}

func startHandlecommand() {
    hostname, err := os.Hostname()
    var COMMAND_ADDRESS = hostname + ":" + global.COMMAND_PORT
    addr, err := net.ResolveUDPAddr("udp", COMMAND_ADDRESS)
    if err != nil {
        fmt.Println("Error resolving in Command server address:", err)
        return
    }

    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        fmt.Println("Error starting Command server:", err)
        return
    }
    defer conn.Close()

    fmt.Println("Command server started on", addr)

    buffer := make([]byte, 4096)
    for {
        n, addr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Error reading from connection:", err)
            continue
        }

        // read the command from buffer
        command := string(buffer[:n])
        fmt.Println("Received command:", command)

        if command == "ls" {
            // send the list of nodes to the sender
            GossipNodesMutex.Lock()
            jsonData, err := json.Marshal(global.Nodes)
            if err != nil {
                fmt.Println("Error serializing data:", err)
                os.Exit(1)
            }
            GossipNodesMutex.Unlock()
            fmt.Println("Received ls command, sending list of nodes...")
            _, err = conn.WriteToUDP(jsonData, addr)
            
            if err != nil {
                fmt.Println("Failed to response to command message:", err)
                return
            }
        } else if command == "lsg" {
            // send the list of nodes to the sender
            GossipNodesMutex.Lock()

            jsonData, err := json.Marshal(global.GossipNodes)
            if err != nil {
                fmt.Println("Error serializing data:", err)
                os.Exit(1)
            }
            GossipNodesMutex.Unlock()
            fmt.Println("Received ls command, sending list of gossipnodes...")
            _, err = conn.WriteToUDP(jsonData, addr)
            
            if err != nil {
                fmt.Println("Failed to response to command message:", err)
                return
            }
        } else if command == "on" {
            // turn on the suspect protocol
            fmt.Println("Received on command, turning on the suspect protocol...")
            global.Protocol = global.SWIM_SUSPIECT_PROROCOL
        } else if command == "off" {
            // turn off the suspect protocol, maintain the SWIM PINGACK protocl
            fmt.Println("Received off command, shutting down the suspect protocol...")
            global.Protocol = global.SWIM_PROROCOL
        } else if command == "kill"{
            // kill the node.go
            fmt.Println("Received kill command, shutting down...")
            os.Exit(0)
        } else {
            // setting DropRate
            global.DropRate, err = strconv.ParseFloat(command, 64)
            fmt.Println("Received drop command, setting DropRate to", global.DropRate)
        }
    }
}

func dialIntroducer() {
    // If the node is not the introducer, then send a JOIN message to the introducer
    // TODO: implement the logic to dial the Introducer and get the list of nodes
    fmt.Println("Dialing introducer...")
    Id = Hostname + ":" + PORT + ":" + strconv.FormatInt(time.Now().UnixNano(), 10)

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
    _, err = conn.Write(data)
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

    fmt.Println("Received message from Introducer:", response)

    GossipNodesMutex.Lock()
    global.Nodes = response
    GossipNodesMutex.Unlock()
    fmt.Println("Joined the cluster")
}

func startServer() {
    addr, err := net.ResolveUDPAddr("udp", ":" + PORT)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}
    
    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        fmt.Println("Error starting server:", err)
        return
    }
    defer conn.Close()

    fmt.Println("Server started on", addr)

    buffer := make([]byte, 4096)
    for {
        n, addr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Error reading from connection:", err)
            continue
        }

        if rand.Float64() < global.DropRate {
            fmt.Println("Dropping packet from", addr)
            continue
        }

        var message pb.SWIMMessage
        err = proto.Unmarshal(buffer[:n], &message)
        if err != nil {
            fmt.Println("Failed to unmarshal message:", err)
            continue
        }
        
        // merge the gossip list
        handleGossip(message)

        //  received message
        fmt.Printf("Server Received message: %+v\n", message)

        if message.Type == pb.SWIMMessage_DIRECT_PING {
            // Response to ping
            // Create a SWIMMessage to send
            if message.TargetId != Id {
                continue
            }

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
            fmt.Println("Sending PONG message to", addr)
            _, err = conn.WriteToUDP(data, addr) // Use Write method instead of WriteTo
            if err != nil {
                fmt.Printf("Failed to send message in response PONG direct ping: %v\n", err)
            }

        	fmt.Println("Message success direct ping to server")

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
                Type: pb.SWIMMessage_DIRECT_PING,
                Sender: message.Sender,
                Target: message.Target,
                TargetId: message.TargetId,
                Membership: gossiplist,
            }
            
            data, err := proto.Marshal(pingMessage)
            if err != nil {
                fmt.Printf("Failed to marshal message: %v\n", err)
                return
            }
            _, err = conn.WriteToUDP(data, targetAddr) // Use Write method instead of WriteTo
            if err != nil {
                fmt.Printf("Failed to send message in indirect ping: %v\n", err)
            }

            fmt.Println("Indirect PING message success sent to target node")

        } else if message.Type == pb.SWIMMessage_JOIN {
            // Introducer response message
			if !Introducer {
                continue
            }
			GossipNodesMutex.Lock()
            // Add the new node to the list of nodes
            global.Nodes[message.Membership[0].MemberID] = global.NodeInfo{ID: message.Membership[0].MemberID , Address: message.Sender, State: global.Alive}
            
            jsonData, err := json.Marshal(global.Nodes)
            if err != nil {
                fmt.Println("Error serializing data:", err)
                os.Exit(1)
            }
            GossipNodesMutex.Unlock()
            _, err = conn.WriteToUDP(jsonData, addr)
            if err != nil {
                fmt.Println("Failed to response to JOIN message:", err)
                return
            }
            // fmt.Println("Introducer response message success sent to new node")
            // fmt.Println("gossipNodes before join: ", global.GossipNodes)
            GossipNodesMutex.Lock()
            global.GossipNodes[message.Membership[0].MemberID] = global.GossipNode{ID: message.Membership[0].MemberID, Address: message.Sender, State: global.Join, Incarnation: 0, Time: time.Now()}
            GossipNodesMutex.Unlock()
            // fmt.Println("gossipNodes after join: ", global.GossipNodes)

            // fmt.Println("Introducer response message success sent to new node")

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
            _, err = conn.WriteToUDP(data, targetAddr) // Use Write method instead of WriteTo
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
	GossipNodesMutex.Lock()
    for _, node := range global.Nodes {
        nodesArray = append(nodesArray, node)
    }
	GossipNodesMutex.Unlock()

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
                GossipNodesMutex.Lock()
                nodesArray = []global.NodeInfo{} // Clear the array
                for _, node := range global.Nodes {
                    nodesArray = append(nodesArray, node)
                }
                GossipNodesMutex.Unlock()

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
    GossipNodesMutex.Lock()

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

    GossipNodesMutex.Unlock()
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

            
            addr, err := net.ResolveUDPAddr("udp", node.Address)
            if err != nil {
                fmt.Println("Error resolving server address:", err)
                return
            }

            conn, err := net.DialUDP("udp", nil, addr)
            if err != nil {
                fmt.Println("Failed to dial random node:", err)
                resultChan <- false
                return
            }
            defer conn.Close()

            indirectPingMessage := &pb.SWIMMessage{
                Type:   pb.SWIMMessage_INDIRECT_PING,
                Sender: conn.LocalAddr().String(),
                Target: node.Address, // Assuming NodeInfo has an Address field
                TargetId: node.ID,
                Membership: gossiplist,
            }
            // Serialize the INDIRECT_PING message using protobuf
            data, err := proto.Marshal(indirectPingMessage)
            if err != nil {
                fmt.Println("Failed to marshal INDIRECT_PING message:", err)
                resultChan <- false
                return
            }

            // Set a write deadline for the connection
            // conn.SetWriteDeadline(time.Now().Add(time.Duration(TIMEOUT_PERIOD) * time.Second))
            
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
            n, _, err := conn.ReadFromUDP(buffer)
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
    udpAddr, err := net.ResolveUDPAddr("udp", node.Address)
    if err != nil {
        fmt.Printf("Failed to resolve address %s: %v\n", node.Address, err)
        return
    }

    conn, err := net.DialUDP("udp", nil, udpAddr)
    if err != nil {
        fmt.Printf("Failed to ping %s: %v\n", node.Address, err)
        // TODO: Handle the case where the direct node is down

        rst := pingIndirect(node)

        if rst == false {
            GossipNodesMutex.Lock()
            if global.Protocol == global.SWIM_PROROCOL {
                fmt.Println("Delete SWIM_PROROCOL; nodeId: ", node.ID)
                delete(global.Nodes, node.ID)

                // add the node to the GossipNodes list
                global.GossipNodes[node.ID] = global.GossipNode{ID: node.ID, Address: node.Address, State: global.Down, Incarnation: 0, Time: time.Now()}
                
                
            } else if global.Protocol == global.SWIM_SUSPIECT_PROROCOL {
                fmt.Println("Delete SWIM_SUSPIECT_PROROCOL; nodeId: ", node.ID)
                global.SuspectedNodes[node.ID] = time.Now()
                global.GossipNodes[node.ID] = global.GossipNode{ID: node.ID, Address: node.Address, State: global.Suspected, Incarnation: int(global.GossipNodes[node.ID].Incarnation), Time: time.Now()}
            }
            GossipNodesMutex.Unlock()
        }
        return
    }
    defer conn.Close()

    // TODO: Send a PING message to the server
    GossipNodesMutex.Lock()
    gossiping := utils.GetGossiplist(global.GossipNodes)
    GossipNodesMutex.Unlock()

    message := &pb.SWIMMessage{
        Type:   pb.SWIMMessage_DIRECT_PING,
        Sender: conn.LocalAddr().String(),
        Target: node.Address,
        TargetId: node.ID,
        Membership: gossiping,
    }

    data, err := proto.Marshal(message)
    if err != nil {
        fmt.Printf("Failed to marshal message: %v\n", err)
        return
    }

    // Set a read deadline for the response
    // conn.SetWriteDeadline(time.Now().Add(time.Duration(TIMEOUT_PERIOD) * time.Second))

    // Send the message using WriteToUDP
    _, err = conn.Write(data)
    if err != nil {
        fmt.Printf("Failed to send message in ping server: %v\n", err)
    }

    // Set a read deadline for the response
    conn.SetReadDeadline(time.Now().Add(time.Duration(TIMEOUT_PERIOD) * time.Second))

    buffer := make([]byte, 4096)
    n, _, err := conn.ReadFromUDP(buffer)
    if err != nil {
        fmt.Printf("No response from %s: %v\n", node.Address, err)
        // TODO: Handle the case where the node is down

        rst := pingIndirect(node)
        
        if rst == false {
            // delete the node from the Nodes list
            GossipNodesMutex.Lock()
            if global.Protocol == global.SWIM_PROROCOL {
                fmt.Println("Delete SWIM_PROROCOL; nodeId: ", node.ID)
                delete(global.Nodes, node.ID)
                // add the node to the GossipNodes list
                global.GossipNodes[node.ID] = global.GossipNode{ID: node.ID, Address: node.Address, State: global.Down, Incarnation: 0, Time: time.Now()}
                
            } else if global.Protocol == global.SWIM_SUSPIECT_PROROCOL {
                fmt.Println("Delete SWIM_SUSPIECT_PROROCOL; nodeId: ", node.ID)
                global.SuspectedNodes[node.ID] = time.Now()
                global.GossipNodes[node.ID] = global.GossipNode{ID: node.ID, Address: node.Address, State: global.Suspected, Incarnation: int(global.GossipNodes[node.ID].Incarnation), Time: time.Now()}
            }
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
    GossipNodesMutex.Lock()
    fmt.Println("Handle Gossip message: ", message.Membership)
    for _, Membership := range message.Membership {
        if Membership.MemberStatus == utils.MapState(global.Down) {
            if Membership.MemberID == Id {
                dialIntroducer()
                return
            }
            // delete the node from the Nodes list
            
            _, exists := global.Nodes[Membership.MemberID];
            if exists {
                delete(global.Nodes, Membership.MemberID)
                // add the node to the GossipNodes list
                global.GossipNodes[Membership.MemberID] = global.GossipNode{ID: Membership.MemberID, Address: Membership.MemberAddress, State: global.Down, Incarnation: 0, Time: time.Now()}
            }
        } else if Membership.MemberStatus == utils.MapState(global.Join) {
            // add the node to the Nodes list

            _, exists := global.Nodes[Membership.MemberID];
            fmt.Println("Join: ", Membership.MemberID)
            fmt.Println("Exists: ", exists)
            if !exists {
                global.Nodes[Membership.MemberID] = global.NodeInfo{ID: Membership.MemberID, Address: Membership.MemberAddress, State: global.Alive}
                global.GossipNodes[Membership.MemberID] = global.GossipNode{ID: Membership.MemberID, Address: Membership.MemberAddress, State: global.Join, Incarnation: 0, Time: time.Now()}
                
            }
        } else if Membership.MemberStatus == utils.MapState(global.Suspected) {
            // add the node to the GossipNodes list
            if Membership.MemberID == Id {         
                global.Incarnation = int(Membership.MemberIncarnation) + 1
                global.GossipNodes[Membership.MemberID] = global.GossipNode{ID: Membership.MemberID, Address: Membership.MemberAddress, State: global.Alive, Incarnation: global.Incarnation, Time: time.Now()}
            } else {
                gossipNode, exists := global.GossipNodes[Membership.MemberID]
                if !exists || (gossipNode.Incarnation <= int(Membership.MemberIncarnation) && gossipNode.State != global.Down) {
                    global.GossipNodes[Membership.MemberID] = global.GossipNode{ID: Membership.MemberID, Address: Membership.MemberAddress, State: global.Suspected, Incarnation: int(Membership.MemberIncarnation), Time: time.Now()}
                }
            }
        } else if Membership.MemberStatus == utils.MapState(global.Alive) {
            // add the node to the GossipNodes list
            gossipNode, exists := global.GossipNodes[Membership.MemberID]
            if !exists || (gossipNode.Incarnation < int(Membership.MemberIncarnation) && gossipNode.State != global.Down) {
                delete(global.SuspectedNodes, Membership.MemberID)
                global.GossipNodes[Membership.MemberID] = global.GossipNode{ID: Membership.MemberID, Address: Membership.MemberAddress, State: global.Alive, Incarnation: int(Membership.MemberIncarnation), Time: time.Now()}
            }
            
        }
    }
    GossipNodesMutex.Unlock()
}

func checkSuspected() {
    for {
        time.Sleep(time.Duration(SUSPECT_TIMEOUT) * time.Second)
        GossipNodesMutex.Lock()
        for id, suspectTime := range global.SuspectedNodes {
            if suspectTime.Before(time.Now().Add(-time.Duration(SUSPECT_TIMEOUT) * time.Second)) {
                // If the suspected node stay too long, it need to be deleted
                delete(global.SuspectedNodes, id)
                delete(global.Nodes, id)
                global.GossipNodes[id] = global.GossipNode{ID: id, Address: global.Nodes[id].Address, State: global.Down, Incarnation: 0, Time: time.Now()}
            }
        }
        GossipNodesMutex.Unlock()
    }
}