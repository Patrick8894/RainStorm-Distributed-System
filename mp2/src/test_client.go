package main

import (
    "log"
    "net"
    pb "mp2/proto" // Update this to your correct package path
    "google.golang.org/protobuf/proto"
)

func main() {
    // Connect to the server
    conn, err := net.Dial("tcp", "localhost:9000")
    if err != nil {
        log.Fatalf("Failed to connect to server: %v", err)
    }
    defer conn.Close()

    // Create a SWIMMessage to send
    message := &pb.SWIMMessage{
        Type:   pb.SWIMMessage_INDIRECT_PING,
        Sender: "node1",
        Target: "node2",
        Membership: []*pb.MembershipInfo{
        },
    }

    // Serialize the message using protobuf
    data, err := proto.Marshal(message)
    if err != nil {
        log.Fatalf("Failed to marshal message: %v", err)
    }

    // Send the message to the server
    _, err = conn.Write(data)
    if err != nil {
        log.Fatalf("Failed to send message: %v", err)
    }

    log.Println("Message sent to server")
}
