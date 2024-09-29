package utils

import (
    pb "mp2/proto"
    "mp2/src/global"
    "fmt"
    "net"
    "google.golang.org/protobuf/proto"
	"time"
)

var GOSSIP_TIMEOUT = 60

// Map global.State to pb.MembershipInfo_State
func MapState(state global.State) pb.MembershipInfo_State {
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
func GetGossiplist(GossipNodes map[string]global.GossipNode) []*pb.MembershipInfo {
    fmt.Println("GossipNodes: ", GossipNodes)
    gossipNodelist := []*pb.MembershipInfo{}
    for _, GossipNode := range GossipNodes {

		// check if the gossipnode is timeout or not
		if GossipNode.Time.Before(time.Now().Add( time.Duration(GOSSIP_TIMEOUT) * time.Second)) {
			continue
		}

        gossipNodelist = append(gossipNodelist, &pb.MembershipInfo{
            MemberID:          GossipNode.ID,
            MemberAddress:     GossipNode.Address,
            MemberStatus:      MapState(GossipNode.State), // Map the state
            MemberIncarnation: int32(GossipNode.Incarnation),
        })
    }
    return gossipNodelist
}

// Corrected function to get the node list
func GetNodelist(Nodes map[string]global.NodeInfo) []*pb.MembershipInfo {
    nodelist := []*pb.MembershipInfo{}
    for _, Node := range Nodes {
        nodelist = append(nodelist, &pb.MembershipInfo{
            MemberID:      Node.ID,
            MemberAddress: Node.Address,
            MemberStatus:  MapState(Node.State), // Map the state
        })
    }
    return nodelist
}

// Corrected function to send a message for receiver
func SendMessage(conn *net.UDPConn, addr *net.UDPAddr, message *pb.SWIMMessage) {
    data, err := proto.Marshal(message)
    if err != nil {
        fmt.Printf("Failed to marshal message: %v\n", err)
        return
    }
    _, err = conn.WriteTo(data, addr) // Use Write method instead of WriteTo
    if err != nil {
        fmt.Printf("Failed to send message: %v\n", err)
    }
}

// Function to read a message
// func ReadMessage(conn net.PacketConn) (*pb.SWIMMessage, error) {
//     buf := make([]byte, 4096)
//     n, err := conn.ReadFrom(buf)
//     if err != nil {
//         fmt.Println("Error reading from connection:", err)
//         return nil, err
//     }
//     var message pb.SWIMMessage
//     err = proto.Unmarshal(buf[:n], &message)
//     if err != nil {
//         fmt.Println("Failed to unmarshal message:", err)
//         return nil, err
//     }
//     return &message, addr, nil
// }