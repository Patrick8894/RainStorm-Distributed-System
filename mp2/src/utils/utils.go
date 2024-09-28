package utils
import ( 
	pb "mp2/proto"
	"mp2/src/global"
	"fmt"
	"net"
	"google.golang.org/protobuf/proto"
)

func get_gossiplist (GosssipNodes make(map[string]GossipNode)) []*pb.MembershipInfo {
	gossipNodelist := [] *pb.MembershipInfo{}

	for _, GossipNode := range GossipNodes {
        gossipNodelist = append(gossipNodelist, &pb.MembershipInfo{
            MemberID:          GossipNode.ID,
            MemberAddress:     GossipNode.Address,
            MemberState:       GossipNode.State,
            MemberIncarnation: GossipNode.Incarnation,
        })
    }
	
	return gossipNodelist	
}

func get_nodelist (Nodes make(map[string]NodeInfo)) []*pb.MembershipInfo {
	nodelist := [] *pb.MembershipInfo{
		for _, Node := range Nodes {
			{
				memberID: Node.ID,
				memberAddress: Node.Address,
				memberState: Node.State,
			},
		}
	}	
	return nodelist
}

func send_message (conn net.Conn, addr net.Addr, message *pb.SWIMMessage) {
	data, err := proto.Marshal(message)
	if err != nil {
		fmt.Println("Failed to marshal message: %v", err)
	}
	_, err = conn.WriteTo(data, addr)
	if err != nil {
		fmt.Println("Failed to send message: %v", err)
	}
}

func read_message (conn net.Conn) (*pb.SWIMMessage, error) {
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return nil
	}
	var message pb.SWIMMessage
	err = proto.Unmarshal(buf[:n], &message)
	if err != nil {
		fmt.Println("Failed to unmarshal message:", err)
		return nil
	}
	return message, nil
}