package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strings"
    "sync"
    "time"
)

func main() {
    // Base domain name pattern
    baseDomain := "fa24-cs425-66%s.cs.illinois.edu"
    start := 1
    end := 10

    // Wrap command-line arguments with '\n' as delimiter and '\0' at the end
    args := strings.Join(os.Args[1:], "\n") + "\x00"

    // Generate the list of domain names
    var ipAddresses []string
    for i := start; i <= end; i++ {
        var formattedNumber string
        if i < 10 {
            formattedNumber = fmt.Sprintf("%d", i)
        } else {
            formattedNumber = fmt.Sprintf("%02d", i)
        }
        ipAddresses = append(ipAddresses, fmt.Sprintf(baseDomain, formattedNumber))
    }

    // ipAddresses = []string{"localhost"}

    // Use a WaitGroup to wait for all goroutines to finish
    var wg sync.WaitGroup

    // Mutex for synchronized printing
    var mu sync.Mutex

    // Iterate over the list of IP addresses
    for _, ipAddress := range ipAddresses {
        wg.Add(1)
        go connectAndSend(ipAddress, args, &wg, &mu)
    }

    // Wait for all goroutines to finish
    wg.Wait()
}

// connectAndSend attempts to establish a TCP connection to the given IP address,
// sends the provided data, and prints the connection status.
func connectAndSend(ip, data string, wg *sync.WaitGroup, mu *sync.Mutex) {
    defer wg.Done()
    // Attempt to establish a TCP connection
    conn, err := net.DialTimeout("tcp", ip+":8080", 5*time.Second)
    if err != nil {
        mu.Lock()
        fmt.Printf("Error connecting to %s: %v\n", ip, err)
        mu.Unlock()
        return
    }
    defer conn.Close()

    // Send the wrapped command-line arguments through the socket
    _, err = conn.Write([]byte(data))
    if err != nil {
        mu.Lock()
        fmt.Printf("Error sending data to %s: %v\n", ip, err)
        mu.Unlock()
        return
    }

    // Read the response from the server until '\0'
    reader := bufio.NewReader(conn)
    var response strings.Builder
    for {
        b, err := reader.ReadByte()
        if err != nil {
            mu.Lock()
            fmt.Printf("Error reading from %s: %v\n", ip, err)
            mu.Unlock()
            return
        }
        if b == '\x00' {
            break
        }
        response.WriteByte(b)
    }

    // Lock the mutex, print the response, and then unlock the mutex
    mu.Lock()
    fmt.Printf(response.String())
    fmt.Printf("Successfully connected to %s and sent data\n", ip)
    mu.Unlock()
}
