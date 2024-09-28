package utils

import (
    pb "mp2/proto"
    "mp2/src/global"
    "fmt"
    "net"
    "google.golang.org/protobuf/proto"
)

// Map global.State to pb.MembershipInfo_State
func mapState(state global.State) pb.MembershipInfo_State {
    switch state {
    case global.Suspected:
        return pb.MembershipInfo_Suspected
    case global.Alive:
        return pb.MembershipInfo_Alive
    case global.Down:
        return pb.MembershipInfo_Down
	case global.Join:
        return pb.MembershipInfo_Join // Default case
	default:
		return pb.MembershipInfo_Suspected
    }
}

// Corrected function to get the gossip list
func get_gossiplist(GossipNodes map[string]global.GossipNode) []*pb.MembershipInfo {
    gossipNodelist := []*pb.MembershipInfo{}
    for _, GossipNode := range GossipNodes {
        gossipNodelist = append(gossipNodelist, &pb.MembershipInfo{
            MemberID:          GossipNode.ID,
            MemberAddress:     GossipNode.Address,
            MemberStatus:      mapState(GossipNode.State), // Map the state
            MemberIncarnation: int32(GossipNode.Incarnation),
        })
    }
    return gossipNodelist
}

// Corrected function to get the node list
func get_nodelist(Nodes map[string]global.NodeInfo) []*pb.MembershipInfo {
    nodelist := []*pb.MembershipInfo{}
    for _, Node := range Nodes {
        nodelist = append(nodelist, &pb.MembershipInfo{
            MemberID:      Node.ID,
            MemberAddress: Node.Address,
            MemberStatus:  mapState(Node.State), // Map the state
        })
    }
    return nodelist
}

// Corrected function to send a message
func send_message(conn net.Conn, addr net.Addr, message *pb.SWIMMessage) () {
    data, err := proto.Marshal(message)
    if err != nil {
        fmt.Printf("Failed to marshal message: %v\n", err)
        return
    }
    _, err = conn.Write(data) // Use Write method instead of WriteTo
    if err != nil {
        fmt.Printf("Failed to send message: %v\n", err)
    }
}

// Function to read a message
func read_message(conn net.Conn) (*pb.SWIMMessage, error) {
    buf := make([]byte, 4096)
    n, err := conn.Read(buf)
    if err != nil {
        fmt.Println("Error reading from connection:", err)
        return nil, err
    }
    var message pb.SWIMMessage
    err = proto.Unmarshal(buf[:n], &message)
    if err != nil {
        fmt.Println("Failed to unmarshal message:", err)
        return nil, err
    }
    return &message, nil
}