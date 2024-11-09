package global

import (
    "encoding/json"
    "fmt"
    "os"
    "hash/crc32"
    "net"
    "sort"
    "strings"
)

type State int

const (
    Suspected State = iota
    Alive
    Down
    Join
)

type NodeInfo struct {
	ID 	string
    Address string
    State   State
}

const (
    HDFSPort           = "8085"
    SWIMPort           = "8082"
    RingMod            = 1024
    ReplicationFactor  = 3
)

// Cluster is a map of ip addresses to node information
var Cluster map[string]NodeInfo

func HashFunc(s string) int {
    /*
    Hash the filename to an integer value from 0 to RingMod - 1.
    */
    h := crc32.ChecksumIEEE([]byte(s))
    return int(h % uint32(RingMod))
}

func GetMembership() map[string]NodeInfo {
    /*
    Get the membership list from local SWIM protocol.
    */
    hostname, err := os.Hostname()
    if err != nil {
        fmt.Println("Error getting hostname:", err)
        return nil
    }

    conn, err := net.Dial("udp", hostname + ":" + SWIMPort)
    if err != nil {
        fmt.Println("Error dialing introducer:", err)
        return nil
    }
    defer conn.Close()

    data := []byte("ls")
    _, err = conn.Write(data)
    if err != nil {
        fmt.Printf("Failed to send message: %v\n", err)
    }

    buffer := make([]byte, 4096)
    if err != nil {
        fmt.Println("No response from select_node:", err)
        return nil
    }

    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Println("Failed to read message:", err)
        return nil
    }

    var response map[string]NodeInfo
    err = json.Unmarshal(buffer[:n], &response)
    if err != nil {
        fmt.Println("Failed to unmarshal message:", err)
        return nil
    }

    return response
}

func FindFileReplicas(filename string) []string {
    /*
    Given a filename, return the ip ddresses of the three replicas.
    */
    fileHash := HashFunc(filename)
    addressHashes := make([]int, 0, len(Cluster))
    addressMap := make(map[int]string)

    // Compute the hash of all addresses in the cluster
    for _, node := range Cluster {
        addressHash := HashFunc(node.Address)
        addressHashes = append(addressHashes, addressHash)
        addressMap[addressHash] = UpdateAddressPort(node.Address, HDFSPort)
    }

    // Sort the address hashes
    sort.Ints(addressHashes)

    // Find at most three replicas with hash values larger or equal to the file hash
    replicas := make([]string, 0, ReplicationFactor)
    for _, hash := range addressHashes {
        if hash >= fileHash {
            replicas = append(replicas, addressMap[hash])
            if len(replicas) == ReplicationFactor {
                return replicas
            }
        }
    }

    // If not enough replicas found, wrap around the ring
    for _, hash := range addressHashes {
        replicas = append(replicas, addressMap[hash])
        if len(replicas) == ReplicationFactor {
            break
        }
    }
    return replicas
}

func UpdateAddressPort(address, newPort string) string {
    parts := strings.Split(address, ":")
    if len(parts) != 2 {
        // Handle error if the address does not match the expected format
        fmt.Println("Invalid address format")
        return address
    }
    host := parts[0]
    return fmt.Sprintf("%s:%s", host, newPort)
}
