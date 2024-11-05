package global

type NodeInfo struct {
	ID 	string
    Address string
    State   State
}

HDFSPort := "8085"
SWIMPort := "8082"

func HashFunc(s string) int {
    h := crc32.ChecksumIEEE([]byte(s))
    return int(h % uint32(ringMod))
}

func GetMembership() map[string]global.NodeInfo {
    conn, err := net.Dial("udp", "localhost:" + global.SWIMPort)
    if err != nil {
        fmt.Println("Error dialing introducer:", err)
        return
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
        return
    }

    var response map[string]global.NodeInfo
    err = json.Unmarshal(buffer[:n], &response)
    if err != nil {
        fmt.Println("Failed to unmarshal message:", err)
        return
    }

    if response == cluster {
        return
    }
}

func FindFileReplicas(filename string) []string {
    fileHash := global.HashFile(filename)
    addressHashes := make([]int, 0, len(cluster))
    addressMap := make(map[int]string)

    // Compute the hash of all addresses in the cluster
    for _, node := range cluster {
        addressHash := global.HashFile(node.Address)
        addressHashes = append(addressHashes, addressHash)
        addressMap[addressHash] = node.Address
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
}