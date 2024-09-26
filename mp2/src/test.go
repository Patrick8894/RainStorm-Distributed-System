package main

import (
    "fmt"
    pb "mp2/proto"  // Import your generated protobuf package
)

func main() {
    // Example usage of the generated protobuf code
    message := &pb.SWIMMessage{
        Type:   pb.SWIMMessage_PING,  // Use the enum value from the generated code
        Sender: "node1",
        Target: "node2",
        Membership: []*pb.MembershipInfo{
            {
                Address: "node1",
                Status:  "Alive",
            },
        },
    }

    fmt.Println("Message:", message)
}