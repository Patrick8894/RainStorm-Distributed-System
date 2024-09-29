// send the command to each node
package main
import (
	"flag"
	"fmt"
	"net"
	"encoding/json"
	"strconv"
	"mp2/src/global"
)

func main (){
	// receive the command from the command line
	select_node := flag.String("s", "", "Select the node to send the command")
	command := flag.String("c", "", "Command to send to each node")
	flag.Parse()

	// Check if the command is provided
	if *command == "" {
		fmt.Println("Please provide a command using --command flag")
		return
	}
	if *select_node == "" {
		fmt.Println("Please provide a node using --node flag")
		return
	}

	// send the command to each node
	if *command == "ls" {
		// send the list command to select_node IP address with UDP
		nodeIndex, err :=  strconv.Atoi(*select_node)
		conn, err := net.Dial("udp", global.Cluster[nodeIndex-1])
		if err != nil {
			fmt.Println("Error dialing introducer:", err)
			return
		}
		defer conn.Close()
		fmt.Println("Send comomand to %v ls", global.Cluster[nodeIndex-1])
		data := []byte(*command)
		_, err = conn.Write(data)
		if err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
		}
		// wait for the response from select_node
		// print the response from select_node
		// Use Json to unmarshal when receiving the whole NodeList
		buffer := make([]byte, 4096)
		n, err := conn.Read(buffer)
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
	
		fmt.Println("Received message from select node ls:")
		for _, node := range response {
			for key, value := range node {
				fmt.Println(key, value)
			}
		}

	} else if *command == "lsg" {
		// send the list command to select_node IP address with UDP
		nodeIndex, err :=  strconv.Atoi(*select_node)
		conn, err := net.Dial("udp", global.Cluster[nodeIndex-1])
		if err != nil {
			fmt.Println("Error dialing introducer:", err)
			return
		}
		defer conn.Close()
		fmt.Println("Send comomand to %v lsg", global.Cluster[nodeIndex-1])
		data := []byte(*command)
		_, err = conn.Write(data)
		if err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
		}
		// wait for the response from select_node
		// print the response from select_node
		// Use Json to unmarshal when receiving the whole NodeList
		buffer := make([]byte, 4096)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("No response from select_node:", err)
			return
		}
	
		var response map[string]global.GossipNode
		err = json.Unmarshal(buffer[:n], &response)
		if err != nil {
			fmt.Println("Failed to unmarshal message:", err)
			return
		}
	
		fmt.Println("Received message from select node lsg:")
		for _, node := range response {
			for key, value := range node {
				fmt.Println(key, value)
			}
		}
	
	} else if *command == "on" {
		nodeIndex, err := strconv.Atoi(*select_node)
		conn, err := net.Dial("udp", global.Cluster[nodeIndex-1])
		fmt.Println("Send comomand to %v ON", global.Cluster[nodeIndex-1])
		defer conn.Close()

		data := []byte(*command)
		_, err = conn.Write(data)
		if err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
		}
	} else if *command == "off" {
		nodeIndex, err := strconv.Atoi(*select_node)
		conn, err := net.Dial("udp", global.Cluster[nodeIndex-1])
		fmt.Println("Send comomand to %v OFF", global.Cluster[nodeIndex-1])
		defer conn.Close()

		data := []byte(*command)
		_, err = conn.Write(data)
		if err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
		}
	} else if *command == "kill" {
		nodeIndex, err := strconv.Atoi(*select_node)
		conn, err := net.Dial("udp", global.Cluster[nodeIndex-1])
		fmt.Println("Send comomand to %v kill", global.Cluster[nodeIndex-1])
		defer conn.Close()

		data := []byte(*command)
		_, err = conn.Write(data)
		if err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
		}
	}

}
