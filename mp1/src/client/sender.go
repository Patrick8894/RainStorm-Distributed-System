package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
)

func main() {
    // Base domain name pattern
    baseDomain := "fa24-cs425-66%02d.cs.illinois.edu"
    start := 1
    end, _ := strconv.Atoi(os.Args[1])

    // Wrap command-line arguments with '\n' as delimiter and '\0' at the end
    args := strings.Join(os.Args[2:], "\n")

    // Generate the list of domain names
    var ipAddresses []string
    for i := start; i <= end; i++ {
        ipAddresses = append(ipAddresses, fmt.Sprintf(baseDomain, i))
    }

    // Use a WaitGroup to wait for all goroutines to finish
    var wg sync.WaitGroup

    // Array to store responses
    responses := make([]string, end-start+1)

    // Iterate over the list of IP addresses
    for idx, ipAddress := range ipAddresses {
        wg.Add(1)
        filenameSuffix := strconv.Itoa(idx + 1)
        fileName := fmt.Sprintf("../../data/vm%s.log", filenameSuffix)
        
        data := args + "\n" + fileName + "\x00"
        go connectAndSend(ipAddress, idx, data, &wg, responses)
    }

    // Wait for all goroutines to finish
    wg.Wait()

    total_lines := 0

    // Print responses in order
    for _, response := range responses {
        for _, c := range response {
            if c == '\n' {
                total_lines++
            }
        }
        fmt.Print(response)
    }
    fmt.Printf("Total lines: %d\n", total_lines)
}

// connectAndSend attempts to establish a TCP connection to the given IP address,
// sends the provided data, and stores the response in the responses array.
func connectAndSend(ip string, idx int, data string, wg *sync.WaitGroup, responses []string) {
    defer wg.Done()
    // Attempt to establish a TCP connection
    conn, err := net.DialTimeout("tcp", ip+":8080", 5*time.Second)
    if err != nil {
        responses[idx] = fmt.Sprintf("Error connecting to %s: %v\n", ip, err)
        return
    }
    defer conn.Close()



    // Send the wrapped command-line arguments through the socket
    _, err = conn.Write([]byte(data))
    if err != nil {
        responses[idx] = fmt.Sprintf("Error sending data to %s: %v\n", ip, err)
        return
    }

    // Read the response from the server until '\0'
    reader := bufio.NewReader(conn)
    var responseBuilder strings.Builder
    for {
        b, err := reader.ReadByte()
        if err != nil {
            responses[idx] = fmt.Sprintf("Error reading response from %s: %v\n", ip, err)
            return
        }
        if b == '\x00' {
            break
        }
        responseBuilder.WriteByte(b)
    }

    // Store the response in the responses array
    responses[idx] = responseBuilder.String()
}
