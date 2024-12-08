package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func main() {
    // Return the column 3 and 4 of the input (OBJECTID, Sign_Type)
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    input := scanner.Text()

	// the delimiter might be changed
    words := strings.Split(input, "\x00")
    if len(words) >= 4 {
        fmt.Printf("%s, %s", strings.TrimSpace(words[2]), strings.TrimSpace(words[3]))
    } else {
        fmt.Printf("Input does not contain enough words.")
    }
}