package main


import (
	"fmt"
	"os"
	"net"
)

func main (){

	// argeument to send to client node
	// input: 
	// word (string)
	// output: 
	// file_name (string), line(int)  

	// detect whether input number of client machine
	// if len(os.Args) < 2 {
	// 	fmt.Println("Wrong argument input length")
	// 	return
	// }

	conn, err := net.Dial("TCP", "localhost:8080")

	if err != nil {
		return
	}
	defer conn.Close()

}