package main


import (
	"fmt"
	// "os"
	"net"
	"os/exec"
	"strings"
)

func main (){
	

	conn, err = net.Listen("tcp")


	// file, err := os.Open(data+"/"+file_name)
	
	// if err != nil {
	// 	fmt.Println("Did not open the file successfully");
	// 	return 
	// }

	// defer file.Close()

	fmt.Println("finished client running");

}


func grep_function (){

	pattern := "log";
	file_path := "../../data/vm1.log";


	cmd := exec.Command("grep", "-cH", pattern, file_path)

	var out strings.Builder

	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil{
		fmt.Printf("Grep Command execution error %q", err)
	}

	fmt.Printf("%q\n", out.String())
}