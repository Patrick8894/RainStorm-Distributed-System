package main

// import (
// 	"bufio"
// 	"fmt"
// 	"net"
// 	//"os"
// )

// func main() {

// 	// argeument to send to client node
// 	// input:
// 	// word (string)
// 	// output:
// 	// file_name (string), line(int)

// 	// detect whether input number of client machine

// 	conn, err := net.Dial("tcp", "localhost:8080")

// 	if err != nil {
// 		return
// 	}

// 	defer conn.Close()

// 	// example word that send to client
// 	pattern := "log"

// 	// send the word to the client
// 	fmt.Fprintf(conn, pattern)
// 	// get the response from the client
// 	messages, err := bufio.NewReader(conn).ReadString('\n')
// 	fmt.Println(messages)
// }
